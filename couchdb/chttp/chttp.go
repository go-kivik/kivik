// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

// Package chttp provides a minimal HTTP driver backend for communicating with
// CouchDB servers.
package chttp

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"

	"github.com/go-kivik/couchdb/v4/internal"
	"github.com/go-kivik/kivik/v4"
)

const typeJSON = "application/json"

// The default UserAgent values
const (
	UserAgent = "Kivik chttp"
	Version   = "4.0.0-prerelease"
)

// Client represents a client connection. It embeds an *http.Client
type Client struct {
	// UserAgents is appended to set the User-Agent header. Typically it should
	// contain pairs of product name and version.
	UserAgents []string

	*http.Client

	rawDSN   string
	dsn      *url.URL
	basePath string
	auth     Authenticator
	authMU   sync.Mutex

	// noGzip will be set to true if the server fails on gzip-encoded requests.
	noGzip bool
}

// New returns a connection to a remote CouchDB server. If credentials are
// included in the URL, requests will be authenticated using Cookie Auth. To
// use HTTP BasicAuth or some other authentication mechanism, do not specify
// credentials in the URL, and instead call the Auth() method later.
func New(client *http.Client, dsn string, options map[string]interface{}) (*Client, error) {
	dsnURL, err := parseDSN(dsn)
	if err != nil {
		return nil, err
	}
	user := dsnURL.User
	dsnURL.User = nil
	c := &Client{
		Client:   client,
		dsn:      dsnURL,
		basePath: strings.TrimSuffix(dsnURL.Path, "/"),
		rawDSN:   dsn,
	}
	if user != nil {
		password, _ := user.Password()
		err := c.Auth(&CookieAuth{
			Username: user.Username(),
			Password: password,
		})
		if err != nil {
			return nil, err
		}
	}
	if gzip, ok := options[internal.OptionNoCompressedRequests]; ok {
		c.noGzip, ok = gzip.(bool)
		if !ok {
			return nil, &kivik.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("OptionNoCompressedRequests is %T, must be bool", gzip)}
		}
	}
	if err := c.setUserAgent(options); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) setUserAgent(options map[string]interface{}) error {
	c.UserAgents = []string{
		fmt.Sprintf("Kivik/%s", kivik.KivikVersion),
		fmt.Sprintf("Kivik CouchDB driver/%s", Version),
	}
	ua, ok := options[internal.OptionUserAgent]
	if !ok {
		return nil
	}
	userAgent, ok := ua.(string)
	if !ok {
		return &kivik.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("OptionUserAgent is %T, must be string", ua)}
	}
	c.UserAgents = append(c.UserAgents, userAgent)
	return nil
}

func parseDSN(dsn string) (*url.URL, error) {
	if dsn == "" {
		return nil, &curlError{
			httpStatus: http.StatusBadRequest,
			error:      errors.New("no URL specified"),
		}
	}
	if !strings.HasPrefix(dsn, "http://") && !strings.HasPrefix(dsn, "https://") {
		dsn = "http://" + dsn
	}
	dsnURL, err := url.Parse(dsn)
	if err != nil {
		return nil, fullError(http.StatusBadRequest, err)
	}
	if dsnURL.Path == "" {
		dsnURL.Path = "/"
	}
	return dsnURL, nil
}

// DSN returns the unparsed DSN used to connect.
func (c *Client) DSN() string {
	return c.rawDSN
}

// Auth authenticates using the provided Authenticator.
func (c *Client) Auth(a Authenticator) error {
	if c.auth != nil {
		return errors.New("auth already set")
	}
	if err := a.Authenticate(c); err != nil {
		return err
	}
	c.auth = a
	return nil
}

// Response represents a response from a CouchDB server.
type Response struct {
	*http.Response

	// ContentType is the base content type, parsed from the response headers.
	ContentType string
}

// DecodeJSON unmarshals the response body into i. This method consumes and
// closes the response body.
func DecodeJSON(r *http.Response, i interface{}) error {
	defer CloseBody(r.Body)
	if err := json.NewDecoder(r.Body).Decode(i); err != nil {
		return &kivik.Error{Status: http.StatusBadGateway, Err: err}
	}
	return nil
}

// DoJSON combines [Client.DoReq], [Client.ResponseError], and
// [Response.DecodeJSON], and closes the response body.
func (c *Client) DoJSON(ctx context.Context, method, path string, opts *Options, i interface{}) error {
	res, err := c.DoReq(ctx, method, path, opts)
	if err != nil {
		return err
	}
	if res.Body != nil {
		defer CloseBody(res.Body)
	}
	if err = ResponseError(res); err != nil {
		return err
	}
	err = DecodeJSON(res, i)
	return err
}

func (c *Client) path(path string) string {
	if c.basePath != "" {
		return c.basePath + "/" + strings.TrimPrefix(path, "/")
	}
	return path
}

// fullPathMatches returns true if the target resolves to match path.
func (c *Client) fullPathMatches(path, target string) bool {
	p, err := url.Parse(path)
	if err != nil {
		// should be impossible
		return false
	}
	p.RawQuery = ""
	t := new(url.URL)
	*t = *c.dsn // shallow copy
	t.Path = c.path(target)
	t.RawQuery = ""
	return t.String() == p.String()
}

