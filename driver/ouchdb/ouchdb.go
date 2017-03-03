// Package ouchdb contains code that is shared between the CouchDB and PouchDB
// drivers. This package is not intended for public consumption. No attempt is
// made to retain forward compatibility with any symbols exported from this
// package.
package ouchdb

import (
	"encoding/json"
	"io"
)

// AllDocsResponse represents a response to a /{db}/_all_docs query.
type AllDocsResponse struct {
	Offset    int         `json:"offset"`
	TotalRows int         `json:"total_rows"`
	UpdateSeq string      `json:"update_seq,omitempty"`
	Rows      interface{} `json:"rows"`
}

// AllDocs parses an _all_docs response.
func AllDocs(body io.Reader, docs interface{}) (offset, totalRows int, updateSeq string, err error) {
	var result struct {
		AllDocsResponse
		// For decoding, we have to override this
		Rows json.RawMessage `json:"rows"`
	}
	dec := json.NewDecoder(body)
	if err := dec.Decode(&result); err != nil {
		return 0, 0, "", err
	}
	return result.Offset, result.TotalRows, result.UpdateSeq, json.Unmarshal(result.Rows, docs)
}

// ExtractRev extracts just the _rev JSON key from a JSON blob.
func ExtractRev(doc []byte) (string, error) {
	var result struct {
		Rev string `json:"_rev"`
	}
	err := json.Unmarshal(doc, &result)
	return result.Rev, err
}
