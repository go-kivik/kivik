package couchserver

import (
	"context"
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