// NewRequest returns a new *http.Request to the CouchDB server, and the
// specified path. The host, schema, etc, of the specified path are ignored.
func (c *Client) NewRequest(ctx context.Context, method, path string, body io.Reader, opts *Options) (*http.Request, error) {
	fullPath := c.path(path)
	reqPath, err := url.Parse(fullPath)
	if err != nil {
		return nil, fullError(http.StatusBadRequest, err)
	}
	u := *c.dsn // Make a copy
	u.Path = reqPath.Path
	u.RawQuery = reqPath.RawQuery
	compress, body := c.compressBody(u.String(), body, opts)
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, &kivik.Error{Status: http.StatusBadRequest, Err: err}
	}
	if compress {
		req.Header.Add("Content-Encoding", "gzip")
	}
	req.Header.Add("User-Agent", c.userAgent())
	return req.WithContext(ctx), nil
}

func (c *Client) shouldCompressBody(path string, body io.Reader, opts *Options) bool {
	if c.noGzip || (opts != nil && opts.NoGzip) {
		return false
	}
	// /_session only supports compression from CouchDB 3.2.
	if c.fullPathMatches(path, "/_session") {
		return false
	}
	if body == nil {
		return false
	}
	return true
}

// compressBody compresses body with gzip compression if appropriate. It will
// return true, and the compressed stream, or false, and the unaltered stream.
func (c *Client) compressBody(path string, body io.Reader, opts *Options) (bool, io.Reader) {
	if !c.shouldCompressBody(path, body, opts) {
		return false, body
	}
	r, w := io.Pipe()
	go func() {
		if closer, ok := body.(io.Closer); ok {
			defer closer.Close()
		}
		gz := gzip.NewWriter(w)
		_, err := io.Copy(gz, body)
		gz.Close()
		w.CloseWithError(err)
	}()
	return true, r
}

// DoReq does an HTTP request. An error is returned only if there was an error
// processing the request. In particular, an error status code, such as 400
// or 500, does _not_ cause an error to be returned.
func (c *Client) DoReq(ctx context.Context, method, path string, opts *Options) (*http.Response, error) {
	if method == "" {
		return nil, errors.New("chttp: method required")
	}
	var body io.Reader
	if opts != nil {
		if opts.GetBody != nil {
			var err error
			opts.Body, err = opts.GetBody()
			if err != nil {
				return nil, err
			}
		}
		if opts.Body != nil {
			body = opts.Body
			defer opts.Body.Close() // nolint: errcheck
		}
	}
	req, err := c.NewRequest(ctx, method, path, body, opts)
	if err != nil {
		return nil, err
	}
	fixPath(req, path)
	setHeaders(req, opts)
	setQuery(req, opts)
	if opts != nil {
		req.GetBody = opts.GetBody
	}

	trace := ContextClientTrace(ctx)
	if trace != nil {
		trace.httpRequest(req)
		trace.httpRequestBody(req)
	}

	response, err := c.Do(req)
	if trace != nil {
		trace.httpResponse(response)
		trace.httpResponseBody(response)
	}
	return response, netError(err)
}

func netError(err error) error {
	if err == nil {
		return nil
	}
	if urlErr, ok := err.(*url.Error); ok {
		// If this error was generated by EncodeBody, it may have an emedded
		// status code (!= 500), which we should honor.
		status := kivik.HTTPStatus(urlErr.Err)
		if status == http.StatusInternalServerError {
			status = http.StatusBadGateway
		}
		return fullError(status, err)
	}
	if status := kivik.HTTPStatus(err); status != http.StatusInternalServerError {
		return err
	}
	return fullError(http.StatusBadGateway, err)
}

// fixPath sets the request's URL.RawPath to work with escaped characters in
// paths.
func fixPath(req *http.Request, path string) {
	// Remove any query parameters
	parts := strings.SplitN(path, "?", 2) // nolint:gomnd
	req.URL.RawPath = "/" + strings.TrimPrefix(parts[0], "/")
}

// BodyEncoder returns a function which returns the encoded body. It is meant
// to be used as a http.Request.GetBody value.
func BodyEncoder(i interface{}) func() (io.ReadCloser, error) {
	return func() (io.ReadCloser, error) {
		return EncodeBody(i), nil
	}
}

// EncodeBody JSON encodes i to an io.ReadCloser. If an encoding error
// occurs, it will be returned on the next read.
func EncodeBody(i interface{}) io.ReadCloser {
	done := make(chan struct{})
	r, w := io.Pipe()
	go func() {
		defer close(done)
		var err error
		switch t := i.(type) {
		case []byte:
			_, err = w.Write(t)
		case json.RawMessage: // Only needed for Go 1.7
			_, err = w.Write(t)
		case string:
			_, err = w.Write([]byte(t))
		default:
			err = json.NewEncoder(w).Encode(i)
			switch err.(type) {
			case *json.MarshalerError, *json.UnsupportedTypeError, *json.UnsupportedValueError:
				err = &kivik.Error{Status: http.StatusBadRequest, Err: err}
			}
		}
		_ = w.CloseWithError(err)
	}()
	return &ebReader{
		ReadCloser: r,
		done:       done,
	}
}

