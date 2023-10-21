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

// Package server provides a CouchDB server via HTTP.
package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gitlab.com/flimzy/httpe"

	"github.com/go-kivik/kivik/v4"
)

// Server is a server instance.
type Server struct {
	mux *chi.Mux
}

// New instantiates a new server instance.
func New() *Server {
	s := &Server{}
	s.mux = chi.NewMux()
	s.mux.Get("/", httpe.ToHandler(s.root()).ServeHTTP)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func serveJSON(w http.ResponseWriter, status int, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, err = io.Copy(w, bytes.NewReader(body))
	return err
}

func (s *Server) root() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		return serveJSON(w, http.StatusOK, map[string]interface{}{
			"couchdb": "Welcome",
			"vendor": map[string]string{
				"name":    "Kivik",
				"version": kivik.Version,
			},
			"version": kivik.Version,
		})
	})
}
