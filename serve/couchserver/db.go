package couchserver

import (
	"encoding/json"
	"net/http"
)

// PutDB handles PUT /{db}
func (h *Handler) PutDB() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h.client.CreateDB(r.Context(), DB(r)); err != nil {
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
		exists, err := h.client.DBExists(r.Context(), DB(r))
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

// GetDB handles GET /{db}
func (h *Handler) GetDB() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db, err := h.client.DB(r.Context(), DB(r))
		if err != nil {
			h.HandleError(w, err)
			return
		}
		stats, err := db.Stats(r.Context())
		if err != nil {
			h.HandleError(w, err)
			return
		}
		w.Header().Set("Cache-Control", "must-revalidate")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(stats)
		if err != nil {
			h.HandleError(w, err)
			return
		}

	}
}

// Flush handles POST /{db}/_ensure_full_commit
func (h *Handler) Flush() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db, err := h.client.DB(r.Context(), DB(r))
		if err != nil {
			h.HandleError(w, err)
			return
		}
		if err := db.Flush(r.Context()); err != nil {
			h.HandleError(w, err)
			return
		}
		w.Header().Set("Content-Type", typeJSON)
		h.HandleError(w, json.NewEncoder(w).Encode(map[string]interface{}{
			"instance_start_time": 0,
			"ok": true,
		}))
	}
}
