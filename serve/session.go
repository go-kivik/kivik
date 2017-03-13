package serve

import (
	"encoding/json"
	"net/http"

	"golang.org/x/net/context"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/auth"
	"github.com/flimzy/kivik/errors"
)

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

func setSession() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			// We store a pointer to a pointer, so the underlying pointer can
			// be updated by the auth process, without losing the reference.
			session := &auth.Session{}
			ctx = context.WithValue(ctx, SessionKey, &session)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
