package serve

import (
	"context"
	"net/http"

	"github.com/flimzy/kivik/auth"
)

// DefaultInsecureSecret is the hash secret used if couch_httpd_auth.secret
// is unconfigured. Please configure couch_httpd_auth.secret, or they're all
// gonna laugh at you!
const DefaultInsecureSecret = "They're all gonna laugh at you!"

// DefaultSessionTimeout is the default session timeout, in seconds, used if
// couch_httpd_auth.timeout is inuset.
const DefaultSessionTimeout = 600

func (s *Service) getAuthSecret() string {
	if s.Conf().IsSet("couch_httpd_auth.secret") {
		return s.Conf().GetString("couch_httpd_auth.secret")
	}
	return DefaultInsecureSecret
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
