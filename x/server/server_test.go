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
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	_ "github.com/go-kivik/kivik/v4/x/fsdb"     // Filesystem driver
	_ "github.com/go-kivik/kivik/v4/x/memorydb" // Memory driver
	"github.com/go-kivik/kivik/v4/x/server/auth"
	"github.com/go-kivik/kivik/v4/x/server/config"
)

var v = validator.New(validator.WithRequiredStructEnabled())

const (
	userAdmin      = "admin"
	userBob        = "bob"
	userAlice      = "alice"
	userCharlie    = "charlie"
	userDavid      = "davic"
	userErin       = "erin"
	userFrank      = "frank"
	userReplicator = "replicator"
	userDBUpdates  = "db_updates"
	userDesign     = "design"
	testPassword   = "abc123"
	roleFoo        = "foo"
	roleBar        = "bar"
	roleBaz        = "baz"
)

func testUserStore(t *testing.T) *auth.MemoryUserStore {
	t.Helper()
	us := auth.NewMemoryUserStore()
	if err := us.AddUser(userAdmin, testPassword, []string{auth.RoleAdmin}); err != nil {
		t.Fatal(err)
	}
	if err := us.AddUser(userBob, testPassword, []string{auth.RoleReader}); err != nil {
		t.Fatal(err)
	}
	if err := us.AddUser(userAlice, testPassword, []string{auth.RoleWriter}); err != nil {
		t.Fatal(err)
	}
	if err := us.AddUser(userCharlie, testPassword, []string{auth.RoleWriter, roleFoo}); err != nil {
		t.Fatal(err)
	}
	if err := us.AddUser(userDavid, testPassword, []string{auth.RoleWriter, roleBar}); err != nil {
		t.Fatal(err)
	}
	if err := us.AddUser(userErin, testPassword, []string{auth.RoleWriter}); err != nil {
		t.Fatal(err)
	}
	if err := us.AddUser(userFrank, testPassword, []string{auth.RoleWriter, roleBaz}); err != nil {
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

type serverTest struct {
	name         string
	client       *kivik.Client
	driver, dsn  string
	init         func(t *testing.T, client *kivik.Client)
	extraOptions []Option
	method       string
	path         string
	headers      map[string]string
	authUser     string
	body         io.Reader
	wantStatus   int
	wantBodyRE   string
	wantJSON     interface{}
	check        func(t *testing.T, client *kivik.Client)

	// if target is specified, it is expected to be a struct into which the
	// response body will be unmarshaled, then validated.
	target interface{}
}

type serverTests []serverTest

func (s serverTests) Run(t *testing.T) {
	t.Helper()
	for _, tt := range s {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			driver, dsn := "fs", "testdata/fsdb"
			if tt.dsn != "" {
				dsn = tt.dsn
			}
			client := tt.client
			if client == nil {
				if tt.driver != "" {
					driver = tt.driver
				}
				if driver == "fs" {
					dsn = testy.CopyTempDir(t, dsn, 0)
					t.Cleanup(func() {
						_ = os.RemoveAll(dsn)
					})
				}
				var err error
				client, err = kivik.New(driver, dsn)
				if err != nil {
					t.Fatal(err)
				}
			}
			if tt.init != nil {
				tt.init(t, client)
			}
			us := testUserStore(t)
			const secret = "foo"
			opts := append([]Option{
				WithUserStores(us),
				WithAuthHandlers(auth.BasicAuth()),
				WithAuthHandlers(auth.CookieAuth(secret, time.Hour)),
			}, tt.extraOptions...)

			s := New(client, opts...)
			body := tt.body
			if body == nil {
				body = strings.NewReader("")
			}
			req, err := http.NewRequestWithContext(t.Context(), tt.method, tt.path, body)
			if err != nil {
				t.Fatal(err)
			}
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			if tt.authUser != "" {
				user, err := us.UserCtx(t.Context(), tt.authUser)
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
				t.Errorf("Unexpected response status: %d %s", res.StatusCode, http.StatusText(res.StatusCode))
			}
			switch {
			case tt.target != nil:
				if err := json.NewDecoder(res.Body).Decode(tt.target); err != nil {
					t.Fatal(err)
				}
				if err := v.Struct(tt.target); err != nil {
					t.Fatalf("response does not match expectations: %s\n%v", err, tt.target)
				}
			case tt.wantBodyRE != "":
				re := regexp.MustCompile(tt.wantBodyRE)
				body, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatal(err)
				}
				if !re.Match(body) {
					t.Errorf("Unexpected response body:\n%s", body)
				}
			default:
				if d := testy.DiffAsJSON(tt.wantJSON, res.Body); d != nil {
					t.Error(d)
				}
			}
			if tt.check != nil {
				tt.check(t, client)
			}
		})
	}
}

