package pouchdb

import (
	"net/url"

	"github.com/flimzy/kivik/driver/pouchdb/bindings"
)

type db struct {
	db *bindings.DB
}

func (d *db) AllDocs(docs interface{}, options url.Values) (offset, totalrows int, err error) {
	return 0, 0, nil
}

func (d *db) Get(docID string, doc interface{}, options url.Values) error {
	return nil
}
