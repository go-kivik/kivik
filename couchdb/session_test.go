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

// GopherJS can't run a test server

package couchdb

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/flimzy/testy"

	kivik "github.com/go-kivik/kivik/v4"
)

func TestSession(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		body      string
		expected  interface{}
		errStatus int
		err       string
	}{
		{
			name:   "valid",
			status: http.StatusOK,
			body:   `{"ok":true,"userCtx":{"name":"admin","roles":["_admin"]},"info":{"authentication_db":"_users","authentication_handlers":["oauth","cookie","default"],"authenticated":"cookie"}}`,
			expected: &kivik.Session{
				Name:                   "admin",
				Roles:                  []string{"_admin"},
				AuthenticationMethod:   "cookie",
				AuthenticationHandlers: []string{"oauth", "cookie", "default"},
				RawResponse:            []byte(`{"ok":true,"userCtx":{"name":"admin","roles":["_admin"]},"info":{"authentication_db":"_users","authentication_handlers":["oauth","cookie","default"],"authenticated":"cookie"}}`),
			},
		},
		{
			name:      "invalid response",
			status:    http.StatusOK,
			body:      `{"userCtx":"asdf"}`,
			errStatus: http.StatusBadGateway,
			err:       "json: cannot unmarshal string into Go ",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(test.body))
			}))
			client, err := kivik.New("couch", s.URL)
			if err != nil {
				t.Fatal(err)
			}
			session, err := client.Session(context.Background())
			testy.StatusErrorRE(t, test.err, test.errStatus, err)
			if d := testy.DiffInterface(test.expected, session); d != nil {
				t.Error(d)
			}
		})
	}
}
