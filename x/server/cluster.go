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

func (s *Server) clusterStatus() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		status, err := s.client.ClusterStatus(r.Context(), options(r))
		if err != nil {
			return err
		}
		return serveJSON(w, http.StatusOK, map[string]string{
			"state": status,
		})
	})
}

func (s *Server) clusterSetup() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		var req struct {
			Action string `json:"action"`
		}
		if err := s.bindJSON(r, &req); err != nil {
			return err
		}
		if err := s.client.ClusterSetup(r.Context(), req.Action); err != nil {
			return err
		}
		return serveJSON(w, http.StatusOK, map[string]bool{
			"ok": true,
		})
	})
}
