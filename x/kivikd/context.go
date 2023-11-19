package kivikd

import (
	"context"
	"net/http"

	"github.com/go-kivik/kivikd/v4/auth"
)

type contextKey struct {
	name string
}

var (
	// SessionKey is a context key used to access the authenticated session.
	SessionKey = &contextKey{"session"}
	// ClientContextKey is a context key used to access the kivik client.
	ClientContextKey = &contextKey{"client"}
	// ServiceContextKey is a context key used to access the serve.Service struct.
	ServiceContextKey = &contextKey{"service"}
)

func setContext(s *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, ClientContextKey, s.Client)
			ctx = context.WithValue(ctx, ServiceContextKey, s)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// MustGetSession returns the user context for the currently authenticated user.
// If no session is set, the function panics.
func MustGetSession(ctx context.Context) *auth.Session {
	s, ok := ctx.Value(SessionKey).(**auth.Session)
	if !ok {
		panic("No session!")
	}
	return *s
}

func mustGetSessionPtr(ctx context.Context) **auth.Session {
	s, ok := ctx.Value(SessionKey).(**auth.Session)
	if !ok {
		panic("No session!")
	}
	return s
}

// GetService extracts the Kivik service from the request.
func GetService(r *http.Request) *Service {
	service := r.Context().Value(ServiceContextKey).(*Service)
	return service
}
