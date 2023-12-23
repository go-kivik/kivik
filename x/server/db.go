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
	"net/http"

	"github.com/go-chi/chi/v5"
	"gitlab.com/flimzy/httpe"
)

func (s *Server) db() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		db := chi.URLParam(r, "db")
		stats, err := s.client.DB(db).Stats(r.Context())
		if err != nil {
			return err
		}
		return serveJSON(w, http.StatusOK, stats)
	})
}

func (s *Server) dbExists() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		db := chi.URLParam(r, "db")
		exists, err := s.client.DBExists(r.Context(), db, options(r))
		if err != nil {
			return err
		}
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			return nil
		}
		w.WriteHeader(http.StatusOK)
		return nil
	})
}

func (s *Server) createDB() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		db := chi.URLParam(r, "db")
		if err := s.client.CreateDB(r.Context(), db, options(r)); err != nil {
			return err
		}
		return serveJSON(w, http.StatusCreated, map[string]interface{}{
			"ok": true,
		})
	})
}

func (s *Server) deleteDB() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		db := chi.URLParam(r, "db")
		if err := s.client.DestroyDB(r.Context(), db, options(r)); err != nil {
			return err
		}
		return serveJSON(w, http.StatusOK, map[string]interface{}{
			"ok": true,
		})
	})
}
