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

func Test_post_purge_RunE(t *testing.T) {
	tests := testy.NewTable()

	tests.Add("missing dsn", cmdTest{
		args:   []string{"post", "purge"},
		status: errors.ErrUsage,
	})
	tests.Add("one rev", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Body: io.NopCloser(strings.NewReader(`{"ok":true}`)),
		}, gunzip(func(t *testing.T, req *http.Request) {
			if req.Method != http.MethodPost {
				t.Errorf("Unexpected method: %v", req.Method)
			}
			if req.URL.Path != "/qwerty/_purge" {
				t.Errorf("Unexpected path: %s", req.URL.Path)
			}
			if d := testy.DiffAsJSON(testy.Snapshot(t), req.Body); d != nil {
				t.Error(d)
			}
		}))

		return cmdTest{
			args: []string{"post", "purge", s.URL + "/qwerty/doc", "--revs", "1-xxx"},
		}
	})
	tests.Add("two revs", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Body: io.NopCloser(strings.NewReader(`{"ok":true}`)),
		}, gunzip(func(t *testing.T, req *http.Request) {
			if req.Method != http.MethodPost {
				t.Errorf("Unexpected method: %v", req.Method)
			}
			if req.URL.Path != "/xxx/_purge" {
				t.Errorf("Unexpected path: %s", req.URL.Path)
			}
			if d := testy.DiffAsJSON(testy.Snapshot(t), req.Body); d != nil {
				t.Error(d)
			}
		}))

		return cmdTest{
			args: []string{"post", "purge", s.URL + "/xxx/doc", "--revs", "1-xxx,2-xxx"},
		}
	})
	tests.Add("from --data", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Body: io.NopCloser(strings.NewReader(`{"ok":true}`)),
		}, gunzip(func(t *testing.T, req *http.Request) {
			if req.Method != http.MethodPost {
				t.Errorf("Unexpected method: %v", req.Method)
			}
			if req.URL.Path != "/db/_purge" {
				t.Errorf("Unexpected path: %s", req.URL.Path)
			}
			if d := testy.DiffAsJSON(testy.Snapshot(t), req.Body); d != nil {
				t.Error(d)
			}
		}))

		return cmdTest{
			args: []string{"post", "purge", s.URL + "/db", "--data", `{"doc":["1-xxx","2-yyy"]}`},
		}
	})
	tests.Add("from --data and dsn", cmdTest{
		args:   []string{"post", "purge", "http://example.com/db/foo", "--data", `{"doc":["1-xxx","2-yyy"]}`},
		status: errors.ErrUsage,
	})

	tests.Run(t, func(t *testing.T, tt cmdTest) {
		tt.Test(t)
	})
}
