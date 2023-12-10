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
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/monoculum/formam/v3"
	"gitlab.com/flimzy/httpe"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/x/server/auth"
)

func init() {
	chi.RegisterMethod("COPY")
}

// Server is a server instance.
type Server struct {
	mux         *chi.Mux
	client      *kivik.Client
	formDecoder *formam.Decoder
	userStores  userStores
	authFuncs   []auth.AuthenticateFunc
}

// New instantiates a new server instance.
func New(client *kivik.Client, options ...Option) *Server {
	s := &Server{
		mux:    chi.NewMux(),
		client: client,
		formDecoder: formam.NewDecoder(&formam.DecoderOptions{
			TagName: "form",
		}),
	}
	for _, option := range options {
		option.apply(s)
	}
	s.routes(s.mux)
	return s
}

func (s *Server) routes(mux *chi.Mux) {
	mux.Use(
		GetHead,
		httpe.ToMiddleware(s.handleErrors),
	)
	mux.Get("/", httpe.ToHandler(s.root()).ServeHTTP)
	mux.Route("/", func(mux chi.Router) {
		mux.Use(
			httpe.ToMiddleware(s.authMiddleware),
		)
		mux.Get("/_active_tasks", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_all_dbs", httpe.ToHandler(s.allDBs()).ServeHTTP)
		mux.Get("/_dbs_info", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/_dbs_info", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_cluster_setup", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/_cluster_setup", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/_db_updates", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_membership", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/_replicate", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_scheduler/jobs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_scheduler/docs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_scheduler/docs/{replicator_db}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_scheduler/docs/{replicator_db}/{doc_id}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_node/{node-name}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_node/{node-name}/_stats", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_node/{node-name}/_prometheus", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_node/{node-name}/_system", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/_node/{node-name}/_restart", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_node/{node-name}/_versions", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/_search_analyze", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_utils", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_utils/", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_up", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_uuids", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/favicon.ico", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_reshard", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_reshard/state", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Put("/_reshard/state", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_reshard/jobs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_reshard/jobs/{jobid}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/_reshard/jobs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Delete("/_reshard/jobs/{jobid}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_reshard/jobs/{jobid}/state", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Put("/_reshard/jobs/{jobid}/state", httpe.ToHandler(s.notImplemented()).ServeHTTP)

		// Auth
		// mux.Post("/_session", httpe.ToHandler(s.startSession()).ServeHTTP)
		// mux.Get("/_session", httpe.ToHandler(s.notImplemented()).ServeHTTP)

		// Config
		mux.Get("/_node/{node-name}/_config", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_node/{node-name}/_config/{section}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/_node/{node-name}/_config/{section}/{key}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Put("/_node/{node-name}/_config/{section}/{key}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Delete("/_node/{node-name}/_config/{section}/{key}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/_node/{node-name}/_config/_reload", httpe.ToHandler(s.notImplemented()).ServeHTTP)

		// Databases
		mux.Head("/{db}", httpe.ToHandler(s.dbExists()).ServeHTTP)
		mux.Get("/{db}", httpe.ToHandler(s.db()).ServeHTTP)
		mux.Put("/{db}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Delete("/{db}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/{db}/_all_docs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_all_docs/queries", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_all_docs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/{db}/_design_docs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_design_docs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_design_docs/queries", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/{db}/_local_docs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_local_docs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_local_docs/queries", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_bulk_get", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_bulk_docs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_find", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_index", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/{db}/_index", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Delete("/{db}/_index/{designdoc}/json/{name}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_explain", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/{db}/_shards", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/{db}/_shards/{docid}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/{db}/_sync_shards", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/{db}/_changes", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_changes", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_compact", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_compact/{ddoc}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_ensure_full_commit", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_view_cleanup", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/{db}/_security", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Put("/{db}/_security", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_purge", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/{db}/_purged_infos_limit", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Put("/{db}/_purged_infos_limit", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_missing_revs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_revs_diff", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/{db}/_revs_limit", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Put("/{db}/_revs_limit", httpe.ToHandler(s.notImplemented()).ServeHTTP)

		// Documents
		mux.Get("/{db}/{docid}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Put("/{db}/{docid}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Delete("/{db}/{docid}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Method("COPY", "/{db}/{docid}", httpe.ToHandler(s.notImplemented()))
		mux.Delete("/{db}/{docid}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/{db}/{docid}/{attname}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/{db}/{docid}/{attname}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Delete("/{db}/{docid}/{attname}", httpe.ToHandler(s.notImplemented()).ServeHTTP)

		// Design docs
		mux.Get("/{db}/_design/{ddoc}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Put("/{db}/_design/{ddoc}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Delete("/{db}/_design/{ddoc}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Method("COPY", "/{db}/_design/{ddoc}", httpe.ToHandler(s.notImplemented()))
		mux.Get("/{db}/_design/{ddoc}/{attname}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Put("/{db}/_design/{ddoc}/{attname}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Delete("/{db}/_design/{ddoc}/{attname}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/{db}/_design/{ddoc}/_view/{view}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/{db}/_design/{ddoc}/_info", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_design/{ddoc}/_view/{view}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_design/{ddoc}/_view/{view}/queries", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/{db}/_design/{ddoc}/_search/{index}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/{db}/_design/{ddoc}/_search_info/{index}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_design/{ddoc}/_update/{func}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_design/{ddoc}/_update/{func}/{docid}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/{db}/_design/{ddoc}/_rewrite/{path}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Put("/{db}/_design/{ddoc}/_rewrite/{path}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_design/{ddoc}/_rewrite/{path}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Delete("/{db}/_design/{ddoc}/_rewrite/{path}", httpe.ToHandler(s.notImplemented()).ServeHTTP)

		mux.Get("/{db}/_partition/{partition}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/{db}/_partition/{partition}/_all_docs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/{db}/_partition/{partition}/_design/{ddoc}/_view/{view}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Post("/{db}/_partition/{partition_id}/_find", httpe.ToHandler(s.notImplemented()).ServeHTTP)
		mux.Get("/{db}/_partition/{partition_id}/_explain", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	})
}

func (s *Server) handleErrors(next httpe.HandlerWithError) httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		if err := next.ServeHTTPWithError(w, r); err != nil {
			status := kivik.HTTPStatus(err)
			ce := &couchError{}
			if !errors.As(err, &ce) {
				ce.Err = strings.ReplaceAll(strings.ToLower(http.StatusText(status)), " ", "_")
				ce.Reason = err.Error()
			}
			return serveJSON(w, status, ce)
		}
		return nil
	})
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

func (s *Server) notImplemented() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		return errNotImplimented
	})
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

func options(r *http.Request) kivik.Option {
	query := r.URL.Query()
	params := make(map[string]interface{}, len(query))
	for k := range query {
		params[k] = query.Get(k)
	}
	return kivik.Params(params)
}

func (s *Server) allDBs() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		dbs, err := s.client.AllDBs(r.Context(), options(r))
		if err != nil {
			fmt.Println(err)
			return err
		}
		return serveJSON(w, http.StatusOK, dbs)
	})
}

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
		exists, err := s.client.DBExists(r.Context(), db)
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

func (s *Server) bind(r *http.Request, v interface{}) error {
	ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	switch ct {
	case "application/json":
		defer r.Body.Close()
		return json.NewDecoder(r.Body).Decode(v)
	case "application/x-www-form-urlencoded":
		defer r.Body.Close()
		if err := r.ParseForm(); err != nil {
			return err
		}
		return s.formDecoder.Decode(r.Form, v)
	default:
		return &couchError{status: http.StatusUnsupportedMediaType, Err: "bad_content_type", Reason: "Content-Type must be 'application/x-www-form-urlencoded' or 'application/json'"}
	}
}
