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
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gitlab.com/flimzy/httpe"
)

func (s *Server) postDoc() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		db := chi.URLParam(r, "db")
		var doc any
		if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
			return err
		}
		id, rev, err := s.client.DB(db).CreateDoc(r.Context(), doc, options(r))
		if err != nil {
			return err
		}
		return serveJSON(w, http.StatusCreated, map[string]any{
			"id":  id,
			"rev": rev,
			"ok":  true,
		})
	})
}

func (s *Server) doc() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		db := chi.URLParam(r, "db")
		id := chi.URLParam(r, "docid")
		var doc any
		err := s.client.DB(db).Get(r.Context(), id, options(r)).ScanDoc(&doc)
		if err != nil {
			return err
		}
		return serveJSON(w, http.StatusOK, doc)
	})
}
