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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"gitlab.com/flimzy/httpe"

	"github.com/go-kivik/kivik/v4"
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

// whichView returns `_all_docs`, `_local_docs`, of `_design_docs`, and whether
// the path ends with /queries.
func whichView(r *http.Request) (string, bool) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if parts[len(parts)-1] == "queries" {
		return parts[len(parts)-2], true
	}
	return parts[len(parts)-1], false
}

func (s *Server) query() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		view, isQueries := whichView(r)
		req := map[string]interface{}{}
		if isQueries {
			var jsonReq struct {
				Queries []map[string]interface{} `json:"queries"`
			}
			if err := s.bind(r, &jsonReq); err != nil {
				return err
			}
			req["queries"] = jsonReq.Queries
		} else {
			if err := s.bind(r, &req); err != nil {
				return err
			}
		}
		db := chi.URLParam(r, "db")
		var viewFunc func(context.Context, ...kivik.Option) *kivik.ResultSet
		switch view {
		case "_all_docs":
			viewFunc = s.client.DB(db).AllDocs
		case "_local_docs":
			viewFunc = s.client.DB(db).LocalDocs
		default:
			return &internal.Error{Status: http.StatusNotFound, Message: fmt.Sprintf("kivik: view %q not found", view)}
		}
		rows := viewFunc(r.Context(), options(r))
		defer rows.Close()

		if err := rows.Err(); err != nil {
			return err
		}

		if _, err := fmt.Fprint(w, `{"rows":[`); err != nil {
			return err
		}

		var row struct {
			ID    string          `json:"id"`
			Key   json.RawMessage `json:"key"`
			Value json.RawMessage `json:"value"`
		}
		var err error
		enc := json.NewEncoder(w)
		for rows.Next() {
			if row.ID != "" { // Easy way to tell if this is the first row
				if _, err = w.Write([]byte(",")); err != nil {
					return err
				}
			}
			row.ID, err = rows.ID()
			if err != nil {
				return err
			}
			if err := rows.ScanKey(&row.Key); err != nil {
				return err
			}
			if err := rows.ScanValue(&row.Value); err != nil {
				return err
			}
			if err := enc.Encode(&row); err != nil {
				return err
			}
		}
		meta, err := rows.Metadata()
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, `],"offset":%d,"total_rows":%d}`, meta.Offset, meta.TotalRows)
		return err
	})
}
