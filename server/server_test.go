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

package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/go-kivik/kivik/v4"
	_ "github.com/go-kivik/kivik/v4/x/fsdb" // Filesystem driver
)

func toJSON(i interface{}) ([]byte, error) {
	switch t := i.(type) {
	case string:
		return []byte(t), nil
	case json.RawMessage:
		return t, nil
	case []byte:
		return t, nil
	case io.Reader:
		return io.ReadAll(t)
	}
	return json.Marshal(i)
}

func diffAsJSON(want, got interface{}) string {
	w, err := toJSON(want)
	if err != nil {
		return fmt.Sprintf("[failed to read want: %s]", err)
	}
	g, err := toJSON(got)
	if err != nil {
		return fmt.Sprintf("[failed to read get: %s]", err)
	}
	var wObj, gObj interface{}
	_ = json.Unmarshal(w, &wObj)
	_ = json.Unmarshal(g, &gObj)
	return cmp.Diff(wObj, gObj)
}

func TestServer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		client     *kivik.Client
		method     string
		path       string
		body       io.Reader
		wantStatus int
		wantJSON   interface{}
	}{
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
			wantStatus: http.StatusNotImplemented,
			wantJSON: map[string]interface{}{
				"error":  "not_implemented",
				"reason": "Feature not implemented",
			},
		},
		{
			name:       "all dbs",
			method:     http.MethodGet,
			path:       "/_all_dbs",
			wantStatus: http.StatusOK,
			wantJSON:   []string{"db1", "db2"},
		},
		{
			name:       "all dbs, descending",
			method:     http.MethodGet,
			path:       "/_all_dbs?descending=true",
			wantStatus: http.StatusOK,
			wantJSON:   []string{"db2", "db1"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client, err := kivik.New("fs", "testdata/fsdb")
			if err != nil {
				t.Fatal(err)
			}
			s := New(client)
			req, err := http.NewRequest(tt.method, tt.path, tt.body)
			if err != nil {
				t.Fatal(err)
			}

			rec := httptest.NewRecorder()
			s.ServeHTTP(rec, req)

			res := rec.Result()
			if res.StatusCode != tt.wantStatus {
				t.Errorf("Unexpected response status: %d", res.StatusCode)
			}
			if d := diffAsJSON(tt.wantJSON, res.Body); d != "" {
				t.Error(d)
			}
		})
	}
}
