package couchdb

import (
	"io"
	"testing"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/test/kt"
)

func TestAllDocs(t *testing.T) {
	client := getClient(t)
	db, err := client.DBContext(kt.CTX, "_users", kivik.Options{"force_commit": true})
	if err != nil {
		t.Fatalf("Failed to connect to db: %s", err)
	}
	rows, err := db.AllDocsContext(kt.CTX, map[string]interface{}{
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
	db, err := client.DBContext(kt.CTX, "_users", kivik.Options{"force_commit": true})
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

func TestEncodeDocID(t *testing.T) {
	tests := []struct {
		Input    string
		Expected string
	}{
		{Input: "foo", Expected: "foo"},
		{Input: "foo/bar", Expected: "foo%2Fbar"},
		{Input: "_design/foo", Expected: "_design/foo"},
		{Input: "_design/foo/bar", Expected: "_design/foo%2Fbar"},
	}
	for _, test := range tests {
		result := encodeDocID(test.Input)
		if result != test.Expected {
			t.Errorf("Unexpected encoded DocID from %s\n\tExpected: %s\n\t  Actual: %s\n", test.Input, test.Expected, result)
		}
	}
}
