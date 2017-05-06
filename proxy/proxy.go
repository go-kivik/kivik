// Package proxy provides a simple proxy for a CouchDB server
package proxy

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
)

// Proxy is an http.Handler which proxies connections to a CouchDB server.
type Proxy struct {
	baseURL *url.URL
	// path is the url.Path with trailing slash removed
	path string
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

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
