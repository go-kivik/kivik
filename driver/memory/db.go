package memory

type db struct {
	*client
	dbName string
}

func (d *db) AllDocs(docs interface{}, opts map[string]interface{}) (offset, total int, err error) {
	return 0, 0, nil
}

func (d *db) Get(docID string, doc interface{}, opts map[string]interface{}) error {
	return nil
}
