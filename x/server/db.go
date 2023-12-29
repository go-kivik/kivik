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
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"gitlab.com/flimzy/httpe"

	"github.com/go-kivik/kivik/v4/driver"
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

const defaultHeartbeat = 60 * time.Second

type heartbeat time.Duration

func (h *heartbeat) UnmarshalText(text []byte) error {
	var value time.Duration
	if string(text) == "true" {
		value = defaultHeartbeat
	} else {
		ms, err := strconv.Atoi(string(text))
		if err != nil {
			return err
		}
		value = time.Duration(ms) * time.Millisecond
	}
	*h = heartbeat(value)
	return nil
}

func (s *Server) dbUpdates() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		var req struct {
			Heartbeat heartbeat `form:"heartbeat"`
		}
		if err := s.bind(r, &req); err != nil {
			return err
		}
		updates := s.client.DBUpdates(r.Context(), options(r))

		if err := updates.Err(); err != nil {
			return err
		}

		defer updates.Close()

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte(`{"results":[`)); err != nil {
			return err
		}

		var update driver.DBUpdate
		var lastSeq string
		for updates.Next() {
			if lastSeq != "" {
				if _, err := w.Write([]byte(",")); err != nil {
					return err
				}
			}
			update.DBName = updates.DBName()
			update.Type = updates.Type()
			update.Seq = updates.Seq()
			lastSeq = update.Seq
			if err := json.NewEncoder(w).Encode(update); err != nil {
				return err
			}
		}
		if _, err := w.Write([]byte(`],"last_seq":"` + lastSeq + "\"}")); err != nil {
			return err
		}

		return updates.Err()
	})
}
