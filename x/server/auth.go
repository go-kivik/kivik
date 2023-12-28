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

package server

import (
	"context"
	"net/http"

	"gitlab.com/flimzy/httpe"

	"github.com/go-kivik/kivik/v4/internal"
	"github.com/go-kivik/kivik/v4/x/server/auth"
)

type contextKey struct{ name string }

var userContextKey = &contextKey{"userCtx"}

type authService struct {
	s *Server
}

var _ auth.Server = (*authService)(nil)

// UserStore returns the aggregate UserStore for the server.
func (s *authService) UserStore() auth.UserStore {
	return s.s.userStores
}

func (s *authService) Bind(r *http.Request, v interface{}) error {
	return s.s.bind(r, v)
}

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

// authMiddleware sets the user context based on the authenticated user, if any.
func (s *Server) authMiddleware(next httpe.HandlerWithError) httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		if len(s.authFuncs) == 0 {
			// Admin party!
			r = r.WithContext(context.WithValue(ctx, userContextKey, &auth.UserContext{
				Name:  "admin",
				Roles: []string{auth.RoleAdmin},
			}))
			return next.ServeHTTPWithError(w, r)
		}

		dw := &doneWriter{ResponseWriter: w}

		var userCtx *auth.UserContext
		var err error
		for _, authFunc := range s.authFuncs {
			userCtx, err = authFunc(dw, r)
			if err != nil {
				return err
			}
			if dw.done {
				return nil
			}
			if userCtx != nil {
				break
			}
		}
		r = r.WithContext(context.WithValue(ctx, userContextKey, userCtx))
		return next.ServeHTTPWithError(w, r)
	})
}

// adminRequired returns Status Forbidden if the session is not authenticated as
// an admin.
func adminRequired(next httpe.HandlerWithError) httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		userCtx, _ := r.Context().Value(userContextKey).(*auth.UserContext)
		if userCtx == nil {
			return &internal.Error{Status: http.StatusUnauthorized, Message: "User not authenticated"}
		}
		if !userCtx.HasRole(auth.RoleAdmin) {
			return &internal.Error{Status: http.StatusForbidden, Message: "Admin privileges required"}
		}
		return next.ServeHTTPWithError(w, r)
	})
}
