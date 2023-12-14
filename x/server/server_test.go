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

package server

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	_ "github.com/go-kivik/kivik/v4/x/fsdb" // Filesystem driver
	"github.com/go-kivik/kivik/v4/x/server/auth"
	"github.com/go-kivik/kivik/v4/x/server/config"
)

const (
	userAdmin      = "admin"
	userReader     = "reader"
	userWriter     = "writer"
	userReplicator = "replicator"
	userDBUpdates  = "db_updates"
	userDesign     = "design"
	testPassword   = "abc123"
)

func testUserStore(t *testing.T) *auth.MemoryUserStore {
	t.Helper()
	us := auth.NewMemoryUserStore()
	if err := us.AddUser(userAdmin, testPassword, []string{auth.RoleAdmin}); err != nil {
		t.Fatal(err)
	}
	if err := us.AddUser(userReader, testPassword, []string{auth.RoleReader}); err != nil {
		t.Fatal(err)
	}
	if err := us.AddUser(userWriter, testPassword, []string{auth.RoleWriter}); err != nil {
		t.Fatal(err)
	}
	if err := us.AddUser(userReplicator, testPassword, []string{auth.RoleReplicator}); err != nil {
		t.Fatal(err)
	}
	if err := us.AddUser(userDBUpdates, testPassword, []string{auth.RoleDBUpdates}); err != nil {
		t.Fatal(err)
	}
	if err := us.AddUser(userDesign, testPassword, []string{auth.RoleDesign}); err != nil {
		t.Fatal(err)
	}
	return us
}

func basicAuth(user string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+testPassword))
}

