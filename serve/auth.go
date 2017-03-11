package serve

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/auth"
	"github.com/flimzy/kivik/authdb"
	"github.com/flimzy/kivik/errors"
)

func authHandler(s *Service, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := s.validate(r)
		if err != nil && errors.StatusCode(err) != kivik.StatusUnauthorized {
			reportError(w, err)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, SessionKey, session)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// validate must return a 401 error if there is an authentication failure.
// No error means the user is permitted.
func (s *Service) validate(r *http.Request) (*auth.Session, error) {
	if s.authHandlers == nil {
		// Perpetual admin party
		return s.createSession("", &authdb.UserContext{Roles: []string{"_admin"}}), nil
	}
	for methodName, handler := range s.authHandlers {
		uCtx, err := handler.Authenticate(r, s.UserStore)
		switch {
		case errors.StatusCode(err) == kivik.StatusUnauthorized:
			continue
		case err != nil:
			return nil, err
		default:
			return s.createSession(methodName, uCtx), nil
		}
	}
	// None of the auth methods succeeded, so return unauthorized
	return s.createSession("", nil), nil
}

func (s *Service) createSession(method string, user *authdb.UserContext) *auth.Session {
	return &auth.Session{
		AuthMethod: method,
		Handlers:   s.authHandlerNames,
		User:       user,
	}
}
