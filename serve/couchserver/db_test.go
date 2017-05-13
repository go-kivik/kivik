package couchserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
)

func TestPutDB(t *testing.T) {
	client, err := kivik.New(context.Background(), "memory", "")
	if err != nil {
		panic(err)
	}
	h := &Handler{Client: client}
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
	client, err := kivik.New(context.Background(), "memory", "")
	if err != nil {
		panic(err)
	}
	h := &Handler{Client: client}
	t.Run("NotExists", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("HEAD", "/notexists", nil)
		handler := h.Main()
		handler.ServeHTTP(w, req)
		resp := w.Result()
		resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404/NotFound, got %s", resp.Status)
		}
	})
	t.Run("Exists", func(t *testing.T) {
		if err := client.CreateDB(context.Background(), "exists"); err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()
		req := httptest.NewRequest("HEAD", "/exists", nil)
		handler := h.Main()
		handler.ServeHTTP(w, req)
		resp := w.Result()
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200/OK, got %s", resp.Status)
		}
	})
}
