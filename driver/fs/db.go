package fs

import "net/url"

type db struct {
	*client
	dbName string
}

func (d *db) AllDocs(docs interface{}, _ url.Values) (offset, totalrows int, err error) {
	return 0, 0, nil
}

func (d *db) Get(docID string, doc interface{}, opts url.Values) error {
	return nil
}
