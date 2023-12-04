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
// +build !js

package kivikd

import (
	"context"
	"net/http"

	"github.com/go-kivik/kivik/v4/x/kivikd/auth"
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
