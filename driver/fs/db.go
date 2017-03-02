package fs

type db struct {
	*client
	dbName string
}

func (d *db) AllDocs(docs interface{}, _ map[string]interface{}) (offset, totalrows int, err error) {
	return 0, 0, nil
}

func (d *db) Get(docID string, doc interface{}, opts map[string]interface{}) error {
	return nil
}
