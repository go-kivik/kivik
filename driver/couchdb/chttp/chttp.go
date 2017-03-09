// Package chttp provides a minimal HTTP driver backend for communicating with
// CouchDB servers.
package chttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
)

// Client represents a client connection.
type Client struct {
	// HTTPClient is a *http.Client, used for handling all requests.
	HTTPClient *http.Client

	dsn *url.URL
}

// DefaultHTTPClient is the default *http.Client to be used when none is
// specified.
var DefaultHTTPClient = &http.Client{}

// New returns a connection to a remote CouchDB server.
func New(dsn string) (*Client, error) {
	dsnURL, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	return &Client{
		HTTPClient: DefaultHTTPClient,
		dsn:        dsnURL,
	}, nil
}

// Options are optional parameters which may be sent with a request.
type Options struct {
	// Accept sets the request's Accept header.
	Accept string
	// ContentType sets the requests's Content-Type header
	ContentType string
	// Body sets the body of the request.
	Body io.Reader
	// JSON is an arbitrary data type which is marshaled to the request's body.
	// It an error to set both Body and JSON on the same request.
	JSON interface{}
}

// Response represents a response from a CouchDB server.
type Response struct {
	// StatusCode is the HTTP status code returned in the response.
	StatusCode int
	// Body is the body of the response.
	Body io.ReadCloser
	// ContentType is the base content type, parsed from the response headers.
	ContentType string
	// ContentParams are any parameters supplied in the content type headers.
	ContentParams map[string]string
	// CacheControl is the content of the Cache-Control header.
	CacheControl string
	// ContentLength is the content of the Content-Length header.
	ContentLength string
	// Etag is the content of the Etag header.
	Etag string
	// TransferEncoding is the content of the Transfer-Encoding header.
	TransferEncoding string
}

// Do does an HTTP request. An error is returned only if there was an error
// processing the request. In particular, an error status code, such as 400
// or 500, does _not_ cause an error to be returned.
func (c *Client) Do(method, path string, opts *Options) (*Response, error) {
	if opts.JSON != nil && opts.Body != nil {
		return nil, errors.New("must not specify both Body and JSON options")
	}
	reqURL := *c.dsn // Make a copy
	reqURL.Path = path
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
		}
	}
	req, err := http.NewRequest(method, reqURL.String(), body)
	if err != nil {
		return nil, err
	}
	if opts != nil {
		if opts.Accept != "" {
			req.Header.Add("Accept", opts.Accept)
		}
		if opts.ContentType != "" {
			req.Header.Add("Content-Type", opts.ContentType)
		}
	}
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	ct, ctParams, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	return &Response{
		StatusCode:       res.StatusCode,
		Body:             res.Body,
		ContentType:      ct,
		ContentParams:    ctParams,
		CacheControl:     res.Header.Get("Cache-Control"),
		ContentLength:    res.Header.Get("Content-Length"),
		Etag:             res.Header.Get("Etag"),
		TransferEncoding: res.Header.Get("Transfer-Encoding"),
	}, nil
}
