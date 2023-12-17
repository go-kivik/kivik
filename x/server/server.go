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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/monoculum/formam/v3"
	"gitlab.com/flimzy/httpe"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/x/server/auth"
	"github.com/go-kivik/kivik/v4/x/server/config"
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
	config      config.Config

	// This is set the first time a sequential UUID is generated, and is used
	// for all subsequent sequential UUIDs.
	sequentialMU              sync.Mutex
	sequentialUUIDPrefix      string
	sequentialUUIDMonotonicID int32
}

// New instantiates a new server instance.
func New(client *kivik.Client, options ...Option) *Server {
	s := &Server{
		mux:    chi.NewMux(),
		client: client,
		formDecoder: formam.NewDecoder(&formam.DecoderOptions{
			TagName: "form",
		}),
		config: config.Default(),
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

	auth := mux.With(
		httpe.ToMiddleware(s.authMiddleware),
	)

	admin := auth.With(
		httpe.ToMiddleware(adminRequired),
	)
	admin.Get("/_active_tasks", httpe.ToHandler(s.activeTasks()).ServeHTTP)
	admin.Get("/_all_dbs", httpe.ToHandler(s.allDBs()).ServeHTTP)
	auth.Get("/_dbs_info", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/_dbs_info", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/_cluster_setup", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/_cluster_setup", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/_db_updates", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/_membership", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/_replicate", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/_scheduler/jobs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/_scheduler/docs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/_scheduler/docs/{replicator_db}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/_scheduler/docs/{replicator_db}/{doc_id}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/_node/{node-name}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/_node/{node-name}/_stats", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/_node/{node-name}/_prometheus", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/_node/{node-name}/_system", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/_node/{node-name}/_restart", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/_node/{node-name}/_versions", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/_search_analyze", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/_utils", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/_utils/", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	mux.Get("/_up", httpe.ToHandler(s.up()).ServeHTTP)
	mux.Get("/_uuids", httpe.ToHandler(s.uuids()).ServeHTTP)
	mux.Get("/favicon.ico", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/_reshard", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/_reshard/state", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Put("/_reshard/state", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/_reshard/jobs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/_reshard/jobs/{jobid}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/_reshard/jobs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Delete("/_reshard/jobs/{jobid}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/_reshard/jobs/{jobid}/state", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Put("/_reshard/jobs/{jobid}/state", httpe.ToHandler(s.notImplemented()).ServeHTTP)

	// Config
	admin.Get("/_node/{node-name}/_config", httpe.ToHandler(s.allConfig()).ServeHTTP)
	admin.Get("/_node/{node-name}/_config/{section}", httpe.ToHandler(s.configSection()).ServeHTTP)
	admin.Get("/_node/{node-name}/_config/{section}/{key}", httpe.ToHandler(s.configKey()).ServeHTTP)
	admin.Put("/_node/{node-name}/_config/{section}/{key}", httpe.ToHandler(s.setConfigKey()).ServeHTTP)
	admin.Delete("/_node/{node-name}/_config/{section}/{key}", httpe.ToHandler(s.deleteConfigKey()).ServeHTTP)
	admin.Post("/_node/{node-name}/_config/_reload", httpe.ToHandler(s.reloadConfig()).ServeHTTP)

	// Databases
	auth.Head("/{db}", httpe.ToHandler(s.dbExists()).ServeHTTP)
	auth.Get("/{db}", httpe.ToHandler(s.db()).ServeHTTP)
	auth.Put("/{db}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Delete("/{db}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/{db}/_all_docs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_all_docs/queries", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_all_docs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/{db}/_design_docs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_design_docs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_design_docs/queries", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/{db}/_local_docs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_local_docs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_local_docs/queries", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_bulk_get", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_bulk_docs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_find", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_index", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/{db}/_index", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Delete("/{db}/_index/{designdoc}/json/{name}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_explain", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/{db}/_shards", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/{db}/_shards/{docid}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/{db}/_sync_shards", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/{db}/_changes", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_changes", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_compact", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_compact/{ddoc}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_ensure_full_commit", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_view_cleanup", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/{db}/_security", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Put("/{db}/_security", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_purge", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/{db}/_purged_infos_limit", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Put("/{db}/_purged_infos_limit", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_missing_revs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_revs_diff", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/{db}/_revs_limit", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Put("/{db}/_revs_limit", httpe.ToHandler(s.notImplemented()).ServeHTTP)

	// Documents
	auth.Get("/{db}/{docid}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Put("/{db}/{docid}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Delete("/{db}/{docid}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Method("COPY", "/{db}/{docid}", httpe.ToHandler(s.notImplemented()))
	auth.Delete("/{db}/{docid}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/{db}/{docid}/{attname}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/{db}/{docid}/{attname}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Delete("/{db}/{docid}/{attname}", httpe.ToHandler(s.notImplemented()).ServeHTTP)

	// Design docs
	auth.Get("/{db}/_design/{ddoc}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Put("/{db}/_design/{ddoc}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Delete("/{db}/_design/{ddoc}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Method("COPY", "/{db}/_design/{ddoc}", httpe.ToHandler(s.notImplemented()))
	auth.Get("/{db}/_design/{ddoc}/{attname}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Put("/{db}/_design/{ddoc}/{attname}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Delete("/{db}/_design/{ddoc}/{attname}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/{db}/_design/{ddoc}/_view/{view}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/{db}/_design/{ddoc}/_info", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_design/{ddoc}/_view/{view}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_design/{ddoc}/_view/{view}/queries", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/{db}/_design/{ddoc}/_search/{index}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/{db}/_design/{ddoc}/_search_info/{index}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_design/{ddoc}/_update/{func}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_design/{ddoc}/_update/{func}/{docid}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/{db}/_design/{ddoc}/_rewrite/{path}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Put("/{db}/_design/{ddoc}/_rewrite/{path}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_design/{ddoc}/_rewrite/{path}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Delete("/{db}/_design/{ddoc}/_rewrite/{path}", httpe.ToHandler(s.notImplemented()).ServeHTTP)

	auth.Get("/{db}/_partition/{partition}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/{db}/_partition/{partition}/_all_docs", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/{db}/_partition/{partition}/_design/{ddoc}/_view/{view}", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Post("/{db}/_partition/{partition_id}/_find", httpe.ToHandler(s.notImplemented()).ServeHTTP)
	auth.Get("/{db}/_partition/{partition_id}/_explain", httpe.ToHandler(s.notImplemented()).ServeHTTP)
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

func (s *Server) conf(ctx context.Context, section, key string, target interface{}) error {
	value, err := s.config.Key(ctx, section, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(value), target)
}
