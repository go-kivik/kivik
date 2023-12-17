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
	"net/http"

	"gitlab.com/flimzy/httpe"

	"github.com/go-kivik/kivik/v4"
)

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

func (s *Server) up() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		return serveJSON(w, http.StatusOK, map[string]interface{}{
			"status": "ok",
		})
	})
}
