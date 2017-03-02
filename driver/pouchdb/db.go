package pouchdb

import (
	"bytes"

	"github.com/flimzy/kivik/driver/common"
	"github.com/flimzy/kivik/driver/pouchdb/bindings"
)

type db struct {
	db *bindings.DB
}

func (d *db) AllDocs(docs interface{}, options map[string]interface{}) (offset, totalrows int, seq string, err error) {
	body, err := d.db.AllDocs(options)
	if err != nil {
		return 0, 0, "", err
	}
	return common.AllDocs(bytes.NewReader(body), docs)
}

func (d *db) Get(docID string, doc interface{}, options map[string]interface{}) error {
	return nil
}
