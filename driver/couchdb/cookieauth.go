package couchdb

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/davecgh/go-spew/spew"
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
	var body io.Reader
	if b, err := json.Marshal(a); err == nil {
		body = bytes.NewBuffer(b)
	} else {
		panic(err)
	}

	resp, err := c.newRequest(http.MethodPost, "/_session").
		AddHeader("Content-Type", jsonType).
		AddHeader("Accept", jsonType).
		Body(body).
		Do()
	if err != nil {
		return err
	}
	spew.Dump(resp)

	return nil
}

// SetCookie sets the cookie directly. This can be used if authentication is
// performed by some non-standard means.
func (a *CookieAuth) SetCookie(cookie string) {
	a.cookie = cookie
}
