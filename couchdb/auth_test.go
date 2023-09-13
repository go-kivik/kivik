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

package couchdb

import (
	"context"
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/internal/nettest"
)

func TestAuthenticationOptions(t *testing.T) {
	type test struct {
		handler func(*testing.T) http.Handler
		setup   func(*testing.T, *client)
		options kivik.Option
		status  int
		err     string
	}

	tests := testy.NewTable()
	tests.Add("BasicAuth", test{
		handler: func(t *testing.T) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if h := r.Header.Get("Authorization"); h != "Basic Ym9iOmFiYzEyMw==" {
					t.Errorf("Unexpected Auth header: %s\n", h)
				}
				w.WriteHeader(200)
				_, _ = w.Write([]byte(`{}`))
			})
		},
		options: BasicAuth("bob", "abc123"),
	})
	tests.Add("CookieAuth", test{
		handler: func(t *testing.T) http.Handler {
			expectedPaths := []string{"/_session", "/"}
			i := -1
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				i++
				if p := r.URL.Path; p != expectedPaths[i] {
					t.Errorf("Unexpected path: %s\n", p)
				}
				w.WriteHeader(200)
				_, _ = w.Write([]byte(`{}`))
			})
		},
		options: CookieAuth("bob", "abc123"),
	})
	tests.Add("ProxyAuth", test{
		handler: func(t *testing.T) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if h := r.Header.Get("X-Auth-CouchDB-UserName"); h != "bob" {
					t.Errorf("Unexpected X-Auth-CouchDB-UserName header: %s", h)
				}
				if h := r.Header.Get("X-Auth-CouchDB-Roles"); h != "users,admins" {
					t.Errorf("Unexpected X-Auth-CouchDB-Roles header: %s", h)
				}
				if h := r.Header.Get("Moo"); h != "adedb8d002eb53a52faba80e82cb1fc6d57bca74" {
					t.Errorf("Token header override failed: %s instead of 'adedb8d002eb53a52faba80e82cb1fc6d57bca74'", h)
				}
				w.WriteHeader(200)
				_, _ = w.Write([]byte(`{}`))
			})
		},
		options: ProxyAuth(
			"bob",
			"abc123",
			[]string{"users", "admins"},
			map[string]string{"X-Auth-CouchDB-Token": "moo"},
		),
	})
	tests.Add("JWTAuth", test{
		handler: func(t *testing.T) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if h := r.Header.Get("Authorization"); h != "Bearer tokentoken" {
					t.Errorf("Unexpected Auth header: %s\n", h)
				}
				w.WriteHeader(200)
				_, _ = w.Write([]byte(`{}`))
			})
		},
		options: JWTAuth("tokentoken"),
	})

	driver := &couch{}
	tests.Run(t, func(t *testing.T, tt test) {
		s := nettest.NewHTTPTestServer(t, tt.handler(t))
		defer s.Close()
		driverClient, err := driver.NewClient(s.URL, tt.options)
		if err != nil {
			t.Fatal(err)
		}
		client := driverClient.(*client)
		if tt.setup != nil {
			tt.setup(t, client)
		}
		_, err = client.Version(context.Background())
		testy.StatusErrorRE(t, tt.err, tt.status, err)
	})
}
