package couchserver

import (
	"encoding/json"
	"net/http"
)

// PutDB handles PUT /{db}
func (h *Handler) PutDB() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := DB(r)
		if err := h.Client.CreateDB(r.Context(), db); err != nil {
			h.HandleError(w, err)
			return
		}
		h.HandleError(w, json.NewEncoder(w).Encode(map[string]interface{}{
			"ok": true,
		}))
	}
}
