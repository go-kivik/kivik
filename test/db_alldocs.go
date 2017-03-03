package test

import "testing"

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
func AllDocsCouch(clients *Clients, _ string, t *testing.T) {
	client := clients.Admin
	db, err := client.DB("_replicator")
	if err != nil {
		t.Errorf("Failed to connect to database: %s", err)
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
