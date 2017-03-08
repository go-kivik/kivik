package couchdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

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
	return d.client.newRequest(http.MethodGet, d.path(docID)).
		Query(params).
		AddHeader("Accept", typeJSON).
		AddHeader("Accept", typeMixed).
		DoJSON(doc)
}

func (d *db) CreateDoc(doc interface{}) (docID, rev string, err error) {
	return "", "", nil
}

func (d *db) Put(docID string, doc interface{}) (rev string, err error) {
	body, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}
	curRev, err := ouchdb.ExtractRev(body)
	if err != nil {
		return "", err
	}
	var result struct {
		Rev string `json:"rev"`
	}
	err = d.client.newRequest(http.MethodPut, d.path(docID)).
		AddHeader("Accept", typeJSON).
		AddHeader("Content-Type", typeJSON).
		If(curRev != "", func(r *request) { r.AddHeader("If-Match", curRev) }).
		Body(bytes.NewReader(body)).
		DoJSON(&result)
	return result.Rev, err
}

func (d *db) Flush() (time.Time, error) {
	result := struct {
		T int64 `json:"instance_start_time,string"`
	}{}
	err := d.client.newRequest(http.MethodPost, d.path("/_ensure_full_commit")).
		AddHeader("Accept", typeJSON).
		AddHeader("Content-Type", typeJSON).
		DoJSON(&result)
	return time.Unix(0, 0).Add(time.Duration(result.T) * time.Microsecond), err
}
