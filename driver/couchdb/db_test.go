package couchdb

import (
	"context"
	"io"
	"testing"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
)

func TestAllDocs(t *testing.T) {
	client := getClient(t)
	db, err := client.DB(context.Background(), "_users", kivik.Options{"force_commit": true})
	if err != nil {
		t.Fatalf("Failed to connect to db: %s", err)
	}
	rows, err := db.AllDocs(context.Background(), map[string]interface{}{
		"include_docs": true,
	})
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
	db, err := client.DB(context.Background(), "_users", kivik.Options{"force_commit": true})
	if err != nil {
		t.Fatalf("Failed to connect to db: %s", err)
	}
	info, err := db.Info(context.Background())
	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if info.Name != "_users" {
		t.Errorf("Unexpected name %s", info.Name)
	}
}
