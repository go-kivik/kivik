package couchserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pressly/chi"
)

func TestDB(t *testing.T) {
	router := chi.NewRouter()
	var result string
	router.Get("/:db", func(_ http.ResponseWriter, r *http.Request) {
		result = DB(r)
	})
	req := httptest.NewRequest("GET", "/foo", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if result != "foo" {
		t.Errorf("Expected '%s', Got '%s'", "foo", result)
	}
}
