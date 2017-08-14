package couchserver

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
)

type mockCreator struct {
	backend
}

func (p *mockCreator) CreateDB(_ context.Context, _ string, _ ...kivik.Options) error {
	return nil
}

type errCreator struct {
	backend
}

func (p *errCreator) CreateDB(_ context.Context, _ string, _ ...kivik.Options) error {
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
		if d := diff.AsJSON(expected, resp.Body); d != nil {
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

func (d *mockNotExists) DBExists(_ context.Context, _ string, _ ...kivik.Options) (bool, error) {
	return false, nil
}

type mockExists struct{ backend }

func (d *mockExists) DBExists(_ context.Context, _ string, _ ...kivik.Options) (bool, error) {
	return true, nil
}

type mockErrExists struct{ backend }

func (d *mockErrExists) DBExists(_ context.Context, _ string, _ ...kivik.Options) (bool, error) {
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

func (c *mockGetFound) DB(_ context.Context, _ string, _ ...kivik.Options) (db, error) {
	return &mockFoundDB{}, nil
}

type mockGetNotFound struct{ backend }

func (c *mockGetNotFound) DB(_ context.Context, _ string, _ ...kivik.Options) (db, error) {
	return nil, errors.Status(http.StatusNotFound, "database not found")
}

type errClient struct{ backend }

func (c *errClient) DB(_ context.Context, _ string, _ ...kivik.Options) (db, error) {
	return nil, errors.New("failure")
}

func TestGetDB(t *testing.T) {
	t.Run("Endpoint exists for GET", func(t *testing.T) {
		h := &Handler{client: &mockGetNotFound{}}
		resp := callEndpointEndClose(h, "GET", "/exists")
		if resp.StatusCode == http.StatusMethodNotAllowed {
			t.Error("Expected another response than method not allowed")
		}
	})

	t.Run("Not found", func(t *testing.T) {
		h := &Handler{client: &mockGetNotFound{}}
		resp := callEndpointEndClose(h, "GET", "/notexists")
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404, got %s", resp.Status)
		}
	})

	t.Run("Found", func(t *testing.T) {
		h := &Handler{client: &mockGetFound{}}
		resp := callEndpointEndClose(h, "GET", "/asdf")
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %s", resp.Status)
		}
	})

	t.Run("Error", func(t *testing.T) {
		h2 := &Handler{client: &errClient{}}
		resp := callEndpointEndClose(h2, "GET", "/error")
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected 500, got %s", resp.Status)
		}
	})

	t.Run("Response", func(t *testing.T) {
		h := &Handler{client: &mockGetFound{}}
		resp := callEndpoint(h, "GET", "/asdf")
		if cc := resp.Header.Get("Cache-Control"); cc != "must-revalidate" {
			t.Errorf("Cache-Control header doesn't match, got %s", cc)
		}
		if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
			t.Errorf("Content-Type header doesn't match, got %s", contentType)
		}
		var body interface{}
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if err = json.Unmarshal(buf, &body); err != nil {
			t.Errorf("JSON error, %s", err)
		}
		expected := testStats
		if difftext := diff.AsJSON(expected, body); difftext != nil {
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

func callEndpointEndClose(h *Handler, method string, path string) *http.Response {
	resp := callEndpoint(h, method, path)
	resp.Body.Close()
	return resp
}
