package couchserver

import (
	"net/http"

	"github.com/go-chi/chi"
)

// DB returns the db name in this request, or "" if none.
func DB(r *http.Request) string {
	return chi.URLParam(r, "db")
}
