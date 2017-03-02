package couchdb

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/flimzy/kivik/driver/ouchdb"
)

type db struct {
	*client
	dbName string
}

func (d *db) path(path string) string {
	return d.dbName + "/" + path
}

func optionsToParams(opts map[string]interface{}) (url.Values, error) {
	params := url.Values{}
	for key, i := range opts {
		var values []string
		switch i.(type) {
		case string:
			values = []string{i.(string)}
		case []string:
			values = i.([]string)
		default:
			return nil, fmt.Errorf("Cannot convert type %T to []string", i)
		}
		for _, value := range values {
			params.Add(key, value)
		}
	}
	return params, nil
}

// AllDocs returns all of the documents in the database.
func (d *db) AllDocs(docs interface{}, opts map[string]interface{}) (offset, totalrows int, seq string, err error) {
	resp, err := d.client.newRequest(http.MethodGet, d.path("_all_docs")).
		AddHeader("Accept", typeJSON).
		Do()
	if err != nil {
		return 0, 0, "", err
	}
	defer resp.Body.Close()
	return ouchdb.AllDocs(resp.Body, docs)
}

// Get fetches the requested document.
func (d *db) Get(docID string, doc interface{}, opts map[string]interface{}) error {
	params, err := optionsToParams(opts)
	if err != nil {
		return err
	}
	return d.client.newRequest(http.MethodGet, d.url(docID, params)).
		AddHeader("Accept", typeJSON).
		AddHeader("Accept", typeMixed).
		DoJSON(doc)
}
