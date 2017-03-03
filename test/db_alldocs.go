package test

import (
	"net/http"
	"testing"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
)

func init() {
	for _, suite := range []string{SuiteCouch16, SuiteCouch20, SuiteCloudant} {
		RegisterTest(suite, "AllDocsCouch", false, AllDocsCouch)
	}
	// for _, suite := range []string{SuitePouch, , SuiteKivikMemory, SuiteKivikServer} {
	// 	RegisterTest(suite, "AllDocs", false, AllDocs)
	// 	// RegisterTest(suite, "AllDocsRW", true, AllDocsRW)
	// }
}

// AllDocsCouch tests the /{db}/_all_docs endpoint for CouchDB-like backends.
func AllDocsCouch(clients *Clients, suite string, t *testing.T) {
	t.Run("Admin", func(t *testing.T) {
		testAllDocsCouch(clients.Admin, t)
	})
	if clients.NoAuth == nil {
		return
	}
	expectedStatus := http.StatusForbidden
	if suite == SuiteCloudant {
		expectedStatus = http.StatusUnauthorized
	}
	t.Run("NoAuth", func(t *testing.T) {
		testAllDocsCouchFailure(clients.NoAuth, expectedStatus, t)
	})
}

func testAllDocsCouch(client *kivik.Client, t *testing.T) {
	db, err := client.DB("_replicator")
	if err != nil {
		t.Errorf("Failed to connect to database: %s", err)
		return
	}
	docs := []interface{}{}
	offset, total, _, err := db.AllDocs(&docs, nil)
	if err != nil {
		t.Errorf("Failed to fetch AllDocs: %s", err)
		return
	}
	if offset != 0 {
		t.Errorf("Expected offset of 0, got %d", offset)
	}
	if total < 1 {
		t.Errorf("Expected total >= 1, got %d", total)
	}
}

func testAllDocsCouchFailure(client *kivik.Client, expectedStatus int, t *testing.T) {
	db, err := client.DB("_replicator")
	if err != nil {
		t.Errorf("Failed to connect to database: %s", err)
		return
	}
	_, _, _, err = db.AllDocs(&struct{}{}, nil)
	switch errors.StatusCode(err) {
	case 0:
		t.Errorf("Expected an error")
	case expectedStatus:
		// expected
	default:
		t.Errorf("Expected %s, got %s", http.StatusText(expectedStatus), err)
	}
}
