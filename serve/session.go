package serve

import (
	"encoding/json"
	"net/http"
)

func getSession(w http.ResponseWriter, r *http.Request) error {
	session := MustGetSession(r.Context())
	w.Header().Add("Content-Type", typeJSON)
	return json.NewEncoder(w).Encode(session)
}
