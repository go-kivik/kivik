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
	// to http.DefaultClient
	HTTPClient *http.Client
}

var _ http.Handler = &Proxy{}

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
		return http.DefaultClient
	}
	return p.HTTPClient
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest(r.Method, p.url(r.URL), r.Body)
	if err != nil {
		proxyError(w, err)
	}
	req = req.WithContext(r.Context())
	res, err := p.client().Do(req)
	if err != nil {
		proxyError(w, err)
	}
	defer res.Body.Close()
	for header, values := range res.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}
	w.WriteHeader(res.StatusCode)
	if _, err := io.Copy(w, res.Body); err != nil {
		proxyError(w, err)
	}
}

func proxyError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "Proxy error: %s", err)
}
