package couchdb

import (
	"testing"

	"github.com/flimzy/kivik/test/kt"
)

func TestAllDocs(t *testing.T) {
	client := getClient(t)
	db, err := client.DBContext(kt.CTX, "_users")
	if err != nil {
		t.Fatalf("Failed to connect to db: %s", err)
	}
	docs := []interface{}{}
	_, _, _, err = db.AllDocsContext(kt.CTX, &docs, nil)
	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
}

func TestDBInfo(t *testing.T) {
	client := getClient(t)
	db, err := client.DBContext(kt.CTX, "_users")
	if err != nil {
		t.Fatalf("Failed to connect to db: %s", err)
	}
	info, err := db.InfoContext(kt.CTX)
	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if info.Name != "_users" {
		t.Errorf("Unexpected name %s", info.Name)
	}
}
