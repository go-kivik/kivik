package couchserver

import (
	"net/http/httptest"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	_ "github.com/go-kivik/kivik/v4/x/memorydb"
)

func TestAllDBs(t *testing.T) {
	client, err := kivik.New("memory", "")
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
	if d := testy.DiffAsJSON(expected, resp.Body); d != nil {
		t.Error(d)
	}
}
