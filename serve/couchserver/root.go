package couchserver

import (
	"encoding/json"
	"net/http"
)

// GetRoot handles requests for: GET /
func (h *Handler) GetRoot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		compatVer, vendName, vendVers := h.vendor()
		w.Header().Set("Content-Type", typeJSON)
		h.HandleError(w, json.NewEncoder(w).Encode(serverInfo{
			CouchDB: "Välkommen",
			Version: compatVer,
			Vendor: vendorInfo{
				Name:    vendName,
				Version: vendVers,
			},
		}))
	}
}
