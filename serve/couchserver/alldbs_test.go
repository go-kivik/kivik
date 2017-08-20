package couchserver

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
	_ "github.com/go-kivik/memorydb"
)

func TestAllDBs(t *testing.T) {
	client, err := kivik.New(context.Background(), "memory", "")
	if err != nil {
		panic(err)
	}
	h := &Handler{client: &clientWrapper{client}}
	handler := h.GetAllDBs()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/_all_dbs", nil)
	handler(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	expected := []string{}
	if d := diff.AsJSON(expected, resp.Body); d != nil {
		t.Error(d)
	}
}
