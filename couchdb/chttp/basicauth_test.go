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

func TestBasicAuthRoundTrip(t *testing.T) {
	type rtTest struct {
		name     string
		auth     *BasicAuth
		req      *http.Request
		expected *http.Response
		cleanup  func()
	}
	tests := []rtTest{
		{
			name: "Provided transport",
			req:  httptest.NewRequest("GET", "/", nil),
			auth: &BasicAuth{
				Username: "foo",
				Password: "bar",
				transport: customTransport(func(req *http.Request) (*http.Response, error) {
					u, p, ok := req.BasicAuth()
					if !ok {
						t.Error("BasicAuth not set in request")
					}
					if u != "foo" || p != "bar" { // nolint: goconst
						t.Errorf("Unexpected user/password: %s/%s", u, p)
					}
					return &http.Response{StatusCode: 200}, nil
				}),
			},
			expected: &http.Response{StatusCode: 200},
		},
		func() rtTest {
			h := func(w http.ResponseWriter, r *http.Request) {
				u, p, ok := r.BasicAuth()
				if !ok {
					t.Error("BasicAuth not set in request")
				}
				if u != "foo" || p != "bar" {
					t.Errorf("Unexpected user/password: %s/%s", u, p)
				}
				w.Header().Set("Date", "Wed, 01 Nov 2017 19:32:41 GMT")
				w.Header().Set("Content-Type", "application/json")
			}
			s := httptest.NewServer(http.HandlerFunc(h))
			return rtTest{
				name: "default transport",
				auth: &BasicAuth{
					Username:  "foo",
					Password:  "bar",
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

func TestJWTAuthRoundTrip(t *testing.T) {
	type rtTest struct {
		name     string
		auth     *JWTAuth
		req      *http.Request
		expected *http.Response
		cleanup  func()
	}
	tests := []rtTest{
		{
			name: "Provided transport",
			req:  httptest.NewRequest("GET", "/", nil),
			auth: &JWTAuth{
				Token: "token",
				transport: customTransport(func(req *http.Request) (*http.Response, error) {
					if h := req.Header.Get("Authorization"); h != "Bearer token" {
						t.Errorf("Unexpected authorization header: %s", h)
					}
					return &http.Response{StatusCode: 200}, nil
				}),
			},
			expected: &http.Response{StatusCode: 200},
		},
		func() rtTest {
			h := func(w http.ResponseWriter, r *http.Request) {
				if h := r.Header.Get("Authorization"); h != "Bearer token" {
					t.Errorf("Unexpected authorization header: %s", h)
				}
				w.Header().Set("Date", "Wed, 01 Nov 2017 19:32:41 GMT")
				w.Header().Set("Content-Type", "application/json")
			}
			s := httptest.NewServer(http.HandlerFunc(h))
			return rtTest{
				name: "default transport",
				auth: &JWTAuth{
					Token:     "token",
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
