// Package proxy provides a simple proxy for a CouchDB server
package proxy

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Proxy is an http.Handler which proxies connections to a CouchDB server.
type Proxy struct {
	baseURL *url.URL
	// path is the url.Path with trailing slash removed
	path string
	// HTTPClient is the HTTP client used to make requests. If unset, defaults
	// DefaultClient, which is distinct from http.DefaultClient in that it does
	// not follow redirects.
	HTTPClient *http.Client
	// StrictMethods will reject any non-standard CouchDB methods immediately,
	// rather than relaying to the CouchDB server.
	StrictMethods bool
}

var _ http.Handler = &Proxy{}

// DefaultClient is the default http.Client used to make requests to the
// backend CouchDB server.
var DefaultClient = &http.Client{
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// New returns a new Proxy instance, which redirects all requests to the
// specified URL.
func New(serverURL string) (*Proxy, error) {
	if serverURL == "" {
		return nil, errors.New("no URL specified")
	}
	parsed, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	if parsed.User != nil {
		return nil, errors.New("proxy URL must not contain auth credentials")
	}
	return &Proxy{
		baseURL: parsed,
		path:    strings.TrimSuffix(parsed.Path, "/"),
	}, nil
}

// url maps the request's URL to the backend server.
func (p *Proxy) url(reqURL *url.URL) string {
	newURL := url.URL{
		Scheme:   p.baseURL.Scheme,
		User:     reqURL.User,
		Host:     p.baseURL.Host,
		RawQuery: reqURL.RawQuery,
	}
	newURL.Path = p.path + "/" + strings.TrimPrefix(reqURL.Path, "/")
	return newURL.String()
}

func (p *Proxy) client() *http.Client {
	if p.HTTPClient == nil {
		return DefaultClient
	}
	return p.HTTPClient
}

// Any other methods are rejected immediately, if StrictMethods is true.
var supportedMethods = []string{"GET", "HEAD", "POST", "PUT", "DELETE", "COPY"}

func (p *Proxy) methodAllowed(method string) bool {
	if !p.StrictMethods {
		return true
	}
	for _, m := range supportedMethods {
		if m == method {
			return true
		}
	}
	return false
}

// cloneRequest clones the original request, for proxying
func (p *Proxy) cloneRequest(r *http.Request) (*http.Request, error) {
	req, err := http.NewRequest(r.Method, p.url(r.URL), r.Body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(r.Context())
	for header, values := range r.Header {
		for _, value := range values {
			req.Header.Set(header, value)
		}
	}
	return req, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !p.methodAllowed(r.Method) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	req, err := p.cloneRequest(r)
	if err != nil {
		ProxyError(w, err)
		return
	}
	res, err := p.client().Do(req)
	if err != nil {
		ProxyError(w, err)
		return
	}
	defer res.Body.Close()
	for header, values := range res.Header {
		for _, value := range values {
			if err := p.setHeader(w, res, r, header, value); err != nil {
				ProxyError(w, err)
				return
			}
		}
	}
	w.WriteHeader(res.StatusCode)
	if _, err := io.Copy(w, res.Body); err != nil {
		ProxyError(w, err)
	}
}

// setHeader copies the response header to the responsewriter, possibly modifying
// it based on the request.
func (p *Proxy) setHeader(w http.ResponseWriter, res *http.Response, req *http.Request, header, value string) error {
	switch header {
	case "Location":
		locURL, err := url.Parse(value)
		if err != nil {
			return err
		}
		newURL := &url.URL{
			Scheme:   req.URL.Scheme,
			Host:     req.URL.Host,
			Path:     locURL.Path,
			RawQuery: req.URL.RawQuery,
		}
		value = newURL.String()
	}
	w.Header().Add(header, value)
	return nil
}

// ProxyError returns a generic error to the client.
func ProxyError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "Proxy error: %s", err)
}
