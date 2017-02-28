package couchdb

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"

	"golang.org/x/net/publicsuffix"
)

// CookieAuth provides CouchDB Cookie auth services as described at
// http://docs.couchdb.org/en/2.0.0/api/server/authn.html#cookie-authentication
type CookieAuth struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	cookie   string
}

// Authenticate initiates a session, and sets HTTP Cookie headers.
func (a *CookieAuth) authenticate(c *client) error {
	if jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List}); err == nil {
		c.httpClient.Jar = jar
	} else {
		return err
	}
	var body io.Reader
	if b, err := json.Marshal(a); err == nil {
		body = bytes.NewBuffer(b)
	} else {
		panic(err)
	}

	_, err := c.newRequest(http.MethodPost, "/_session").
		AddHeader("Content-Type", jsonType).
		AddHeader("Accept", jsonType).
		Body(body).
		Do()
	return err
}

// SetCookie sets the cookie directly. This can be used if authentication is
// performed by some non-standard means.
func (a *CookieAuth) SetCookie(cookie string) {
	a.cookie = cookie
}
