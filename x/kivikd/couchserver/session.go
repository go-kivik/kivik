package couchserver

import (
	"encoding/json"
	"net/http"

	"github.com/go-kivik/kivikd/v4/auth"
)

// GetSession serves GET /_session
func (h *Handler) GetSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, ok := r.Context().Value(h.SessionKey).(**auth.Session)
		if !ok {
			panic("No session!")
		}
		w.Header().Add("Content-Type", typeJSON)
		h.HandleError(w, json.NewEncoder(w).Encode(s))
	}
}
