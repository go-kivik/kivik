package serve

import (
	"encoding/json"
	"net/http"
	"time"
)

func flush(w http.ResponseWriter, r *http.Request) error {
	params := getParams(r)
	client := getClient(r)
	db, err := client.DB(params["db"])
	if err != nil {
		return err
	}
	ts, err := db.Flush()
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", typeJSON)
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"instance_start_time": int64(ts.Sub(time.Unix(0, 0)).Seconds() * 1e6),
		"ok": true,
	})
}
