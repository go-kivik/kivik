package couchserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
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
		h := &Handler{Client: &mockCreator{}}
		resp := callEndpoint(h, "PUT", "/foo")
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %s", resp.Status)
		}
		expected := map[string]interface{}{
			"ok": true,
		}
		if d := diff.AsJSON(expected, resp.Body); d != "" {
			t.Error(d)
		}
	})
	t.Run("Error", func(t *testing.T) {
		h := &Handler{Client: &errCreator{}}
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
		h := &Handler{Client: &mockNotExists{}}
		resp := callEndpoint(h, "HEAD", "/notexist")
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404/NotFound, got %s", resp.Status)
		}
	})

	t.Run("Exists", func(t *testing.T) {
		h := &Handler{Client: &mockExists{}}
		resp := callEndpoint(h, "HEAD", "/exists")
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200/OK, got %s", resp.Status)
		}
	})

	t.Run("Error", func(t *testing.T) {
		h := &Handler{Client: &mockErrExists{}}
		resp := callEndpoint(h, "HEAD", "/exists")
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected 500, got %s", resp.Status)
		}
	})
}

type mockGetFound struct{ backend }

func (c *mockGetFound) DBExists(_ context.Context, _ string, _ ...kivik.Options) (bool, error) {
	return true, nil
}

type mockGetNotFound struct{ backend }

func (c *mockGetNotFound) DBExists(_ context.Context, _ string, _ ...kivik.Options) (bool, error) {
	return false, nil
}

type errClient struct{ backend }

func (c *errClient) DBExists(_ context.Context, _ string, _ ...kivik.Options) (bool, error) {
	return false, errors.New("failure")
}

func TestGetDB(t *testing.T) {
	t.Run("Endpoint exists for GET", func(t *testing.T) {
		h := &Handler{Client: &mockGetNotFound{}}
		resp := callEndpoint(h, "GET", "/exists")
		if resp.StatusCode == http.StatusMethodNotAllowed {
			t.Error("Expected another response than method not allowed")
		}
	})

	t.Run("Not found", func(t *testing.T) {
		h := &Handler{Client: &mockGetNotFound{}}
		resp := callEndpoint(h, "GET", "/notexists")
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404, got %s", resp.Status)
		}
	})

	t.Run("Found", func(t *testing.T) {
		h := &Handler{Client: &mockGetFound{}}
		resp := callEndpoint(h, "GET", "/asdf")
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %s", resp.Status)
		}
	})

	t.Run("Error", func(t *testing.T) {
		h2 := &Handler{Client: &errClient{}}
		resp := callEndpoint(h2, "GET", "/error")
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected 500, got %s", resp.Status)
		}
	})
}

func callEndpoint(h *Handler, method string, path string) *http.Response {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	handler := h.Main()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	resp.Body.Close()
	return resp
}
