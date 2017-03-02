package common

import (
	"encoding/json"
	"io"
)

// AllDocs parses an _all_docs response.
func AllDocs(body io.Reader, docs interface{}) (offset, totalRows int, err error) {
	var result struct {
		Offset    int             `json:"offset"`
		TotalRows int             `json:"total_rows"`
		Rows      json.RawMessage `json:"rows"`
	}
	dec := json.NewDecoder(body)
	if err := dec.Decode(&result); err != nil {
		return 0, 0, err
	}
	return result.Offset, result.TotalRows, json.Unmarshal(result.Rows, docs)
}
