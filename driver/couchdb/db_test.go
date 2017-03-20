package couchdb

import (
	"io"
	"testing"

	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/test/kt"
)

func TestAllDocs(t *testing.T) {
	client := getClient(t)
	db, err := client.DBContext(kt.CTX, "_users")
	if err != nil {
		t.Fatalf("Failed to connect to db: %s", err)
	}
	rows, err := db.AllDocsContext(kt.CTX, nil)
	if err != nil {
		t.Fatalf("Failed: %s", err)
	}

	for {
		err := rows.Next(&driver.Row{})
		if err != nil {
			if err != io.EOF {
				t.Fatalf("Iteration failed: %s", err)
			}
			break
		}
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
