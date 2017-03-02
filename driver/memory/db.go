package memory

import "net/url"

type db struct {
	*client
	dbName string
}

func (d *db) AllDocs(docs interface{}, opts url.Values) (offset, total int, err error) {
	return 0, 0, nil
}

func (d *db) Get(docID string, doc interface{}, opts url.Values) error {
	return nil
}
