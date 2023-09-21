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

package cmd

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/xkivik/v4/cmd/kivik/errors"
)

const methodCopy = "COPY"

func Test_copy_RunE(t *testing.T) {
	tests := testy.NewTable()

	tests.Add("missing dsn", cmdTest{
		args:   []string{"copy"},
		status: errors.ErrUsage,
	})
	tests.Add("missing target", cmdTest{
		args:   []string{"copy", "http://example.com/foo/bar"},
		status: errors.ErrUsage,
	})
	tests.Add("invalid target", cmdTest{
		args:   []string{"copy", "http://example.com/foo/bar", "%xx"},
		status: errors.ErrUsage,
	})
	tests.Add("remote COPY", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Header: http.Header{
				"ETag": {`"2-62e778c9ec09214dd685a981dcc24074"`},
			},
			Body: io.NopCloser(strings.NewReader(`{"id": "target","ok": true,"rev": "2-62e778c9ec09214dd685a981dcc24074"}`)),
		}, func(t *testing.T, req *http.Request) {
			if req.Method != methodCopy {
				t.Errorf("Unexpected method: %v", req.Method)
			}
			if req.URL.Path != "/jjj/src" {
				t.Errorf("Unexpected path: %s", req.URL.Path)
			}
			if d := testy.DiffHTTPRequest(testy.Snapshot(t), req, standardReplacements...); d != nil {
				t.Error(d)
			}
		})

		return cmdTest{
			args: []string{"--debug", "copy", s.URL + "/jjj/src", "target"},
		}
	})
	tests.Add("emulated COPY", func(t *testing.T) interface{} {
		ss := testy.ServeResponseValidator(t, &http.Response{
			Header: http.Header{
				"Content-Type": {"application/json"},
				"ETag":         {`"2-62e778c9ec09214dd685a981dcc24074"`},
			},
			Body: io.NopCloser(strings.NewReader(`{"id": "target","ok": true,"rev": "2-62e778c9ec09214dd685a981dcc24074"}`)),
		}, func(t *testing.T, req *http.Request) {
			if req.Method != http.MethodGet {
				t.Errorf("Unexpected source method: %v", req.Method)
			}
			if req.URL.Path != "/asdf/src" {
				t.Errorf("Unexpected source path: %s", req.URL.Path)
			}
		})
		ts := testy.ServeResponseValidator(t, &http.Response{
			Header: http.Header{
				"ETag": {`"2-62e778c9ec09214dd685a981dcc24074"`},
			},
			Body: io.NopCloser(strings.NewReader(`{"id": "target","ok": true,"rev": "2-62e778c9ec09214dd685a981dcc24074"}`)),
		}, gunzip(func(t *testing.T, req *http.Request) {
			if req.Method != http.MethodPut {
				t.Errorf("Unexpected target method: %v", req.Method)
			}
			if req.URL.Path != "/qwerty/target" {
				t.Errorf("Unexpected target path: %s", req.URL.Path)
			}
			if d := testy.DiffHTTPRequest(testy.Snapshot(t), req, standardReplacements...); d != nil {
				t.Error(d)
			}
		}))

		return cmdTest{
			args: []string{"--debug", "copy", ss.URL + "/asdf/src", ts.URL + "/qwerty/target"},
		}
	})
	tests.Add("remote COPY with rev", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Header: http.Header{
				"ETag": {`"2-62e778c9ec09214dd685a981dcc24074"`},
			},
			Body: io.NopCloser(strings.NewReader(`{"id": "target","ok": true,"rev": "2-62e778c9ec09214dd685a981dcc24074"}`)),
		}, func(t *testing.T, req *http.Request) {
			if req.Method != methodCopy {
				t.Errorf("Unexpected method: %v", req.Method)
			}
			if req.URL.Path != "/jkl/src" {
				t.Errorf("Unexpected path: %s", req.URL.Path)
			}
			if d := testy.DiffHTTPRequest(testy.Snapshot(t), req, standardReplacements...); d != nil {
				t.Error(d)
			}
		})

		return cmdTest{
			args: []string{"--debug", "copy", s.URL + "/jkl/src", "target?rev=3-xxx"},
		}
	})
	tests.Add("remote COPY with --target-rev", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Header: http.Header{
				"ETag": {`"2-62e778c9ec09214dd685a981dcc24074"`},
			},
			Body: io.NopCloser(strings.NewReader(`{"id": "target","ok": true,"rev": "2-62e778c9ec09214dd685a981dcc24074"}`)),
		}, func(t *testing.T, req *http.Request) {
			if req.Method != methodCopy {
				t.Errorf("Unexpected method: %v", req.Method)
			}
			if req.URL.Path != "/jkl/src" {
				t.Errorf("Unexpected path: %s", req.URL.Path)
			}
			if d := testy.DiffHTTPRequest(testy.Snapshot(t), req, standardReplacements...); d != nil {
				t.Error(d)
			}
		})

		return cmdTest{
			args: []string{"--debug", "copy", s.URL + "/jkl/src", "target", "--target-rev", "3-lkjds"},
		}
	})
	tests.Add("emulated COPY with rev", func(t *testing.T) interface{} {
		ss := testy.ServeResponseValidator(t, &http.Response{
			Header: http.Header{
				"Content-Type": {"application/json"},
				"ETag":         {`"2-62e778c9ec09214dd685a981dcc24074"`},
			},
			Body: io.NopCloser(strings.NewReader(`{"id": "target","ok": true,"rev": "2-62e778c9ec09214dd685a981dcc24074"}`)),
		}, func(t *testing.T, req *http.Request) {
			if req.Method != http.MethodGet {
				t.Errorf("Unexpected source method: %v", req.Method)
			}
			if req.URL.Path != "/asdf/src" {
				t.Errorf("Unexpected source path: %s", req.URL.Path)
			}
		})
		ts := testy.ServeResponseValidator(t, &http.Response{
			Header: http.Header{
				"ETag": {`"2-62e778c9ec09214dd685a981dcc24074"`},
			},
			Body: io.NopCloser(strings.NewReader(`{"id": "target","ok": true,"rev": "2-62e778c9ec09214dd685a981dcc24074"}`)),
		}, gunzip(func(t *testing.T, req *http.Request) {
			if req.Method != http.MethodPut {
				t.Errorf("Unexpected target method: %v", req.Method)
			}
			if req.URL.Path != "/qwerty/target" {
				t.Errorf("Unexpected target path: %s", req.URL.Path)
			}
			if d := testy.DiffHTTPRequest(testy.Snapshot(t), req, standardReplacements...); d != nil {
				t.Error(d)
			}
		}))

		return cmdTest{
			args: []string{"--debug", "copy", ss.URL + "/asdf/src", ts.URL + "/qwerty/target?rev=7-qhk"},
		}
	})
	tests.Add("from context", func(t *testing.T) interface{} {
		return cmdTest{
			args:   []string{"--debug", "copy", "--config", "./testdata/copy.yaml", "--target", "target"},
			status: errors.ErrUnavailable,
		}
	})

	tests.Run(t, func(t *testing.T, tt cmdTest) {
		tt.Test(t)
	})
}
