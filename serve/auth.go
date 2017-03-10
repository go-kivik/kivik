package serve

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
)

func authHandler(s *Service, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := s.validate(r); err != nil {
			reportError(w, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// validate must return a 401 error if there is an authentication failure.
// No error means the user is permitted.
func (s *Service) validate(r *http.Request) error {
	if r.URL.Path == "/" {
		return nil
	}
	if r.URL.Path == "/_session" && r.Method == http.MethodGet {
		return nil
	}
	if r.URL.Path == "/_uuids" {
		return nil
	}
	adminParty, err := s.isAdminParty(r.Context())
	if err != nil {
		return err
	}
	if adminParty {
		return nil
	}
	if username, password, ok := r.BasicAuth(); ok {
		valid, err := s.AuthHandler.Validate(r.Context(), username, password)
		if err != nil {
			s.Error("AuthHandler failed for username '%s': %s", username, err)
			return errors.Status(kivik.StatusInternalServerError, "authentication failure")
		}
		if valid {
			return nil
		}
	}
	return errors.Status(kivik.StatusUnauthorized, "not authorized")
}

func (s *Service) isAdminParty(ctx context.Context) (bool, error) {
	if s.AuthHandler == nil {
		// Perpetual Admin Party!
		return true, nil
	}
	sec, err := s.Config().GetSectionContext(ctx, "admins")
	return len(sec) == 0, err
}
