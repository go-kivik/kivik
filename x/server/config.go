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
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gitlab.com/flimzy/httpe"

	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/x/server/config"
)

const nodeLocal = "_local"

func (s *Server) allConfig() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		if node := chi.URLParam(r, "node-name"); node != nodeLocal {
			return &internal.Error{Status: http.StatusNotFound, Message: fmt.Sprintf("no such node: %s", node)}
		}
		conf, err := s.config.All(r.Context())
		if err != nil {
			return err
		}
		return serveJSON(w, http.StatusOK, conf)
	})
}

func (s *Server) configSection() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		if node := chi.URLParam(r, "node-name"); node != nodeLocal {
			return &internal.Error{Status: http.StatusNotFound, Message: fmt.Sprintf("no such node: %s", node)}
		}
		section, err := s.config.Section(r.Context(), chi.URLParam(r, "section"))
		if err != nil {
			return err
		}
		return serveJSON(w, http.StatusOK, section)
	})
}

func (s *Server) configKey() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		if node := chi.URLParam(r, "node-name"); node != nodeLocal {
			return &internal.Error{Status: http.StatusNotFound, Message: fmt.Sprintf("no such node: %s", node)}
		}
		key, err := s.config.Key(r.Context(), chi.URLParam(r, "section"), chi.URLParam(r, "key"))
		if err != nil {
			return err
		}
		return serveJSON(w, http.StatusOK, key)
	})
}

func (s *Server) reloadConfig() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		if node := chi.URLParam(r, "node-name"); node != nodeLocal {
			return &internal.Error{Status: http.StatusNotFound, Message: fmt.Sprintf("no such node: %s", node)}
		}
		if err := s.config.Reload(r.Context()); err != nil {
			return err
		}
		return serveJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})
}

func (s *Server) setConfigKey() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		if node := chi.URLParam(r, "node-name"); node != nodeLocal {
			return &internal.Error{Status: http.StatusNotFound, Message: fmt.Sprintf("no such node: %s", node)}
		}
		confWriter, ok := s.config.(config.Writer)
		if !ok {
			return &internal.Error{Status: http.StatusMethodNotAllowed, Message: "configuration is read-only"}
		}
		var newValue string
		if err := json.NewDecoder(r.Body).Decode(&newValue); err != nil {
			return &internal.Error{Status: http.StatusBadRequest, Err: err}
		}
		oldValue, err := confWriter.SetKey(r.Context(), chi.URLParam(r, "section"), chi.URLParam(r, "key"), newValue)
		if err != nil {
			return err
		}
		return serveJSON(w, http.StatusOK, oldValue)
	})
}

func (s *Server) deleteConfigKey() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		if node := chi.URLParam(r, "node-name"); node != nodeLocal {
			return &internal.Error{Status: http.StatusNotFound, Message: fmt.Sprintf("no such node: %s", node)}
		}
		confWriter, ok := s.config.(config.Writer)
		if !ok {
			return &internal.Error{Status: http.StatusMethodNotAllowed, Message: "configuration is read-only"}
		}
		oldValue, err := confWriter.Delete(r.Context(), chi.URLParam(r, "section"), chi.URLParam(r, "key"))
		if err != nil {
			return err
		}
		return serveJSON(w, http.StatusOK, oldValue)
	})
}
