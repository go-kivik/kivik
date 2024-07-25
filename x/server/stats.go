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

	"gitlab.com/flimzy/httpe"
)

func (s *Server) allDBsStats() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		stats, err := s.client.AllDBsStats(r.Context(), options(r))
		if err != nil {
			return err
		}
		return serveJSON(w, http.StatusOK, stats)
	})
}

func (s *Server) dbsStats() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		var req struct {
			Keys []string `json:"keys"`
		}
		if err := s.bind(r, &req); err != nil {
			return err
		}
		stats, err := s.client.DBsStats(r.Context(), req.Keys)
		if err != nil {
			return err
		}
		return serveJSON(w, http.StatusOK, stats)
	})
}
