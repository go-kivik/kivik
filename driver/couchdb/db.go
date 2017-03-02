package couchdb

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type db struct {
	*client
	dbName string
}

func (d *db) path(path string) string {
	return d.dbName + "/" + path
}

// AllDocs returns all of the documents in the database.
func (d *db) AllDocs(docs interface{}, opts url.Values) (offset, totalrows int, err error) {
	var result struct {
		Offset    int             `json:"offset"`
		TotalRows int             `json:"total_rows"`
		Rows      json.RawMessage `json:"rows"`
	}
	err = d.client.newRequest(http.MethodGet, d.path("_all_docs")).
		AddHeader("Accept", typeJSON).
		DoJSON(&result)
	if err != nil {
		return 0, 0, err
	}
	return result.Offset, result.TotalRows, json.Unmarshal(result.Rows, docs)
}

// Get fetches the requested document.
func (d *db) Get(docID string, doc interface{}, opts url.Values) error {
	return d.client.newRequest(http.MethodGet, d.url(docID, opts)).
		AddHeader("Accept", typeJSON).
		AddHeader("Accept", typeMixed).
		DoJSON(doc)
}