func TestServer(t *testing.T) {
	t.Parallel()

	tests := serverTests{
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
			wantStatus: http.StatusOK,
			wantJSON:   []interface{}{},
		},
		{
			name:       "all dbs",
			method:     http.MethodGet,
			path:       "/_all_dbs",
			headers:    map[string]string{"Authorization": basicAuth(userAdmin)},
			wantStatus: http.StatusOK,
			wantJSON:   []string{"bobsdb", "db1", "db2"},
		},
		{
			name:       "all dbs, cookie auth",
			method:     http.MethodGet,
			path:       "/_all_dbs",
			authUser:   userAdmin,
			wantStatus: http.StatusOK,
			wantJSON:   []string{"bobsdb", "db1", "db2"},
		},
		{
			name:       "all dbs, non-admin",
			method:     http.MethodGet,
			path:       "/_all_dbs",
			headers:    map[string]string{"Authorization": basicAuth(userBob)},
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
			wantJSON:   []string{"db2", "db1", "bobsdb"},
		},
		{
			name:       "db info",
			method:     http.MethodGet,
			path:       "/db2",
			headers:    map[string]string{"Authorization": basicAuth(userAdmin)},
			wantStatus: http.StatusOK,
			wantJSON: map[string]interface{}{
				"db_name":         "db2",
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
			path:       "/db2",
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
			authUser:   userBob,
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
			name: "too many uuids",
			extraOptions: []Option{
				WithConfig(&readOnlyConfig{
					Config: config.Default(),
				}),
			},
			method:     http.MethodGet,
			path:       "/_uuids?count=99999",
			wantStatus: http.StatusBadRequest,
			wantJSON: map[string]interface{}{
				"error":  "bad_request",
				"reason": "count must not exceed 1000",
			},
		},
		{
			name: "invalid count",
			extraOptions: []Option{
				WithConfig(&readOnlyConfig{
					Config: config.Default(),
				}),
			},
			method:     http.MethodGet,
			path:       "/_uuids?count=chicken",
			wantStatus: http.StatusBadRequest,
			wantJSON: map[string]interface{}{
				"error":  "bad_request",
				"reason": "count must be a positive integer",
			},
		},
		{
			name: "random uuids",
			extraOptions: []Option{
				WithConfig(&readOnlyConfig{
					Config: config.Map(
						map[string]map[string]string{
							"uuids": {"algorithm": "random"},
						},
					),
				}),
			},
			method:     http.MethodGet,
			path:       "/_uuids",
			wantStatus: http.StatusOK,
			target: new(struct {
				UUIDs []string `json:"uuids" validate:"required,len=1,dive,required,len=32,hexadecimal"`
			}),
		},
		{
			name: "many random uuids",
			extraOptions: []Option{
				WithConfig(&readOnlyConfig{
					Config: config.Map(
						map[string]map[string]string{
							"uuids": {"algorithm": "random"},
						},
					),
				}),
			},
			method:     http.MethodGet,
			path:       "/_uuids?count=10",
			wantStatus: http.StatusOK,
			target: new(struct {
				UUIDs []string `json:"uuids" validate:"required,len=10,dive,required,len=32,hexadecimal"`
			}),
		},
		{
			name: "sequential uuids",
			extraOptions: []Option{
				WithConfig(&readOnlyConfig{
					Config: config.Default(),
				}),
			},
			method:     http.MethodGet,
			path:       "/_uuids",
			wantStatus: http.StatusOK,
			target: new(struct {
				UUIDs []string `json:"uuids" validate:"required,len=1,dive,required,len=32,hexadecimal"`
			}),
		},
		{
			name: "many random uuids",
			extraOptions: []Option{
				WithConfig(&readOnlyConfig{
					Config: config.Default(),
				}),
			},
			method:     http.MethodGet,
			path:       "/_uuids?count=10",
			wantStatus: http.StatusOK,
			target: new(struct {
				UUIDs []string `json:"uuids" validate:"required,len=10,dive,required,len=32,hexadecimal"`
			}),
		},
		{
			name: "one utc random uuid",
			extraOptions: []Option{
				WithConfig(&readOnlyConfig{
					Config: config.Map(
						map[string]map[string]string{
							"uuids": {"algorithm": "utc_random"},
						},
					),
				}),
			},
			method:     http.MethodGet,
			path:       "/_uuids",
			wantStatus: http.StatusOK,
			target: new(struct {
				UUIDs []string `json:"uuids" validate:"required,len=1,dive,required,len=32,hexadecimal"`
			}),
		},
		{
			name: "10 utc random uuids",
			extraOptions: []Option{
				WithConfig(&readOnlyConfig{
					Config: config.Map(
						map[string]map[string]string{
							"uuids": {"algorithm": "utc_random"},
						},
					),
				}),
			},
			method:     http.MethodGet,
			path:       "/_uuids?count=10",
			wantStatus: http.StatusOK,
			target: new(struct {
				UUIDs []string `json:"uuids" validate:"required,len=10,dive,required,len=32,hexadecimal"`
			}),
		},
		{
			name: "one utc id uuid",
			extraOptions: []Option{
				WithConfig(&readOnlyConfig{
					Config: config.Map(
						map[string]map[string]string{
							"uuids": {
								"algorithm":     "utc_id",
								"utc_id_suffix": "oink",
							},
						},
					),
				}),
			},
			method:     http.MethodGet,
			path:       "/_uuids",
			wantStatus: http.StatusOK,
			target: new(struct {
				UUIDs []string `json:"uuids" validate:"required,len=1,dive,required,len=18,endswith=oink"`
			}),
		},
		{
			name: "10 utc id uuids",
			extraOptions: []Option{
				WithConfig(&readOnlyConfig{
					Config: config.Map(
						map[string]map[string]string{
							"uuids": {
								"algorithm":     "utc_id",
								"utc_id_suffix": "oink",
							},
						},
					),
				}),
			},
			method:     http.MethodGet,
			path:       "/_uuids?count=10",
			wantStatus: http.StatusOK,
			target: new(struct {
				UUIDs []string `json:"uuids" validate:"required,len=10,dive,required,len=18,endswith=oink"`
			}),
		},
		{
			name:       "create db",
			method:     http.MethodPut,
			path:       "/db3",
			authUser:   userAdmin,
			wantStatus: http.StatusCreated,
			wantJSON: map[string]interface{}{
				"ok": true,
			},
		},
		{
			name:       "delete db, not found",
			method:     http.MethodDelete,
			path:       "/db3",
			authUser:   userAdmin,
			wantStatus: http.StatusNotFound,
			wantJSON: map[string]interface{}{
				"error":  "not_found",
				"reason": "database does not exist",
			},
		},
		{
			name:       "delete db",
			method:     http.MethodDelete,
			path:       "/db2",
			authUser:   userAdmin,
			wantStatus: http.StatusOK,
			wantJSON: map[string]interface{}{
				"ok": true,
			},
		},
		{
			name:   "post document",
			driver: "memory",
			init: func(t *testing.T, client *kivik.Client) { //nolint:thelper // not a helper
				if err := client.CreateDB(t.Context(), "db1", nil); err != nil {
					t.Fatal(err)
				}
			},
			method:     http.MethodPost,
			path:       "/db1",
			body:       strings.NewReader(`{"foo":"bar"}`),
			authUser:   userAdmin,
			wantStatus: http.StatusCreated,
			target: &struct {
				ID  string `json:"id" validate:"required,uuid"`
				Rev string `json:"rev" validate:"required,startswith=1-"`
				OK  bool   `json:"ok" validate:"required,eq=true"`
			}{},
		},
		{
			name:       "get document",
			method:     http.MethodGet,
			path:       "/db1/foo",
			authUser:   userAdmin,
			wantStatus: http.StatusOK,
			wantJSON: map[string]interface{}{
				"_id":  "foo",
				"_rev": "1-beea34a62a215ab051862d1e5d93162e",
				"foo":  "bar",
			},
		},
		{
			name:       "all dbs stats",
			method:     http.MethodGet,
			path:       "/_dbs_info",
			authUser:   userAdmin,
			wantStatus: http.StatusOK,
			wantJSON: []map[string]interface{}{
				{
					"compact_running": false,
					"data_size":       0,
					"db_name":         "bobsdb",
					"disk_size":       0,
					"doc_count":       0,
					"doc_del_count":   0,
					"update_seq":      "",
				},
				{
					"compact_running": false,
					"data_size":       0,
					"db_name":         "db1",
					"disk_size":       0,
					"doc_count":       0,
					"doc_del_count":   0,
					"update_seq":      "",
				},
				{
					"compact_running": false,
					"data_size":       0,
					"db_name":         "db2",
					"disk_size":       0,
					"doc_count":       0,
					"doc_del_count":   0,
					"update_seq":      "",
				},
			},
		},
		{
			name:       "dbs stats",
			method:     http.MethodPost,
			path:       "/_dbs_info",
			authUser:   userAdmin,
			headers:    map[string]string{"Content-Type": "application/json"},
			body:       strings.NewReader(`{"keys":["db1","notfound"]}`),
			wantStatus: http.StatusOK,
			wantJSON: []map[string]interface{}{
				{
					"compact_running": false,
					"data_size":       0,
					"db_name":         "db1",
					"disk_size":       0,
					"doc_count":       0,
					"doc_del_count":   0,
					"update_seq":      "",
				},
				nil,
			},
		},
		{
			name:       "get security",
			method:     http.MethodGet,
			path:       "/db1/_security",
			authUser:   userAdmin,
			wantStatus: http.StatusOK,
			wantJSON: map[string]interface{}{
				"admins": map[string]interface{}{
					"names": []string{"superuser"},
					"roles": []string{"admins"},
				},
				"members": map[string]interface{}{
					"names": []string{"user1", "user2"},
					"roles": []string{"developers"},
				},
			},
		},
		func() serverTest {
			const want = `{"admins":{"names":["superuser"],"roles":["admins"]},"members":{"names":["user1","user2"],"roles":["developers"]}}`
			return serverTest{
				name:       "put security",
				method:     http.MethodPut,
				path:       "/db2/_security",
				authUser:   userAdmin,
				headers:    map[string]string{"Content-Type": "application/json"},
				body:       strings.NewReader(want),
				wantStatus: http.StatusOK,
				wantJSON: map[string]interface{}{
					"ok": true,
				},
				check: func(t *testing.T, client *kivik.Client) { //nolint:thelper // Not a helper
					sec, err := client.DB("db2").Security(t.Context())
					if err != nil {
						t.Fatal(err)
					}
					if d := testy.DiffAsJSON([]byte(want), sec); d != nil {
						t.Errorf("Unexpected final result: %s", d)
					}
				},
			}
		}(),
		{
			name:       "put security, unauthorized",
			method:     http.MethodPut,
			path:       "/db2/_security",
			headers:    map[string]string{"Content-Type": "application/json"},
			body:       strings.NewReader(`{"admins":{"names":["bob"]}}`),
			wantStatus: http.StatusUnauthorized,
			wantJSON: map[string]interface{}{
				"error":  "unauthorized",
				"reason": "User not authenticated",
			},
		},
		{
			name:       "put security, no admin access",
			method:     http.MethodPut,
			authUser:   userBob,
			path:       "/db2/_security",
			headers:    map[string]string{"Content-Type": "application/json"},
			body:       strings.NewReader(`{"admins":{"names":["bob"]}}`),
			wantStatus: http.StatusForbidden,
			wantJSON: map[string]interface{}{
				"error":  "forbidden",
				"reason": "User lacks sufficient privileges",
			},
		},
		{
			name:       "put security, correct admin user",
			method:     http.MethodPut,
			authUser:   userErin,
			path:       "/bobsdb/_security",
			headers:    map[string]string{"Content-Type": "application/json"},
			body:       strings.NewReader(`{"admins":{"names":["bob"]}}`),
			wantStatus: http.StatusOK,
			wantJSON: map[string]interface{}{
				"ok": true,
			},
		},
		{
			name:       "put security, correct admin role",
			method:     http.MethodPut,
			authUser:   userFrank,
			path:       "/bobsdb/_security",
			headers:    map[string]string{"Content-Type": "application/json"},
			body:       strings.NewReader(`{"admins":{"names":["bob"]}}`),
			wantStatus: http.StatusOK,
			wantJSON: map[string]interface{}{
				"ok": true,
			},
		},
		{
			name:       "db info, unauthenticated",
			method:     http.MethodHead,
			path:       "/bobsdb",
			wantStatus: http.StatusUnauthorized,
			wantJSON: map[string]interface{}{
				"error":  "unauthorized",
				"reason": "User not authenticated",
			},
		},
		{
			name:       "db info, authenticated wrong user, wrong role",
			method:     http.MethodHead,
			authUser:   userAlice,
			path:       "/bobsdb",
			wantStatus: http.StatusForbidden,
			wantJSON: map[string]interface{}{
				"error":  "forbidden",
				"reason": "User lacks sufficient privileges",
			},
		},
		{
			name:       "db info, authenticated correct user",
			method:     http.MethodHead,
			authUser:   userBob,
			path:       "/bobsdb",
			wantStatus: http.StatusOK,
		},
		{
			name:       "db info, authenticated wrong role",
			method:     http.MethodHead,
			authUser:   userCharlie,
			path:       "/bobsdb",
			wantStatus: http.StatusForbidden,
			wantJSON: map[string]interface{}{
				"error":  "forbidden",
				"reason": "User lacks sufficient privileges",
			},
		},
		{
			name:       "db info, authenticated correct role",
			method:     http.MethodHead,
			authUser:   userDavid,
			path:       "/bobsdb",
			wantStatus: http.StatusOK,
		},
		{
			name:       "db info, authenticated as admin user",
			method:     http.MethodHead,
			authUser:   userErin,
			path:       "/bobsdb",
			wantStatus: http.StatusOK,
		},
		{
			name:       "db info, authenticated as admin role",
			method:     http.MethodHead,
			authUser:   userFrank,
			path:       "/bobsdb",
			wantStatus: http.StatusOK,
		},
	}

	tests.Run(t)
}

type readOnlyConfig struct {
	config.Config
	// To prevent the embedded methods from being accessible
	SetKey int
	Delete int
}
