package couchserver

import (
	"net/http/httptest"
	"testing"

	"github.com/flimzy/diff"
)

func TestGetRoot(t *testing.T) {
	h := Handler{
		CompatVersion: "1.6.1",
		Vendor:        "Acme",
		VendorVersion: "10.0",
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	handler := h.GetRoot()
	handler(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	expected := map[string]interface{}{
		"couchdb": "Välkommen",
		"version": "1.6.1",
		"vendor": map[string]string{
			"version": "10.0",
			"name":    "Acme",
		},
	}
	if d := diff.AsJSON(expected, resp.Body); d != nil {
		t.Error(d)
	}
}
