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
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/xkivik/v4/cmd/kivik/errors"
)

func Test_put_RunE(t *testing.T) {
	tests := testy.NewTable()

	tests.Add("missing document", cmdTest{
		args:   []string{"put"},
		status: errors.ErrUsage,
	})
	tests.Add("full url on command line", cmdTest{
		args:   []string{"--debug", "put", "http://localhost:1/foo/bar", "-d", "{}"},
		status: errors.ErrUnavailable,
	})
	tests.Add("json data string", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Body: io.NopCloser(strings.NewReader(`{"ok":true,"rev":"1-xxx"}`)),
		}, gunzip(func(t *testing.T, req *http.Request) {
			defer req.Body.Close() // nolint:errcheck
			if d := testy.DiffAsJSON(testy.Snapshot(t), req.Body); d != nil {
				t.Error(d)
			}
		}))

		return cmdTest{
			args: []string{"--debug", "put", s.URL + "/foo/bar", "--data", `{"foo":"bar"}`},
		}
	})
	tests.Add("json data stdin", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Body: io.NopCloser(strings.NewReader(`{"ok":true,"rev":"1-xxx"}`)),
		}, gunzip(func(t *testing.T, req *http.Request) {
			defer req.Body.Close() // nolint:errcheck
			if d := testy.DiffAsJSON(testy.Snapshot(t), req.Body); d != nil {
				t.Error(d)
			}
		}))

		return cmdTest{
			args:  []string{"--debug", "put", s.URL + "/foo/bar", "--data-file", "-"},
			stdin: `{"foo":"bar"}`,
		}
	})
	tests.Add("json data file", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Body: io.NopCloser(strings.NewReader(`{"ok":true,"rev":"1-xxx"}`)),
		}, gunzip(func(t *testing.T, req *http.Request) {
			defer req.Body.Close() // nolint:errcheck
			if d := testy.DiffAsJSON(testy.Snapshot(t), req.Body); d != nil {
				t.Error(d)
			}
		}))

		return cmdTest{
			args:  []string{"--debug", "put", s.URL + "/foo/bar", "--data-file", "./testdata/doc.json"},
			stdin: `{"foo":"bar"}`,
		}
	})
	tests.Add("yaml data string", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Body: io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
		}, gunzip(func(t *testing.T, req *http.Request) {
			defer req.Body.Close() // nolint:errcheck
			if d := testy.DiffAsJSON(testy.Snapshot(t), req.Body); d != nil {
				t.Error(d)
			}
		}))

		return cmdTest{
			args: []string{"--debug", "put", s.URL + "/foo/bar", "--yaml", "--data", `foo: bar`},
		}
	})
	tests.Add("auto put config", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(strings.NewReader(`"old"`)),
		}, gunzip(func(t *testing.T, req *http.Request) {
			content, _ := io.ReadAll(req.Body)
			if string(content) != `"baz"` {
				t.Errorf("Unexpected request body: %s", string(content))
			}
			if req.Method != http.MethodPut {
				t.Errorf("Unexpected method: %s", req.Method)
			}
			if req.URL.Path != "/_node/_local/_config/foo/bar" {
				t.Errorf("unexpected path: %s", req.URL.Path)
			}
		}))

		return cmdTest{
			args: []string{"put", s.URL + "/_node/_local/_config/foo/bar", "-d", "baz"},
		}
	})
	tests.Add("auto put security", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(strings.NewReader(`"old"`)),
		}, gunzip(func(t *testing.T, req *http.Request) {
			if req.Method != http.MethodPut {
				t.Errorf("Unexpected method: %s", req.Method)
			}
			want := json.RawMessage(`{}`)
			if d := testy.DiffAsJSON(want, req.Body); d != nil {
				t.Errorf("Unexpected request body: %s", d)
			}
			if req.URL.Path != "/foo/_security" {
				t.Errorf("unexpected path: %s", req.URL.Path)
			}
		}))

		return cmdTest{
			args: []string{"put", s.URL + "/foo/_security", "-d", "{}"},
		}
	})

	tests.Run(t, func(t *testing.T, tt cmdTest) {
		tt.Test(t)
	})
}
