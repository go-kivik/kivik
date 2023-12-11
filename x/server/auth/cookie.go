// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/internal"
)

type cookieAuth struct {
	secret  string
	timeout time.Duration
	s       Server
}

// CookieAuth returns a cookie auth handler.
func CookieAuth(secret string, sessionTimeout time.Duration) Handler {
	return &cookieAuth{
		secret:  secret,
		timeout: sessionTimeout,
	}
}

func (a *cookieAuth) Init(s Server) (string, AuthenticateFunc) {
	a.s = s
	return "cookie", // For compatibility with the name used by CouchDB
		a.Authenticate
}

func (a *cookieAuth) Authenticate(w http.ResponseWriter, r *http.Request) (*UserContext, error) {
	if r.URL.Path == "/_session" {
		switch r.Method {
		case http.MethodPost:
			return nil, a.postSession(w, r)
		case http.MethodDelete:
			return nil, deleteSession(w)
		}
	}
	return a.validateCookie(r)
}

func (a *cookieAuth) validateCookie(r *http.Request) (*UserContext, error) {
	cookie, err := r.Cookie(kivik.SessionCookieName)
	if err != nil {
		return nil, nil
	}
	name, _, err := DecodeCookie(cookie.Value)
	if err != nil {
		return nil, nil
	}
	user, err := a.s.UserStore().UserCtx(r.Context(), name)
	if err != nil {
		// Failed to look up the user
		return nil, err
	}
	valid, err := a.ValidateCookie(user, cookie.Value)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, &internal.Error{Status: http.StatusUnauthorized, Message: "invalid cookie"}
	}
	return user, nil
}

func (a *cookieAuth) postSession(w http.ResponseWriter, r *http.Request) error {
	var authData struct {
		Name     *string `form:"name" json:"name"`
		Password string  `form:"password" json:"password"`
	}
	if err := a.s.Bind(r, &authData); err != nil {
		return err
	}
	if authData.Name == nil {
		return &internal.Error{Status: http.StatusBadRequest, Message: "request body must contain a username"}
	}
	user, err := a.s.UserStore().Validate(r.Context(), *authData.Name, authData.Password)
	if err != nil {
		return err
	}
	next, err := redirectURL(r)
	if err != nil {
		return err
	}

	// Success, so create a cookie
	token := CreateAuthToken(*authData.Name, user.Salt, a.secret, time.Now().Unix())
	w.Header().Set("Cache-Control", "must-revalidate")
	http.SetCookie(w, &http.Cookie{
		Name:     kivik.SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(a.timeout.Seconds()),
		HttpOnly: true,
	})
	w.Header().Add("Content-Type", typeJSON)
	if next != "" {
		w.Header().Add("Location", next)
		w.WriteHeader(http.StatusFound)
	}
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":    true,
		"name":  user.Name,
		"roles": user.Roles,
	})
}

func redirectURL(r *http.Request) (string, error) {
	next, ok := stringQueryParam(r, "next")
	if !ok {
		return "", nil
	}
	if !strings.HasPrefix(next, "/") {
		return "", &internal.Error{Status: http.StatusBadRequest, Message: "redirection url must be relative to server root"}
	}
	if strings.HasPrefix(next, "//") {
		// Possible schemaless url
		return "", &internal.Error{Status: http.StatusBadRequest, Message: "invalid redirection url"}
	}
	parsed, err := url.Parse(next)
	if err != nil {
		return "", &internal.Error{Status: http.StatusBadRequest, Message: "invalid redirection url"}
	}
	return parsed.String(), nil
}

func deleteSession(w http.ResponseWriter) error {
	http.SetCookie(w, &http.Cookie{
		Name:     kivik.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	w.Header().Add("Content-Type", typeJSON)
	w.Header().Set("Cache-Control", "must-revalidate")
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"ok": true,
	})
}

// CreateAuthToken hashes a username, salt, timestamp, and the server secret
// into an authentication token.
func CreateAuthToken(name, salt, secret string, time int64) string {
	if secret == "" {
		panic("secret must be set")
	}
	if salt == "" {
		panic("salt must be set")
	}
	sessionData := fmt.Sprintf("%s:%X", name, time)
	h := hmac.New(sha1.New, []byte(secret+salt))
	_, _ = h.Write([]byte(sessionData))
	hashData := string(h.Sum(nil))
	return base64.RawURLEncoding.EncodeToString([]byte(sessionData + ":" + hashData))
}

// stringQueryParam extracts a query parameter as string.
func stringQueryParam(r *http.Request, key string) (string, bool) {
	values := r.URL.Query()
	if _, ok := values[key]; !ok {
		return "", false
	}
	return values.Get(key), true
}

// DecodeCookie decodes a Base64-encoded cookie, and returns its component
// parts.
func DecodeCookie(cookie string) (name string, created int64, err error) {
	data, err := base64.RawURLEncoding.DecodeString(cookie)
	if err != nil {
		return "", 0, err
	}
	const partCount = 3
	parts := bytes.SplitN(data, []byte(":"), partCount)
	t, err := strconv.ParseInt(string(parts[1]), 16, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid timestamp: %w", err)
	}
	return string(parts[0]), t, nil
}

// ValidateCookie validates the provided cookie against the configured UserStore.
func (a *cookieAuth) ValidateCookie(user *UserContext, cookie string) (bool, error) {
	name, t, err := DecodeCookie(cookie)
	if err != nil {
		return false, err
	}
	token := CreateAuthToken(name, user.Salt, a.secret, t)
	return token == cookie, nil
}