func TestServer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		extraOptions []Option
		client       *kivik.Client
		method       string
		path         string
		headers      map[string]string
		authUser     string
		body         io.Reader
		wantStatus   int
		wantJSON     interface{}
	}{
		{
			name:       "root",
			method:     http.MethodGet,
			path:       "/",
			wantStatus: http.StatusOK,
			wantJSON: map[string]interface{}{
				"couchdb": "Welcome",
				"vendor": map[string]interface{}{
					"name":    "Kivik",
					"version": kivik.Version,
				},
				"version": kivik.Version,
			},
		},
		{
			name:       "active tasks",
			method:     http.MethodGet,
			path:       "/_active_tasks",
			headers:    map[string]string{"Authorization": basicAuth(userAdmin)},
			wantStatus: http.StatusNotImplemented,
			wantJSON: map[string]interface{}{
				"error":  "not_implemented",
				"reason": "Feature not implemented",
			},
		},
		{
			name:       "all dbs",
			method:     http.MethodGet,
			path:       "/_all_dbs",
			headers:    map[string]string{"Authorization": basicAuth(userAdmin)},
			wantStatus: http.StatusOK,
			wantJSON:   []string{"db1", "db2"},
		},
		{
			name:       "all dbs, cookie auth",
			method:     http.MethodGet,
			path:       "/_all_dbs",
			authUser:   userAdmin,
			wantStatus: http.StatusOK,
			wantJSON:   []string{"db1", "db2"},
		},
		{
			name:       "all dbs, non-admin",
			method:     http.MethodGet,
			path:       "/_all_dbs",
			headers:    map[string]string{"Authorization": basicAuth(userReader)},
			wantStatus: http.StatusForbidden,
			wantJSON: map[string]interface{}{
				"error":  "forbidden",
				"reason": "Admin privileges required",
			},
		},
		{
			name:       "all dbs, descending",
			method:     http.MethodGet,
			path:       "/_all_dbs?descending=true",
			headers:    map[string]string{"Authorization": basicAuth(userAdmin)},
			wantStatus: http.StatusOK,
			wantJSON:   []string{"db2", "db1"},
		},
		{
			name:       "db info",
			method:     http.MethodGet,
			path:       "/db1",
			headers:    map[string]string{"Authorization": basicAuth(userAdmin)},
			wantStatus: http.StatusOK,
			wantJSON: map[string]interface{}{
				"db_name":         "db1",
				"compact_running": false,
				"data_size":       0,
				"disk_size":       0,
				"doc_count":       0,
				"doc_del_count":   0,
				"update_seq":      "",
			},
		},
		{
			name:       "db info HEAD",
			method:     http.MethodHead,
			path:       "/db1",
			headers:    map[string]string{"Authorization": basicAuth(userAdmin)},
			wantStatus: http.StatusOK,
		},
		{
			name:       "start session, no content type header",
			method:     http.MethodPost,
			path:       "/_session",
			body:       strings.NewReader(`name=root&password=abc123`),
			wantStatus: http.StatusUnsupportedMediaType,
			wantJSON: map[string]interface{}{
				"error":  "bad_content_type",
				"reason": "Content-Type must be 'application/x-www-form-urlencoded' or 'application/json'",
			},
		},
		{
			name:       "start session, invalid content type",
			method:     http.MethodPost,
			path:       "/_session",
			body:       strings.NewReader(`name=root&password=abc123`),
			headers:    map[string]string{"Content-Type": "application/xml"},
			wantStatus: http.StatusUnsupportedMediaType,
			wantJSON: map[string]interface{}{
				"error":  "bad_content_type",
				"reason": "Content-Type must be 'application/x-www-form-urlencoded' or 'application/json'",
			},
		},
		{
			name:       "start session, no user name",
			method:     http.MethodPost,
			path:       "/_session",
			body:       strings.NewReader(`{}`),
			headers:    map[string]string{"Content-Type": "application/json"},
			wantStatus: http.StatusBadRequest,
			wantJSON: map[string]interface{}{
				"error":  "bad_request",
				"reason": "request body must contain a username",
			},
		},
		{
			name:       "start session, success",
			method:     http.MethodPost,
			path:       "/_session",
			body:       strings.NewReader(`{"name":"admin","password":"abc123"}`),
			headers:    map[string]string{"Content-Type": "application/json"},
			wantStatus: http.StatusOK,
			wantJSON: map[string]interface{}{
				"ok":    true,
				"name":  userAdmin,
				"roles": []string{"_admin"},
			},
		},
		{
			name:       "delete session",
			method:     http.MethodDelete,
			path:       "/_session",
			authUser:   userAdmin,
			wantStatus: http.StatusOK,
			wantJSON: map[string]interface{}{
				"ok": true,
			},
		},
		{
			name:       "_up",
			method:     http.MethodGet,
			path:       "/_up",
			wantStatus: http.StatusOK,
			wantJSON: map[string]interface{}{
				"status": "ok",
			},
		},
		{
			name:       "all config",
			method:     http.MethodGet,
			path:       "/_node/_local/_config",
			authUser:   userAdmin,
			wantStatus: http.StatusOK,
			wantJSON: map[string]interface{}{
				"couchdb": map[string]interface{}{
					"users_db_suffix": "_users",
				},
			},
		},
		{
			name:       "all config, non-admin",
			method:     http.MethodGet,
			path:       "/_node/_local/_config",
			authUser:   userReader,
			wantStatus: http.StatusForbidden,
			wantJSON: map[string]interface{}{
				"error":  "forbidden",
				"reason": "Admin privileges required",
			},
		},
		{
			name:       "all config, no such node",
			method:     http.MethodGet,
			path:       "/_node/asdf/_config",
			authUser:   userAdmin,
			wantStatus: http.StatusNotFound,
			wantJSON: map[string]interface{}{
				"error":  "not_found",
				"reason": "no such node: asdf",
			},
		},
		{
			name:       "config section",
			method:     http.MethodGet,
			path:       "/_node/_local/_config/couchdb",
			authUser:   userAdmin,
			wantStatus: http.StatusOK,
			wantJSON: map[string]interface{}{
				"users_db_suffix": "_users",
			},
		},
		{
			name:       "config key",
			method:     http.MethodGet,
			path:       "/_node/_local/_config/couchdb/users_db_suffix",
			authUser:   userAdmin,
			wantStatus: http.StatusOK,
			wantJSON:   "_users",
		},
		{
			name:       "reload config",
			method:     http.MethodPost,
			path:       "/_node/_local/_config/_reload",
			authUser:   userAdmin,
			wantStatus: http.StatusOK,
			wantJSON:   map[string]bool{"ok": true},
		},
		{
			name:       "set new config key",
			method:     http.MethodPut,
			path:       "/_node/_local/_config/foo/bar",
			body:       strings.NewReader(`"oink"`),
			authUser:   userAdmin,
			wantStatus: http.StatusOK,
			wantJSON:   "",
		},
		{
			name:       "set existing config key",
			method:     http.MethodPut,
			path:       "/_node/_local/_config/couchdb/users_db_suffix",
			body:       strings.NewReader(`"oink"`),
			authUser:   userAdmin,
			wantStatus: http.StatusOK,
			wantJSON:   "_users",
		},
		{
			name:       "delete existing config key",
			method:     http.MethodDelete,
			path:       "/_node/_local/_config/couchdb/users_db_suffix",
			authUser:   userAdmin,
			wantStatus: http.StatusOK,
			wantJSON:   "_users",
		},
		{
			name:       "delete non-existent config key",
			method:     http.MethodDelete,
			path:       "/_node/_local/_config/foo/bar",
			authUser:   userAdmin,
			wantStatus: http.StatusNotFound,
			wantJSON: map[string]interface{}{
				"error":  "not_found",
				"reason": "unknown_config_value",
			},
		},
		{
			name: "set config not supported by config backend",
			extraOptions: []Option{
				WithConfig(&readOnlyConfig{
					Config: config.Default(),
				}),
			},
			method:     http.MethodPut,
			path:       "/_node/_local/_config/foo/bar",
			body:       strings.NewReader(`"oink"`),
			authUser:   userAdmin,
			wantStatus: http.StatusMethodNotAllowed,
			wantJSON: map[string]interface{}{
				"error":  "method_not_allowed",
				"reason": "configuration is read-only",
			},
		},
		{
			name: "delete config not supported by config backend",
			extraOptions: []Option{
				WithConfig(&readOnlyConfig{
					Config: config.Default(),
				}),
			},
			method:     http.MethodDelete,
			path:       "/_node/_local/_config/foo/bar",
			authUser:   userAdmin,
			wantStatus: http.StatusMethodNotAllowed,
			wantJSON: map[string]interface{}{
				"error":  "method_not_allowed",
				"reason": "configuration is read-only",
			},
		},
		{
			name: "uuids",
			extraOptions: []Option{
				WithConfig(&readOnlyConfig{
					Config: config.Default(),
				}),
			},
			method:     http.MethodGet,
			path:       "/_uuids",
			wantStatus: http.StatusOK,
			wantJSON: map[string]interface{}{
				"uuids": []string{"x"},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client, err := kivik.New("fs", "testdata/fsdb")
			if err != nil {
				t.Fatal(err)
			}
			us := testUserStore(t)
			const secret = "foo"
			opts := append([]Option{
				WithUserStores(us),
				WithAuthHandlers(auth.BasicAuth()),
				WithAuthHandlers(auth.CookieAuth(secret, time.Hour)),
			}, tt.extraOptions...)

			s := New(client, opts...)
			req, err := http.NewRequest(tt.method, tt.path, tt.body)
			if err != nil {
				t.Fatal(err)
			}
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			if tt.authUser != "" {
				user, err := us.UserCtx(context.Background(), tt.authUser)
				if err != nil {
					t.Fatal(err)
				}
				req.AddCookie(&http.Cookie{
					Name:  kivik.SessionCookieName,
					Value: auth.CreateAuthToken(user.Name, user.Salt, secret, time.Now().Unix()),
				})
			}

			rec := httptest.NewRecorder()
			s.ServeHTTP(rec, req)

			res := rec.Result()
			if res.StatusCode != tt.wantStatus {
				t.Errorf("Unexpected response status: %d", res.StatusCode)
			}
			if d := testy.DiffAsJSON(tt.wantJSON, res.Body); d != nil {
				t.Error(d)
			}
		})
	}
}

type readOnlyConfig struct {
	config.Config
	// To prevent the embedded methods from being accessible
	SetKey int
	Delete int
}
