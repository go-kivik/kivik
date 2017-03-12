// Package cookie provides standard CouchDB cookie auth as described at
// http://docs.couchdb.org/en/2.0.0/api/server/authn.html#cookie-authentication
package cookie

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/auth"
	"github.com/flimzy/kivik/authdb"
	"github.com/flimzy/kivik/errors"
	"github.com/flimzy/kivik/serve"
)

// Auth provides CouchDB Cookie authentication.
type Auth struct{}

var _ auth.Handler = &Auth{}

// MethodName returns "cookie"
func (a *Auth) MethodName() string {
	return "cookie" // For compatibility with the name used by CouchDB
}

// Authenticate authenticates a request with cookie auth against the user store.
func (a *Auth) Authenticate(r *http.Request, store authdb.UserStore) (*authdb.UserContext, error) {
	cookie, err := r.Cookie(serve.SessionCookieName)
	if err != nil {
		return nil, nil
	}
	name, t, err := decodeCookie(cookie.Value)
	if err != nil {
		// Invalid cookie, continue as though there is no cookie
		return nil, nil
	}
	user, err := store.UserCtx(r.Context(), name)
	if err != nil {
		// Failed to look up the user
		return nil, nil
	}
	s := serve.GetService(r)
	token, err := s.CreateAuthToken(r.Context(), name, user.Salt, t)
	if err != nil {
		return nil, nil
	}
	if token != cookie.Value {
		return nil, errors.Status(kivik.StatusUnauthorized, "bad cookie")
	}
	return user, nil
}

func decodeCookie(cookie string) (name string, created int64, err error) {
	data, err := base64.RawURLEncoding.DecodeString(cookie)
	if err != nil {
		return "", 0, err
	}
	parts := bytes.Split(data, []byte(":"))
	if len(parts) != 3 {
		return "", 0, errors.New("invalid cookie")
	}
	t, err := strconv.ParseInt(string(parts[1]), 16, 64)
	if err != nil {
		return "", 0, errors.Wrap(err, "invalid timestamp")
	}
	return string(parts[0]), t, nil
}
