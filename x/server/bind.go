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
	"encoding/json"
	"mime"
	"net/http"
)

// bind binds the request to v if it is of type application/json or
// application/x-www-form-urlencoded.
func (s *Server) bind(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	switch ct {
	case "application/json":

		return json.NewDecoder(r.Body).Decode(v)
	case "application/x-www-form-urlencoded":
		if err := r.ParseForm(); err != nil {
			return err
		}
		return s.formDecoder.Decode(r.Form, v)
	default:
		return &couchError{status: http.StatusUnsupportedMediaType, Err: "bad_content_type", Reason: "Content-Type must be 'application/x-www-form-urlencoded' or 'application/json'"}
	}
}

// bindJSON works like bind, but for endpoints that require application/json.
func (s *Server) bindJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	switch ct {
	case "application/json":
		return json.NewDecoder(r.Body).Decode(v)
	default:
		return &couchError{status: http.StatusUnsupportedMediaType, Err: "bad_content_type", Reason: "Content-Type must be 'application/json'"}
	}
}
