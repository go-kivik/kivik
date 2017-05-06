// Package proxy provides a simple proxy for a CouchDB server
package proxy

import (
	"errors"
	"net/url"
)

// Proxy is an http.Handler which proxies connections to a CouchDB server.
type Proxy struct {
	baseURL *url.URL
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
	return &Proxy{baseURL: parsed}, nil
}
