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

package couchserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	internal "github.com/go-kivik/kivik/v4/int/errors"
)

type mockCreator struct {
	backend
}

func (p *mockCreator) CreateDB(_ context.Context, _ string, _ ...kivik.Option) error {
	return nil
}

type errCreator struct {
	backend
}

func (p *errCreator) CreateDB(_ context.Context, _ string, _ ...kivik.Option) error {
	return errors.New("failure")
}

func TestPutDB(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		h := &Handler{client: &mockCreator{}}
		resp := callEndpoint(h, "PUT", "/foo")
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %s", resp.Status)
		}
		expected := map[string]interface{}{
			"ok": true,
		}
		if d := testy.DiffAsJSON(expected, resp.Body); d != nil {
			t.Error(d)
		}
	})
	t.Run("Error", func(t *testing.T) {
		h := &Handler{client: &errCreator{}}
		resp := callEndpoint(h, "PUT", "/foo")
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected 500, got %s", resp.Status)
		}
	})
}

type mockNotExists struct{ backend }

func (d *mockNotExists) DBExists(_ context.Context, _ string, _ ...kivik.Option) (bool, error) {
	return false, nil
}

type mockExists struct{ backend }

func (d *mockExists) DBExists(_ context.Context, _ string, _ ...kivik.Option) (bool, error) {
	return true, nil
}

type mockErrExists struct{ backend }

func (d *mockErrExists) DBExists(_ context.Context, _ string, _ ...kivik.Option) (bool, error) {
	return false, errors.New("failure")
}

func TestHeadDB(t *testing.T) {
	t.Run("NotExists", func(t *testing.T) {
		h := &Handler{client: &mockNotExists{}}
		resp := callEndpoint(h, "HEAD", "/notexist")
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404/NotFound, got %s", resp.Status)
		}
	})

	t.Run("Exists", func(t *testing.T) {
		h := &Handler{client: &mockExists{}}
		resp := callEndpoint(h, "HEAD", "/exists")
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200/OK, got %s", resp.Status)
		}
	})

	t.Run("Error", func(t *testing.T) {
		h := &Handler{client: &mockErrExists{}}
		resp := callEndpoint(h, "HEAD", "/exists")
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected 500, got %s", resp.Status)
		}
	})
}

type mockFoundDB struct{ db }

// expected := map[string]interface{}{
// 	"committed_update_seq": 292786,
// 	"compact_running":      false,
// 	"data_size":            65031503,
// 	"disk_format_version":  6,
// 	"instance_start_time":  "1376269325408900",
// 	"purge_seq":            1,
// }

var testStats = &kivik.DBStats{
	Name:           "receipts",
	CompactRunning: false,
	DocCount:       6146,
	DeletedCount:   64637,
	UpdateSeq:      "292786",
	DiskSize:       137433211,
	ActiveSize:     1,
}

func (d *mockFoundDB) Stats(_ context.Context) (*kivik.DBStats, error) {
	return testStats, nil
}

type mockGetFound struct{ backend }

func (c *mockGetFound) DB(_ context.Context, _ string, _ ...kivik.Option) (db, error) {
	return &mockFoundDB{}, nil
}

type mockGetNotFound struct{ backend }

func (c *mockGetNotFound) DB(_ context.Context, _ string, _ ...kivik.Option) (db, error) {
	return nil, &internal.Error{Status: http.StatusNotFound, Message: "database not found"}
}

type errClient struct{ backend }

func (c *errClient) DB(_ context.Context, _ string, _ ...kivik.Option) (db, error) {
	return nil, errors.New("failure")
}

func TestGetDB(t *testing.T) {
	t.Run("Endpoint exists for GET", func(t *testing.T) {
		h := &Handler{client: &mockGetNotFound{}}
		resp := callEndpointEndClose(h, "/exists")
		if resp.StatusCode == http.StatusMethodNotAllowed {
			t.Error("Expected another response than method not allowed")
		}
	})

	t.Run("Not found", func(t *testing.T) {
		h := &Handler{client: &mockGetNotFound{}}
		resp := callEndpointEndClose(h, "/notexists")
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404, got %s", resp.Status)
		}
	})

	t.Run("Found", func(t *testing.T) {
		h := &Handler{client: &mockGetFound{}}
		resp := callEndpointEndClose(h, "/asdf")
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %s", resp.Status)
		}
	})

	t.Run("Error", func(t *testing.T) {
		h2 := &Handler{client: &errClient{}}
		resp := callEndpointEndClose(h2, "/error")
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected 500, got %s", resp.Status)
		}
	})

	t.Run("Response", func(t *testing.T) {
		h := &Handler{client: &mockGetFound{}}
		resp := callEndpoint(h, "GET", "/asdf")
		t.Cleanup(func() {
			_ = resp.Body.Close()
		})
		if cc := resp.Header.Get("Cache-Control"); cc != "must-revalidate" {
			t.Errorf("Cache-Control header doesn't match, got %s", cc)
		}
		if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
			t.Errorf("Content-Type header doesn't match, got %s", contentType)
		}
		var body interface{}
		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(buf, &body); err != nil {
			t.Errorf("JSON error, %s", err)
		}
		expected := testStats
		if difftext := testy.DiffAsJSON(expected, body); difftext != nil {
			t.Error(difftext)
		}
	})
}

func callEndpoint(h *Handler, method string, path string) *http.Response {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	handler := h.Main()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	return resp
}

func callEndpointEndClose(h *Handler, path string) *http.Response {
	resp := callEndpoint(h, http.MethodGet, path)
	_ = resp.Body.Close()
	return resp
}
