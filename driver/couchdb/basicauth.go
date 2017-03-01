package couchdb

import "net/http"

// BasicAuth provides basic HTTP Authentication services.
type BasicAuth struct {
	Name      string
	Password  string
	transport http.RoundTripper
}

// RoundTrip fulfills the http.RoundTripper interface. It sets HTTP Basic Auth
// on
func (a *BasicAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(a.Name, a.Password)
	transport := a.transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	return transport.RoundTrip(req)
}

// Authenticate sets HTTP Basic Auth on the request.
func (a *BasicAuth) authenticate(c *client) error {
	transport := c.httpClient.Transport
	if auth, ok := transport.(*BasicAuth); ok {
		transport = auth.transport
	}
	a.transport = transport
	c.httpClient.Transport = a
	return nil
}
