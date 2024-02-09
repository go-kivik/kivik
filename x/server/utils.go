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

	"gitlab.com/flimzy/httpe"

	"github.com/go-kivik/kivik/v4"
)

func (s *Server) root() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, _ *http.Request) error {
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

func (s *Server) up() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, _ *http.Request) error {
		return serveJSON(w, http.StatusOK, map[string]interface{}{
			"status": "ok",
		})
	})
}

// activeTasks returns a list of running tasks. For now it always returns an
// empty list, as this server doesn't support running asynchronous tasks. But it
// may be expanded in the future.
func (s *Server) activeTasks() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, _ *http.Request) error {
		return serveJSON(w, http.StatusOK, []interface{}{})
	})
}
