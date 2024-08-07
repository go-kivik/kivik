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

//go:build !js

// Package cookie provides standard CouchDB cookie auth as described at
// http://docs.couchdb.org/en/2.0.0/api/server/authn.html#cookie-authentication
package cookie

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-kivik/kivik/v4"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/x/kivikd"
	"github.com/go-kivik/kivik/v4/x/kivikd/auth"
	"github.com/go-kivik/kivik/v4/x/kivikd/authdb"
	"github.com/go-kivik/kivik/v4/x/kivikd/cookies"
)

const typeJSON = "application/json"

// Auth provides CouchDB Cookie authentication.
type Auth struct{}

var _ auth.Handler = &Auth{}

// MethodName returns "cookie"
func (a *Auth) MethodName() string {
	return "cookie" // For compatibility with the name used by CouchDB
}

// Authenticate authenticates a request with cookie auth against the user store.
func (a *Auth) Authenticate(w http.ResponseWriter, r *http.Request) (*authdb.UserContext, error) {
	if r.URL.Path == "/_session" {
		switch r.Method {
		case http.MethodPost:
			return nil, postSession(w, r)
		case http.MethodDelete:
			return nil, deleteSession(w)
		}
	}
	return a.validateCookie(r)
}

func (a *Auth) validateCookie(r *http.Request) (*authdb.UserContext, error) {
	store := kivikd.GetService(r).UserStore
	cookie, err := r.Cookie(kivik.SessionCookieName)
	if err != nil {
		return nil, nil
	}
	name, _, err := cookies.DecodeCookie(cookie.Value)
	if err != nil {
		return nil, nil
	}
	user, err := store.UserCtx(r.Context(), name)
	if err != nil {
		// Failed to look up the user
		return nil, nil
	}
	s := kivikd.GetService(r)
	valid, err := s.ValidateCookie(user, cookie.Value)
	if err != nil || !valid {
		return nil, nil
	}
	return user, nil
}

func postSession(w http.ResponseWriter, r *http.Request) error {
	authData := struct {
		Name     *string `form:"name" json:"name"`
		Password string  `form:"password" json:"password"`
	}{}
	if err := kivikd.BindParams(r, &authData); err != nil {
		return &internal.Error{Status: http.StatusBadRequest, Message: "unable to parse request data"}
	}
	if authData.Name == nil {
		return &internal.Error{Status: http.StatusBadRequest, Message: "request body must contain a username"}
	}
	s := kivikd.GetService(r)
	user, err := s.UserStore.Validate(r.Context(), *authData.Name, authData.Password)
	if err != nil {
		return err
	}
	next, err := redirectURL(r)
	if err != nil {
		return err
	}

	// Success, so create a cookie
	token, err := s.CreateAuthToken(*authData.Name, user.Salt, time.Now().Unix())
	if err != nil {
		return err
	}
	w.Header().Set("Cache-Control", "must-revalidate")
	http.SetCookie(w, &http.Cookie{
		Name:     kivik.SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   getSessionTimeout(s),
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
	next, ok := kivikd.StringQueryParam(r, "next")
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

func getSessionTimeout(s *kivikd.Service) int {
	if s.Conf().IsSet("couch_httpd_auth.timeout") {
		return s.Conf().GetInt("couch_httpd_auth.timeout")
	}
	return kivikd.DefaultSessionTimeout
}
