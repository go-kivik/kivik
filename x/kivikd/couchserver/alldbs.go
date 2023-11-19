package couchserver

import (
	"encoding/json"
	"net/http"
)

// GetAllDBs handles GET /_all_dbs
func (h *Handler) GetAllDBs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allDBs, err := h.client.AllDBs(r.Context())
		if err != nil {
			h.HandleError(w, err)
			return
		}
		w.Header().Set("Content-Type", typeJSON)
		h.HandleError(w, json.NewEncoder(w).Encode(allDBs))
	}
}
