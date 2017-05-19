package couchserver

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFavicoDefault(t *testing.T) {
	h := &Handler{}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/favicon.ico", nil)
	handler := h.GetFavicon()
	handler(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	buf := &bytes.Buffer{}
	buf.ReadFrom(resp.Body)
	expected, err := Asset("favicon.ico")
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Unexpected file contents. %d bytes instead of %d", buf.Len(), len(expected))
	}
}

func TestFavicoNotFound(t *testing.T) {
	h := &Handler{Favicon: "some/invalid/path"}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/favicon.ico", nil)
	handler := h.GetFavicon()
	handler(w, req)
	resp := w.Result()
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status: %d, actual: %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestFavicoExternalFile(t *testing.T) {
	h := &Handler{Favicon: "./files/favicon.ico"}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/favicon.ico", nil)
	handler := h.GetFavicon()
	handler(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	buf := &bytes.Buffer{}
	buf.ReadFrom(resp.Body)
	expected, err := Asset("favicon.ico")
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Unexpected file contents. %d bytes instead of %d", buf.Len(), len(expected))
	}
}
