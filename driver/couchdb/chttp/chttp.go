// Package chttp provides a minimal HTTP driver backend for communicating with
// CouchDB servers.
package chttp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/url"

	"github.com/flimzy/kivik"
	"github.com/pkg/errors"
)

// Client represents a client connection. It embeds an *http.Client
type Client struct {
	*http.Client

	rawDSN string
	dsn    *url.URL
	auth   Authenticator
}

// New calls NewContext() with a Background context.
func New(dsn string) (*Client, error) {
	return NewContext(context.Background(), dsn)
}

// NewContext returns a connection to a remote CouchDB server. If credentials are
// included in the URL, CookieAuth is attempted first, with BasicAuth used as
// a fallback. If both fail, an error is returned. If you wish to use some other
// authentication mechanism, do not specify credentials in the URL, and instead
// call the Auth() method later.
func NewContext(ctx context.Context, dsn string) (*Client, error) {
	dsnURL, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	user := dsnURL.User
	dsnURL.User = nil
	c := &Client{
		Client: &http.Client{},
		dsn:    dsnURL,
		rawDSN: dsn,
	}
	if user != nil {
		password, _ := user.Password()
		if err := c.defaultAuth(ctx, user.Username(), password); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// DSN returns the unparsed DSN used to connect.
func (c *Client) DSN() string {
	return c.rawDSN
}

func (c *Client) defaultAuth(ctx context.Context, username, password string) error {
	err := c.Auth(ctx, &CookieAuth{
		Username: username,
		Password: password,
	})
	if err == nil {
		return nil
	}
	return c.Auth(ctx, &BasicAuth{
		Username: username,
		Password: password,
	})
}

// Auth authenticates using the provided Authenticator.
func (c *Client) Auth(ctx context.Context, a Authenticator) error {
	if c.auth != nil {
		return errors.New("auth already set; log out first")
	}
	if err := a.Authenticate(ctx, c); err != nil {
		return err
	}
	c.auth = a
	return nil
}

// Logout logs out after authentication.
func (c *Client) Logout(ctx context.Context) error {
	if c.auth == nil {
		return errors.New("not authenticated")
	}
	err := c.auth.Logout(ctx, c)
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
	// ForceCommit adds the X-Couch-Full-Commit: true header to requests
	ForceCommit bool
}

// Response represents a response from a CouchDB server.
type Response struct {
	*http.Response

	// ContentType is the base content type, parsed from the response headers.
	ContentType string
}

// DecodeJSON unmarshals the response body into i. This method consumes and
// closes the response body.
func (r *Response) DecodeJSON(i interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(i)
}

// DoJSON combines DoReq() and, ResponseError(), and (*Response).DecodeJSON(), and
// discards the response.
func (c *Client) DoJSON(ctx context.Context, method, path string, opts *Options, i interface{}) (*Response, error) {
	res, err := c.DoReq(ctx, method, path, opts)
	if err != nil {
		return res, err
	}
	if err = ResponseError(res.Response); err != nil {
		return res, err
	}
	return res, res.DecodeJSON(i)
}

// NewRequest returns a new *http.Request to the CouchDB server, and the
// specified path. The host, schema, etc, of the specified path are ignored.
func (c *Client) NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	reqPath, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	url := *c.dsn // Make a copy
	url.Path = reqPath.Path
	url.RawQuery = reqPath.RawQuery
	req, err := http.NewRequest(method, url.String(), body)
	if err != nil {
		return nil, err
	}
	return req.WithContext(ctx), nil
}

// DoReq does an HTTP request. An error is returned only if there was an error
// processing the request. In particular, an error status code, such as 400
// or 500, does _not_ cause an error to be returned.
func (c *Client) DoReq(ctx context.Context, method, path string, opts *Options) (*Response, error) {
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
			opts.ContentType = kivik.TypeJSON
		}
	}
	req, err := c.NewRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	setHeaders(req, opts)

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	ct, _, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err != nil {
		return nil, errors.Wrap(err, "invalid Content-Type header in HTTP response")
	}
	return &Response{
		Response:    res,
		ContentType: ct,
	}, nil
}

func setHeaders(req *http.Request, opts *Options) {
	accept := kivik.TypeJSON
	contentType := kivik.TypeJSON
	if opts != nil {
		if opts.Accept != "" {
			accept = opts.Accept
		}
		if opts.ContentType != "" {
			contentType = opts.ContentType
		}
		if opts.ForceCommit {
			req.Header.Add("X-Couch-Full-Commit", "true")
		}
	}
	req.Header.Add("Accept", accept)
	req.Header.Add("Content-Type", contentType)
}

// DoError is the same as DoReq(), followed by checking the response error. This
// method is meant for cases where the only information you need from the
// response is the status code. It unconditionally closes the response body.
func (c *Client) DoError(ctx context.Context, method, path string, opts *Options) (*Response, error) {
	res, err := c.DoReq(ctx, method, path, opts)
	if err != nil {
		return res, err
	}
	defer res.Response.Body.Close()
	return res, ResponseError(res.Response)
}