type ebReader struct {
	io.ReadCloser
	done <-chan struct{}
}

var _ io.ReadCloser = &ebReader{}

func (r *ebReader) Close() error {
	err := r.ReadCloser.Close()
	<-r.done
	return err
}

func setHeaders(req *http.Request, opts *Options) {
	accept := typeJSON
	contentType := typeJSON
	if opts != nil {
		if opts.Accept != "" {
			accept = opts.Accept
		}
		if opts.ContentType != "" {
			contentType = opts.ContentType
		}
		if opts.FullCommit {
			req.Header.Add("X-Couch-Full-Commit", "true")
		}
		if opts.IfNoneMatch != "" {
			inm := "\"" + strings.Trim(opts.IfNoneMatch, "\"") + "\""
			req.Header.Set("If-None-Match", inm)
		}
		if opts.ContentLength != 0 {
			req.ContentLength = opts.ContentLength
		}
		for k, v := range opts.Header {
			if _, ok := req.Header[k]; !ok {
				req.Header[k] = v
			}
		}
	}
	req.Header.Add("Accept", accept)
	req.Header.Add("Content-Type", contentType)
}

func setQuery(req *http.Request, opts *Options) {
	if opts == nil || len(opts.Query) == 0 {
		return
	}
	if req.URL.RawQuery == "" {
		req.URL.RawQuery = opts.Query.Encode()
		return
	}
	req.URL.RawQuery = strings.Join([]string{req.URL.RawQuery, opts.Query.Encode()}, "&")
}

// DoError is the same as DoReq(), followed by checking the response error. This
// method is meant for cases where the only information you need from the
// response is the status code. It unconditionally closes the response body.
func (c *Client) DoError(ctx context.Context, method, path string, opts *Options) (*http.Response, error) {
	res, err := c.DoReq(ctx, method, path, opts)
	if err != nil {
		return res, err
	}
	if res.Body != nil {
		defer CloseBody(res.Body)
	}
	err = ResponseError(res)
	return res, err
}

// ETag returns the unquoted ETag value, and a bool indicating whether it was
// found.
func ETag(resp *http.Response) (string, bool) {
	if resp == nil {
		return "", false
	}
	etag, ok := resp.Header["Etag"]
	if !ok {
		etag, ok = resp.Header["ETag"] // nolint: staticcheck
	}
	if !ok {
		return "", false
	}
	return strings.Trim(etag[0], `"`), ok
}

// GetRev extracts the revision from the response's Etag header
func GetRev(resp *http.Response) (rev string, err error) {
	if err = ResponseError(resp); err != nil {
		return "", err
	}
	rev, ok := ETag(resp)
	if ok {
		return rev, nil
	}
	return extractRev(resp)
}

// When the ETag header is missing, which can happen, for example, when doing
// a request with revs_info=true.  This means we need to look through the
// body of the request for the revision. Fortunately, CouchDB tends to send
// the _id and _rev fields first, so we shouldn't need to parse the entire
// body. The important thing is that resp.Body must be restored, so that the
// normal document scanning can take place as usual.
func extractRev(resp *http.Response) (string, error) {
	if resp == nil || resp.Request == nil || resp.Request.Method == http.MethodHead {
		return "", errors.New("unable to determine document revision")
	}
	buf := &bytes.Buffer{}
	r := io.TeeReader(resp.Body, buf)
	defer func() {
		// Restore the original resp.Body
		resp.Body = struct {
			io.Reader
			io.Closer
		}{
			Reader: io.MultiReader(buf, resp.Body),
			Closer: resp.Body,
		}
	}()
	rev, err := readRev(r)
	if err != nil {
		return "", fmt.Errorf("unable to determine document revision: %w", err)
	}
	return rev, nil
}

// readRev searches r for a `_rev` field, and returns its value without reading
// the rest of the JSON stream.
func readRev(r io.Reader) (string, error) {
	dec := json.NewDecoder(r)
	tk, err := dec.Token()
	if err != nil {
		return "", err
	}
	if tk != json.Delim('{') {
		return "", fmt.Errorf("Expected %q token, found %q", '{', tk)
	}
	for dec.More() {
		tk, err = dec.Token()
		if err != nil {
			return "", err
		}
		if tk == "_rev" {
			tk, err = dec.Token()
			if err != nil {
				return "", err
			}
			if value, ok := tk.(string); ok {
				return value, nil
			}
			return "", fmt.Errorf("found %q in place of _rev value", tk)
		}
	}

	return "", errors.New("_rev key not found in response body")
}

func (c *Client) userAgent() string {
	ua := fmt.Sprintf("%s/%s (Language=%s; Platform=%s/%s)",
		UserAgent, Version, runtime.Version(), runtime.GOARCH, runtime.GOOS)
	return strings.Join(append([]string{ua}, c.UserAgents...), " ")
}
