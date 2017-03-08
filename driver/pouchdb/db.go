package pouchdb

import (
	"bytes"

	"github.com/flimzy/kivik/driver/ouchdb"
	"github.com/flimzy/kivik/driver/pouchdb/bindings"
)

type db struct {
	db *bindings.DB
}

func (d *db) AllDocs(docs interface{}, options map[string]interface{}) (offset, totalrows int, updateSeq string, err error) {
	body, err := d.db.AllDocs(options)
	if err != nil {
		return 0, 0, "", err
	}
	return ouchdb.AllDocs(bytes.NewReader(body), docs)
}

func (d *db) Get(docID string, doc interface{}, options map[string]interface{}) error {
	return nil
}

func (d *db) CreateDoc(doc interface{}) (docID, rev string, err error) {
	return "", "", nil
}

func (d *db) Put(docID string, doc interface{}) (rev string, err error) {
	return "", nil
}

func (d *db) Delete(docID, rev string) (newRev string, err error) {
	return "", nil
}
