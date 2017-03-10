package couchdb

import "testing"

func TestAllDocs(t *testing.T) {
	client := getClient(t)
	db, err := client.DBContext(CTX, "_users")
	if err != nil {
		t.Fatalf("Failed to connect to db: %s", err)
	}
	docs := []interface{}{}
	_, _, _, err = db.AllDocsContext(CTX, &docs, nil)
	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
}
