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

func e(h httpe.HandlerWithError) httpe.HandlerFunc {
	return httpe.ToHandler(h).ServeHTTP
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

	auth := mux.With(
		httpe.ToMiddleware(s.authMiddleware),
	)

	admin := auth.With(
		httpe.ToMiddleware(adminRequired),
	)
	auth.Get("/", e(s.root()))
	admin.Get("/_active_tasks", e(s.activeTasks()))
	admin.Get("/_all_dbs", e(s.allDBs()))
	auth.Get("/_dbs_info", e(s.allDBsStats()))
	auth.Post("/_dbs_info", e(s.dbsStats()))
	auth.Get("/_cluster_setup", e(s.notImplemented()))
	auth.Post("/_cluster_setup", e(s.notImplemented()))
	auth.Post("/_db_updates", e(s.notImplemented()))
	auth.Get("/_membership", e(s.notImplemented()))
	auth.Post("/_replicate", e(s.notImplemented()))
	auth.Get("/_scheduler/jobs", e(s.notImplemented()))
	auth.Get("/_scheduler/docs", e(s.notImplemented()))
	auth.Get("/_scheduler/docs/{replicator_db}", e(s.notImplemented()))
	auth.Get("/_scheduler/docs/{replicator_db}/{doc_id}", e(s.notImplemented()))
	auth.Get("/_node/{node-name}", e(s.notImplemented()))
	auth.Get("/_node/{node-name}/_stats", e(s.notImplemented()))
	auth.Get("/_node/{node-name}/_prometheus", e(s.notImplemented()))
	auth.Get("/_node/{node-name}/_system", e(s.notImplemented()))
	admin.Post("/_node/{node-name}/_restart", e(s.notImplemented()))
	auth.Get("/_node/{node-name}/_versions", e(s.notImplemented()))
	auth.Post("/_search_analyze", e(s.notImplemented()))
	auth.Get("/_utils", e(s.notImplemented()))
	auth.Get("/_utils/", e(s.notImplemented()))
	mux.Get("/_up", e(s.up()))
	mux.Get("/_uuids", e(s.uuids()))
	mux.Get("/favicon.ico", e(s.notImplemented()))
	auth.Get("/_reshard", e(s.notImplemented()))
	auth.Get("/_reshard/state", e(s.notImplemented()))
	auth.Put("/_reshard/state", e(s.notImplemented()))
	auth.Get("/_reshard/jobs", e(s.notImplemented()))
	auth.Get("/_reshard/jobs/{jobid}", e(s.notImplemented()))
	auth.Post("/_reshard/jobs", e(s.notImplemented()))
	auth.Delete("/_reshard/jobs/{jobid}", e(s.notImplemented()))
	auth.Get("/_reshard/jobs/{jobid}/state", e(s.notImplemented()))
	auth.Put("/_reshard/jobs/{jobid}/state", e(s.notImplemented()))

	// Config
	admin.Get("/_node/{node-name}/_config", e(s.allConfig()))
	admin.Get("/_node/{node-name}/_config/{section}", e(s.configSection()))
	admin.Get("/_node/{node-name}/_config/{section}/{key}", e(s.configKey()))
	admin.Put("/_node/{node-name}/_config/{section}/{key}", e(s.setConfigKey()))
	admin.Delete("/_node/{node-name}/_config/{section}/{key}", e(s.deleteConfigKey()))
	admin.Post("/_node/{node-name}/_config/_reload", e(s.reloadConfig()))

	// Databases
	auth.Route("/{db}", func(db chi.Router) {
		admin := db.With(
			httpe.ToMiddleware(adminRequired),
		)
		member := db.With(
			httpe.ToMiddleware(s.dbMembershipRequired),
		)
		dbAdmin := member.With(
			httpe.ToMiddleware(s.dbAdminRequired),
		)

		member.Head("/", e(s.dbExists()))
		member.Get("/", e(s.db()))
		admin.Put("/", e(s.createDB()))
		admin.Delete("/", e(s.deleteDB()))
		member.Get("/_all_docs", e(s.notImplemented()))
		member.Post("/_all_docs/queries", e(s.notImplemented()))
		member.Post("/_all_docs", e(s.notImplemented()))
		member.Get("/_design_docs", e(s.notImplemented()))
		member.Post("/_design_docs", e(s.notImplemented()))
		member.Post("/_design_docs/queries", e(s.notImplemented()))
		member.Get("/_local_docs", e(s.notImplemented()))
		member.Post("/_local_docs", e(s.notImplemented()))
		member.Post("/_local_docs/queries", e(s.notImplemented()))
		member.Post("/_bulk_get", e(s.notImplemented()))
		member.Post("/_bulk_docs", e(s.notImplemented()))
		member.Post("/_find", e(s.notImplemented()))
		member.Post("/_index", e(s.notImplemented()))
		member.Get("/_index", e(s.notImplemented()))
		member.Delete("/_index/{designdoc}/json/{name}", e(s.notImplemented()))
		member.Post("/_explain", e(s.notImplemented()))
		member.Get("/_shards", e(s.notImplemented()))
		member.Get("/_shards/{docid}", e(s.notImplemented()))
		member.Get("/_sync_shards", e(s.notImplemented()))
		member.Get("/_changes", e(s.notImplemented()))
		member.Post("/_changes", e(s.notImplemented()))
		admin.Post("/_compact", e(s.notImplemented()))
		admin.Post("/_compact/{ddoc}", e(s.notImplemented()))
		member.Post("/_ensure_full_commit", e(s.notImplemented()))
		member.Post("/_view_cleanup", e(s.notImplemented()))
		member.Get("/_security", e(s.getSecurity()))
		dbAdmin.Put("/_security", e(s.putSecurity()))
		member.Post("/_purge", e(s.notImplemented()))
		member.Get("/_purged_infos_limit", e(s.notImplemented()))
		member.Put("/_purged_infos_limit", e(s.notImplemented()))
		member.Post("/_missing_revs", e(s.notImplemented()))
		member.Post("/_revs_diff", e(s.notImplemented()))
		member.Get("/_revs_limit", e(s.notImplemented()))
		dbAdmin.Put("/_revs_limit", e(s.notImplemented()))

		// Documents
		member.Post("/", e(s.postDoc()))
		member.Get("/{docid}", e(s.doc()))
		member.Put("/{docid}", e(s.notImplemented()))
		member.Delete("/{docid}", e(s.notImplemented()))
		member.Method("COPY", "/{db}/{docid}", httpe.ToHandler(s.notImplemented()))
		member.Delete("/{docid}", e(s.notImplemented()))
		member.Get("/{docid}/{attname}", e(s.notImplemented()))
		member.Get("/{docid}/{attname}", e(s.notImplemented()))
		member.Delete("/{docid}/{attname}", e(s.notImplemented()))

		// Design docs
		member.Get("/_design/{ddoc}", e(s.notImplemented()))
		dbAdmin.Put("/_design/{ddoc}", e(s.notImplemented()))
		dbAdmin.Delete("/_design/{ddoc}", e(s.notImplemented()))
		dbAdmin.Method("COPY", "/{db}/_design/{ddoc}", httpe.ToHandler(s.notImplemented()))
		member.Get("/_design/{ddoc}/{attname}", e(s.notImplemented()))
		dbAdmin.Put("/_design/{ddoc}/{attname}", e(s.notImplemented()))
		dbAdmin.Delete("/_design/{ddoc}/{attname}", e(s.notImplemented()))
		member.Get("/_design/{ddoc}/_view/{view}", e(s.notImplemented()))
		member.Get("/_design/{ddoc}/_info", e(s.notImplemented()))
		member.Post("/_design/{ddoc}/_view/{view}", e(s.notImplemented()))
		member.Post("/_design/{ddoc}/_view/{view}/queries", e(s.notImplemented()))
		member.Get("/_design/{ddoc}/_search/{index}", e(s.notImplemented()))
		member.Get("/_design/{ddoc}/_search_info/{index}", e(s.notImplemented()))
		member.Post("/_design/{ddoc}/_update/{func}", e(s.notImplemented()))
		member.Post("/_design/{ddoc}/_update/{func}/{docid}", e(s.notImplemented()))
		member.Get("/_design/{ddoc}/_rewrite/{path}", e(s.notImplemented()))
		member.Put("/_design/{ddoc}/_rewrite/{path}", e(s.notImplemented()))
		member.Post("/_design/{ddoc}/_rewrite/{path}", e(s.notImplemented()))
		member.Delete("/_design/{ddoc}/_rewrite/{path}", e(s.notImplemented()))

		member.Get("/_partition/{partition}", e(s.notImplemented()))
		member.Get("/_partition/{partition}/_all_docs", e(s.notImplemented()))
		member.Get("/_partition/{partition}/_design/{ddoc}/_view/{view}", e(s.notImplemented()))
		member.Post("/_partition/{partition_id}/_find", e(s.notImplemented()))
		member.Get("/_partition/{partition_id}/_explain", e(s.notImplemented()))
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
