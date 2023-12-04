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

package kivikd

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/go-kivik/kivik/v4/x/kivikd/auth"
	"github.com/go-kivik/kivik/v4/x/kivikd/authdb"
	"github.com/go-kivik/kivik/v4/x/kivikd/logger"
)

func TestLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	l := logger.New(buf)
	mw := loggerMiddleware(l)
	req := httptest.NewRequest("GET", "/foo", nil)
	session := &auth.Session{User: &authdb.UserContext{Name: "bob"}}
	ctx := context.WithValue(req.Context(), SessionKey, &session)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Go away!"))
	})
	mw(handler).ServeHTTP(w, req)
	expectedRE := regexp.MustCompile(`^192\.0\.2\.1 bob \[\d\d\d\d-\d\d-\d\d \d\d:\d\d:\d\d.+] ` +
		`\([\d.]+[µn]s\) "GET /foo HTTP/1.1" 401 8 "" ""`)
	if !expectedRE.Match(buf.Bytes()) {
		t.Errorf("Log does not match. Got:\n%s\n", buf.String())
	}
}

func TestLoggerNoAuth(t *testing.T) {
	buf := &bytes.Buffer{}
	l := logger.New(buf)
	mw := loggerMiddleware(l)
	req := httptest.NewRequest("GET", "/foo", nil)
	session := &auth.Session{}
	ctx := context.WithValue(req.Context(), SessionKey, &session)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Go away!"))
	})
	mw(handler).ServeHTTP(w, req)
	expectedRE := regexp.MustCompile(`^192\.0\.2\.1  \[\d\d\d\d-\d\d-\d\d \d\d:\d\d:\d\d.+] ` +
		`\([\d.]+[µn]s\) "GET /foo HTTP/1.1" 401 8 "" ""`)
	if !expectedRE.Match(buf.Bytes()) {
		t.Errorf("Log does not match. Got:\n%s\n", buf.String())
	}
}
