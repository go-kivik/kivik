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

func TestPutDB(t *testing.T) {
	_, h := createClient(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/foo", nil)
	handler := h.Main()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	expected := map[string]interface{}{
		"ok": true,
	}
	if d := diff.AsJSON(expected, resp.Body); d != "" {
		t.Error(d)
	}
}

func TestHeadDB(t *testing.T) {
	client, h := createClient(t)
	t.Run("NotExists", func(t *testing.T) {
		resp := callEndpoint(h, "HEAD", "/notexist")
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404/NotFound, got %s", resp.Status)
		}
	})
	t.Run("Exists", func(t *testing.T) {
		if err := client.CreateDB(context.Background(), "exists"); err != nil {
			t.Fatal(err)
		}
		resp := callEndpoint(h, "HEAD", "/exists")
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200/OK, got %s", resp.Status)
		}
	})
}

type errClient struct {
	*kivik.Client
}

func (c *errClient) DBExists(_ context.Context, _ string, _ ...kivik.Options) (bool, error) {
	return false, errors.New("aaah!")
}

func TestGetDB(t *testing.T) {
	client, h := createClient(t)
	t.Run("Endpoint exists for GET", func(t *testing.T) {
		resp := callEndpoint(h, "GET", "/exists")
		if resp.StatusCode == http.StatusMethodNotAllowed {
			t.Error("Expected another response than method not allowed")
		}
	})

	t.Run("Not found", func(t *testing.T) {
		resp := callEndpoint(h, "GET", "/notexists")
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404, got %s", resp.Status)
		}
	})

	t.Run("Found", func(t *testing.T) {
		path := "exists"
		if err := client.CreateDB(context.Background(), path); err != nil {
			t.Fatal(err)
		}
		resp := callEndpoint(h, "GET", "/"+path)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %s", resp.Status)
		}
	})

	t.Run("Error", func(t *testing.T) {
		myErrClient := errClient{client}

		h2 := &Handler{Client: &myErrClient}
		resp := callEndpoint(h2, "GET", "/error")
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected 500, got %s", resp.Status)
		}
	})
}

func TestFlush(t *testing.T) {
	// TODO
}

func createClient(t *testing.T) (*kivik.Client, *Handler) {
	client, err := kivik.New(context.Background(), "memory", "")
	if err != nil {
		t.Fatal(err)
	}
	h := &Handler{Client: client}
	return client, h
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
