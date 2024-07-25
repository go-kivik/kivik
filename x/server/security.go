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

package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gitlab.com/flimzy/httpe"

	"github.com/go-kivik/kivik/v4"
)

func (s *Server) getSecurity() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		dbName := chi.URLParam(r, "db")
		security, err := s.client.DB(dbName).Security(r.Context())
		if err != nil {
			return err
		}
		return serveJSON(w, http.StatusOK, security)
	})
}

func (s *Server) putSecurity() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		dbName := chi.URLParam(r, "db")
		var security kivik.Security
		if err := s.bind(r, &security); err != nil {
			return err
		}
		if err := s.client.DB(dbName).SetSecurity(r.Context(), &security); err != nil {
			return err
		}
		return serveJSON(w, http.StatusOK, map[string]bool{
			"ok": true,
		})
	})
}
