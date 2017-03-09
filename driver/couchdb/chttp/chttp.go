// Package chttp provides a minimal HTTP driver backend for communicating with
// CouchDB servers.
package chttp

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// HTTP methods supported by this library.
const (
	MethodGet    = "GET"
	MethodPut    = "PUT"
	MethodPost   = "POST"
	MethodHead   = "HEAD"
	MethodDelete = "DELETE"
	MethodCopy   = "COPY"
)

const (
	typeJSON = "application/json"
	typeText = "text/plain"
)

// Client represents a client connection. It embeds an *http.Client
type Client struct {
	*http.Client

	dsn  *url.URL
	auth Authenticator
}

// defaultHTTPClient is the default *http.Client to be used when none is
// specified.
var defaultHTTPClient = &http.Client{}

// New returns a connection to a remote CouchDB server.
func New(dsn string) (*Client, error) {
	dsnURL, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	return &Client{
		Client: defaultHTTPClient,
		dsn:    dsnURL,
	}, nil
}

// Auth authenticates using the provided Authenticator.
func (c *Client) Auth(a Authenticator) error {
	if c.auth != nil {
		return errors.New("auth already set; log out first")
	}
	if err := a.Authenticate(c); err != nil {
		return err
	}
	c.auth = a
	return nil
}

// Logout logs out after authentication.
func (c *Client) Logout() error {
	if c.auth == nil {
		return errors.New("not authenticated")
	}
	err := c.auth.Logout(c)
	c.auth = nil
	return err
}

// Options are optional parameters which may be sent with a request.
type Options struct {
	// Accept sets the request's Accept header. Defaults to "application/json".
	// To specify any, use "*/*".
	Accept string
	// ContentType sets the requests's Content-Type header. Defaults to "application/json".
	ContentType string
	// Body sets the body of the request.
	Body io.Reader
	// JSON is an arbitrary data type which is marshaled to the request's body.
	// It an error to set both Body and JSON on the same request. When this is
	// set, ContentType is unconditionally set to 'application/json'.
	JSON interface{}
}

// Response represents a response from a CouchDB server.
type Response struct {
	*http.Response

	// ContentType is the base content type, parsed from the response headers.
	ContentType string
	// ContentParams are any parameters supplied in the content type headers.
	ContentParams map[string]string
	// CacheControl is the content of the Cache-Control header.
	CacheControl string
	// Etag is the content of the Etag header.
	Etag string
	// TransferEncoding is the content of the Transfer-Encoding header.
	TransferEncoding string
}

// DecodeJSON unmarshals the response body into i. This method consumes and
// closes the response body.
func (r *Response) DecodeJSON(i interface{}) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	return dec.Decode(i)
}

// DoJSON combines DoReq() and, ResponseError(), and (*Response).DecodeJSON(), and
// discards the response. It is intended for cases where the only information
// needed from the response is the status code and JSON body.
func (c *Client) DoJSON(method, path string, opts *Options, i interface{}) error {
	res, err := c.DoReq(method, path, opts)
	if err != nil {
		return err
	}
	if err = ResponseError(res.Response); err != nil {
		return err
	}
	return res.DecodeJSON(i)
}

// NewRequest returns a new *http.Request to the CouchDB server, and the
// specified path. The host, schema, etc, of the specified path are ignored.
func (c *Client) NewRequest(method, path string, body io.Reader) (*http.Request, error) {
	reqPath, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	url := *c.dsn // Make a copy
	url.Path = reqPath.Path
	url.RawQuery = reqPath.RawQuery
	return http.NewRequest(method, url.String(), body)
}

// DoReq does an HTTP request. An error is returned only if there was an error
// processing the request. In particular, an error status code, such as 400
// or 500, does _not_ cause an error to be returned.
func (c *Client) DoReq(method, path string, opts *Options) (*Response, error) {
	if opts != nil && opts.JSON != nil && opts.Body != nil {
		return nil, errors.New("must not specify both Body and JSON options")
	}
	var body io.Reader
	if opts != nil {
		switch {
		case opts.Body != nil:
			body = opts.Body
		case opts.JSON != nil:
			buf := &bytes.Buffer{}
			body = buf
			enc := json.NewEncoder(buf)
			if err := enc.Encode(opts.JSON); err != nil {
				return nil, err
			}
			opts.ContentType = typeJSON
		}
	}
	req, err := c.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	setHeaders(req, opts)

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	ct, ctParams, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err != nil {
		return nil, errors.Wrap(err, "invalid Content-Type header in HTTP response")
	}
	return &Response{
		Response:         res,
		ContentType:      ct,
		ContentParams:    ctParams,
		CacheControl:     res.Header.Get("Cache-Control"),
		Etag:             res.Header.Get("Etag"),
		TransferEncoding: res.Header.Get("Transfer-Encoding"),
	}, nil
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
	}
	req.Header.Add("Accept", accept)
	req.Header.Add("Content-Type", contentType)
}

// DoError is the same as DoReq(), followed by checking the response error. This
// method is meant for cases where the only information you need from the
// response is the status code.
func (c *Client) DoError(method, path string, opts *Options) error {
	res, err := c.DoReq(method, path, opts)
	if err != nil {
		return err
	}
	return ResponseError(res.Response)
}
