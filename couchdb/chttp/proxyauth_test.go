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

package chttp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/flimzy/testy"
)

const (
	rolesTest    = "users,admins"
	secretTest   = "abc123"
	tokenTest    = "adedb8d002eb53a52faba80e82cb1fc6d57bca74"
	usernameTest = "bob"
)

func TestProxyAuthRoundTrip(t *testing.T) {
	type rtTest struct {
		name     string
		auth     *ProxyAuth
		req      *http.Request
		expected *http.Response
		cleanup  func()
	}
	tests := []rtTest{
		{
			name: "Provided transport",
			req:  httptest.NewRequest("GET", "/", nil),
			auth: &ProxyAuth{
				Username: usernameTest,
				Secret:   secretTest,
				Roles:    []string{"users", "admins"},
				transport: customTransport(func(req *http.Request) (*http.Response, error) {
					username := req.Header.Get("X-Auth-CouchDB-UserName")
					if username != usernameTest {
						t.Errorf("Unexpected X-Auth-CouchDB-UserName value: %s", username)
					}

					roles := req.Header.Get("X-Auth-CouchDB-Roles")
					if roles != rolesTest {
						t.Errorf("Unexpected X-Auth-CouchDB-Roles value: %s", roles)
					}

					token := req.Header.Get("X-Auth-CouchDB-Token")
					if token != tokenTest {
						t.Errorf("Unexpected X-Auth-CouchDB-Token value: %s", token)
					}

					return &http.Response{StatusCode: 200}, nil
				}),
			},
			expected: &http.Response{StatusCode: 200},
		},
		{
			name: "Secret is an empty string",
			req:  httptest.NewRequest("GET", "/", nil),
			auth: &ProxyAuth{
				Username: usernameTest,
				Secret:   "",
				Roles:    []string{"users", "admins"},
				transport: customTransport(func(req *http.Request) (*http.Response, error) {
					token := req.Header.Get("X-Auth-CouchDB-Token")
					if token != "" {
						t.Error("Setting secret to an empty string did not unset the X-Auth-CouchDB-Token header")
					}

					return &http.Response{StatusCode: 200}, nil
				}),
			},
			expected: &http.Response{StatusCode: 200},
		},
		{
			name: "Overridden header names",
			req:  httptest.NewRequest("GET", "/", nil),
			auth: &ProxyAuth{
				Username: usernameTest,
				Secret:   secretTest,
				Roles:    []string{"users", "admins"},
				Headers: http.Header{
					"X-Auth-Couchdb-Token":    []string{"moo"},
					"X-Auth-Couchdb-Username": []string{"cow"},
					"X-Auth-Couchdb-Roles":    []string{"bovine"},
				},
				transport: customTransport(func(req *http.Request) (*http.Response, error) {
					username := req.Header.Get("Cow")
					if username != usernameTest {
						t.Error("Username header override failed")
					}

					roles := req.Header.Get("Bovine")
					if roles != rolesTest {
						t.Error("Roles header override failed")
					}

					token := req.Header.Get("Moo")
					if token != tokenTest {
						t.Error("Token header override failed")
					}

					return &http.Response{StatusCode: 200}, nil
				}),
			},
			expected: &http.Response{StatusCode: 200},
		},
		func() rtTest {
			h := func(w http.ResponseWriter, r *http.Request) {
				username := r.Header.Get("X-Auth-CouchDB-UserName")
				if username != usernameTest {
					t.Errorf("Unexpected X-Auth-CouchDB-UserName value: %s", username)
				}

				roles := r.Header.Get("X-Auth-CouchDB-Roles")
				if roles != rolesTest {
					t.Errorf("Unexpected X-Auth-CouchDB-Roles value: %s", roles)
				}

				token := r.Header.Get("X-Auth-CouchDB-Token")
				if token != tokenTest {
					t.Errorf("Unexpected X-Auth-CouchDB-Token value: %s", token)
				}

				w.Header().Set("Date", "Wed, 01 Nov 2017 19:32:41 GMT")
				w.Header().Set("Content-Type", "application/json")
			}
			s := httptest.NewServer(http.HandlerFunc(h))
			return rtTest{
				name: "default transport",
				auth: &ProxyAuth{
					Username:  usernameTest,
					Secret:    secretTest,
					Roles:     []string{"users", "admins"},
					transport: http.DefaultTransport,
				},
				req: httptest.NewRequest("GET", s.URL, nil),
				expected: &http.Response{
					Status:     "200 OK",
					StatusCode: 200,
					Proto:      "HTTP/1.1",
					ProtoMajor: 1,
					ProtoMinor: 1,
					Header: http.Header{
						"Content-Length": {"0"},
						"Content-Type":   {"application/json"},
						"Date":           {"Wed, 01 Nov 2017 19:32:41 GMT"},
					},
				},
				cleanup: func() { s.Close() },
			}
		}(),
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := test.auth.RoundTrip(test.req)
			if err != nil {
				t.Fatal(err)
			}
			res.Body = nil
			res.Request = nil
			if d := testy.DiffInterface(test.expected, res); d != nil {
				t.Error(d)
			}
		})
	}
}
