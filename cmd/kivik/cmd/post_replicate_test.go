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
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/xkivik/v4/cmd/kivik/errors"
)

func Test_post_replicate_RunE(t *testing.T) {
	tests := testy.NewTable()

	tests.Add("missing dsn", cmdTest{
		args:   []string{"post", "replicate"},
		status: errors.ErrUsage,
	})
	tests.Add("no source or target", cmdTest{
		args:   []string{"post", "replicate", "http://example.com/db"},
		status: errors.ErrUsage,
	})
	tests.Add("full dsns", func(t *testing.T) interface{} {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodHead {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if r.Method != http.MethodPost {
				t.Errorf("Unexpected method: %s", r.Method)
			}
			if d := testy.DiffAsJSON(testy.Snapshot(t), gunzipBody(t, r.Body)); d != nil {
				t.Error(d)
			}
			_, _ = w.Write([]byte(`{"ok":true,"session_id":"87bf1c2a565f20976c4cb19a22528b7e","source_last_seq":"6-g1AAAABteJzLYWBgYMpgTmHgzcvPy09JdcjLz8gvLskBCScyJNX___8_K4M5kS0XKMBunmRiYmRmhK4Yh_Y8FiDJ0ACk_oNMSWTIAgDY6SGt","replication_id_version":4,"history":[{"session_id":"87bf1c2a565f20976c4cb19a22528b7e","start_time":"Sun, 25 Apr 2021 19:53:34 GMT","end_time":"Sun, 25 Apr 2021 19:53:35 GMT","start_last_seq":0,"end_last_seq":"6-g1AAAABteJzLYWBgYMpgTmHgzcvPy09JdcjLz8gvLskBCScyJNX___8_K4M5kS0XKMBunmRiYmRmhK4Yh_Y8FiDJ0ACk_oNMSWTIAgDY6SGt","recorded_seq":"6-g1AAAABteJzLYWBgYMpgTmHgzcvPy09JdcjLz8gvLskBCScyJNX___8_K4M5kS0XKMBunmRiYmRmhK4Yh_Y8FiDJ0ACk_oNMSWTIAgDY6SGt","missing_checked":2,"missing_found":2,"docs_read":2,"docs_written":2,"doc_write_failures":0}]}
			`))
		}))

		return cmdTest{
			args: []string{"--debug", "post", "replicate", s.URL, "-O", "source=http://example.com/foo", "-O", "target=http://example.com/bar"},
		}
	})
	tests.Add("objects", func(t *testing.T) interface{} {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodHead {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if r.Method != http.MethodPost {
				t.Errorf("Unexpected method: %s", r.Method)
			}
			if d := testy.DiffAsJSON(testy.Snapshot(t), gunzipBody(t, r.Body)); d != nil {
				t.Error(d)
			}
			_, _ = w.Write([]byte(`{"ok":true}`))
		}))

		return cmdTest{
			args: []string{"--debug", "post", "replicate", s.URL, "-O", `source={"url":"http://example.com/foo"}`, "-O", `target={"url":"http://example.com/bar"}`},
		}
	})
	tests.Add("options", func(t *testing.T) interface{} {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodHead {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if r.Method != http.MethodPost {
				t.Errorf("Unexpected method: %s", r.Method)
			}
			if d := testy.DiffAsJSON(testy.Snapshot(t), gunzipBody(t, r.Body)); d != nil {
				t.Error(d)
			}
			_, _ = w.Write([]byte(`{"ok":true}`))
		}))

		return cmdTest{
			args: []string{"--debug", "post", "replicate", s.URL, "-O", "source=http://example.com/foo", "-O", "target=http://example.com/bar", "-B", "cancel=true", "-B", "continuous=true", "-B", "create_target=true", "-O", "doc_ids=foo,bar", "-O", "filter=oink", "-O", "source_proxy=http://localhost:9999/", "-O", "target_proxy=http://localhost:1111/"},
		}
	})

	tests.Run(t, func(t *testing.T, tt cmdTest) {
		tt.Test(t)
	})
}
