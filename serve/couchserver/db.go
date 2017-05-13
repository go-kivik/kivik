package couchserver

import (
	"encoding/json"
	"net/http"
)

// PutDB handles PUT /{db}
func (h *Handler) PutDB() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h.Client.CreateDB(r.Context(), DB(r)); err != nil {
			h.HandleError(w, err)
			return
		}
		h.HandleError(w, json.NewEncoder(w).Encode(map[string]interface{}{
			"ok": true,
		}))
	}
}

// HeadDB handles HEAD /{db}
func (h *Handler) HeadDB() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exists, err := h.Client.DBExists(r.Context(), DB(r))
		if err != nil {
			h.HandleError(w, err)
			return
		}
		if exists {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}
