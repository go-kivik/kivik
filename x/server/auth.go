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

package server

import (
	"context"
	"net/http"

	"gitlab.com/flimzy/httpe"

	"github.com/go-kivik/kivik/v4/x/server/auth"
)

type contextKey struct{ name string }

var userContextKey = &contextKey{"userCtx"}

func (s *Server) UserStore() auth.UserStore {
	return nil
	// return s.userStore
}

func (s *Server) ValidateCookie(user *auth.UserContext, cookie string) (bool, error) {
	return false, nil
}

func (s *Server) authMiddleware(next httpe.HandlerWithError) httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		if len(s.authFuncs) == 0 {
			// Admin party!
			r = r.WithContext(context.WithValue(ctx, userContextKey, &auth.UserContext{
				Name:  "admin",
				Roles: []string{"_admin"},
			}))
			return next.ServeHTTPWithError(w, r)
		}

		var userCtx *auth.UserContext
		var err error
		for _, authFunc := range s.authFuncs {
			userCtx, err = authFunc(w, r)
			if err != nil {
				return err
			}
			if userCtx != nil {
				break
			}
		}
		if userCtx == nil {
			return &couchError{status: http.StatusUnauthorized, Err: "unauthorized", Reason: "Authentication required."}
		}
		r = r.WithContext(context.WithValue(ctx, userContextKey, userCtx))
		return next.ServeHTTPWithError(w, r)
	})
}

// func (s *Server) startSession() httpe.HandlerWithError {
// 	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
// 		var req struct {
// 			Name     *string `json:"name" form:"name"`
// 			Password string  `json:"password" form:"password"`
// 		}
// 		if err := s.bind(r, &req); err != nil {
// 			return err
// 		}
// 		if req.Name == nil {
// 			return &couchError{status: http.StatusBadRequest, Err: "bad_request", Reason: "request body must contain a username"}
// 		}
// 		return nil
// 	})
// }
