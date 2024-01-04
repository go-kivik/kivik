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
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"gitlab.com/flimzy/httpe"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal"
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

const (
	defaultHeartbeat = heartbeat(60 * time.Second)
	defaultTimeout   = 60000 // milliseconds

	feedTypeNormal     = "normal"
	feedTypeLongpoll   = "longpoll"
	feedTypeContinuous = "continuous"
)

type heartbeat time.Duration

func (h *heartbeat) UnmarshalText(text []byte) error {
	var value heartbeat
	if string(text) == "true" {
		value = defaultHeartbeat
	} else {
		ms, err := strconv.Atoi(string(text))
		if err != nil {
			return err
		}
		value = heartbeat(ms) * heartbeat(time.Millisecond)
	}
	*h = value
	return nil
}

func (s *Server) dbUpdates() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		switch feed := r.URL.Query().Get("feed"); feed {
		case "", feedTypeNormal:
			return s.serveNormalDBUpdates(w, r)
		case feedTypeContinuous, feedTypeLongpoll:
			return s.serveContinuousDBUpdates(w, r)
		default:
			return &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("kivik: feed type %q not supported", feed)}
		}
	})
}

func (s *Server) serveNormalDBUpdates(w http.ResponseWriter, r *http.Request) error {
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

	enc := json.NewEncoder(w)

	var update driver.DBUpdate
	for updates.Next() {
		if update.DBName != "" { // Easy way to tell if this is the first result
			if _, err := w.Write([]byte(",")); err != nil {
				return err
			}
		}
		update.DBName = updates.DBName()
		update.Type = updates.Type()
		update.Seq = updates.Seq()
		if err := enc.Encode(&update); err != nil {
			return err
		}
	}
	if err := updates.Err(); err != nil {
		return err
	}

	lastSeq, err := updates.LastSeq()
	if err != nil {
		return err
	}

	if _, err := w.Write([]byte(`],"last_seq":"` + lastSeq + "\"}")); err != nil {
		return err
	}

	return nil
}

func (s *Server) serveContinuousDBUpdates(w http.ResponseWriter, r *http.Request) error {
	req := struct {
		Heartbeat heartbeat `form:"heartbeat"`
	}{
		Heartbeat: defaultHeartbeat,
	}
	if err := s.bind(r, &req); err != nil {
		return err
	}

	ticker := time.NewTicker(time.Duration(req.Heartbeat))
	updates := s.client.DBUpdates(r.Context(), options(r))
	if err := updates.Err(); err != nil {
		return err
	}

	defer updates.Close()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	nextUpdate := make(chan *driver.DBUpdate)
	go func() {
		for updates.Next() {
			nextUpdate <- &driver.DBUpdate{
				DBName: updates.DBName(),
				Type:   updates.Type(),
				Seq:    updates.Seq(),
			}
		}
		close(nextUpdate)
	}()

	enc := json.NewEncoder(w)

loop:
	for {
		select {
		case <-ticker.C:
			if _, err := w.Write([]byte("\n")); err != nil {
				return err
			}
		case update, ok := <-nextUpdate:
			if !ok {
				break loop
			}
			ticker.Reset(time.Duration(req.Heartbeat))
			if err := enc.Encode(update); err != nil {
				return err
			}
			if _, err := w.Write([]byte("\n")); err != nil {
				return err
			}
		}
	}

	return updates.Err()
}
