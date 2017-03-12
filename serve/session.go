package serve

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
)

// SessionCookieName is the name of the CouchDB session cookie.
const SessionCookieName = "AuthSession"

// DefaultInsecureSecret is the hash secret used if couch_httpd_auth.secret
// is unconfigured. Please configure couch_httpd_auth.secret, or they're all
// gonna laugh at you!
const DefaultInsecureSecret = "They're all gonna laugh at you!"

// DefaultSessionTimeout is the default session timeout, in seconds, used if
// couch_httpd_auth.timeout is inuset.
const DefaultSessionTimeout = 600

func getSession(w http.ResponseWriter, r *http.Request) error {
	session := MustGetSession(r.Context())
	w.Header().Add("Content-Type", typeJSON)
	return json.NewEncoder(w).Encode(session)
}

func (s *Service) getAuthSecret(ctx context.Context) (string, error) {
	secret, err := s.Config().GetContext(ctx, "couch_httpd_auth", "secret")
	if errors.StatusCode(err) == kivik.StatusNotFound {
		return DefaultInsecureSecret, nil
	}
	if err != nil {
		return "", err
	}
	return secret, nil
}

func (s *Service) getSessionTimeout(ctx context.Context) (int, error) {
	timeout, err := s.Config().GetContext(ctx, "couch_httpd_auth", "timeout")
	if errors.StatusCode(err) == kivik.StatusNotFound {
		return DefaultSessionTimeout, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(timeout)
}

func postSession(w http.ResponseWriter, r *http.Request) error {
	authData := struct {
		Name     *string `form:"name" json:"name"`
		Password string  `form:"password" json:"password"`
	}{}
	if err := BindParams(r, &authData); err != nil {
		return errors.Status(kivik.StatusBadRequest, "unable to parse request data")
	}
	if authData.Name == nil {
		return errors.Status(kivik.StatusBadRequest, "request body must contain a username")
	}
	s := GetService(r)
	user, err := s.UserStore.Validate(r.Context(), *authData.Name, authData.Password)
	if err != nil {
		return err
	}
	timeout, err := s.getSessionTimeout(r.Context())
	if err != nil {
		return err
	}
	next, err := redirectURL(r)
	if err != nil {
		return err
	}

	// Success, so create a cookie
	token, err := s.CreateAuthToken(r.Context(), *authData.Name, user.Salt, time.Now().Unix())
	if err != nil {
		return err
	}
	w.Header().Set("Cache-Control", "must-revalidate")
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   timeout,
		HttpOnly: true,
	})
	w.Header().Add("Content-Type", typeJSON)
	if next != "" {
		w.Header().Add("Location", next)
		w.WriteHeader(kivik.StatusFound)
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
		return "", errors.Status(kivik.StatusBadRequest, "redirection url must be relative to server root")
	}
	parsed, err := url.Parse(next)
	if err != nil {
		return "", errors.Status(kivik.StatusBadRequest, "invalid redirection url")
	}
	return parsed.String(), nil
}

func deleteSession(w http.ResponseWriter, r *http.Request) error {
	return nil
}
