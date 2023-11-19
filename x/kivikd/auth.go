package kivikd

import (
	"net/http"

	"github.com/go-kivik/kivikd/v4/auth"
	"github.com/go-kivik/kivikd/v4/authdb"
)

type doneWriter struct {
	http.ResponseWriter
	done bool
}

func (w *doneWriter) WriteHeader(status int) {
	w.done = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *doneWriter) Write(b []byte) (int, error) {
	w.done = true
	return w.ResponseWriter.Write(b)
}

func authHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dw := &doneWriter{ResponseWriter: w}
		s := GetService(r)
		session, err := s.validate(dw, r)
		if err != nil {
			reportError(w, err)
			return
		}
		sessionPtr := mustGetSessionPtr(r.Context())
		*sessionPtr = session
		if dw.done {
			// The auth handler already responded to the request
			return
		}
		next.ServeHTTP(w, r)
	})
}

// validate must return a 401 error if there is an authentication failure.
// No error means the user is permitted.
func (s *Service) validate(w http.ResponseWriter, r *http.Request) (*auth.Session, error) {
	if s.authHandlers == nil {
		// Perpetual admin party
		return s.createSession("", &authdb.UserContext{Roles: []string{"_admin"}}), nil
	}
	for methodName, handler := range s.authHandlers {
		uCtx, err := handler.Authenticate(w, r)
		if err != nil {
			return nil, err
		}
		if uCtx != nil {
			return s.createSession(methodName, uCtx), nil
		}
	}
	// None of the auth methods succeeded, so return unauthorized
	return s.createSession("", nil), nil
}

func (s *Service) createSession(method string, user *authdb.UserContext) *auth.Session {
	return &auth.Session{
		AuthMethod: method,
		AuthDB:     s.Conf().GetString("couch_httpd_auth.authentication_db"),
		Handlers:   s.authHandlerNames,
		User:       user,
	}
}
