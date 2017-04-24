package serve

import (
	"encoding/json"
	"net/http"
)

func flush(w http.ResponseWriter, r *http.Request) error {
	params := getParams(r)
	client := getClient(r)
	db, err := client.DB(r.Context(), params["db"])
	if err != nil {
		return err
	}
	if err = db.Flush(r.Context()); err != nil {
		return err
	}
	w.Header().Set("Content-Type", typeJSON)
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"instance_start_time": 0,
		"ok": true,
	})
}
