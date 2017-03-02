package test

import "github.com/flimzy/kivik"

func init() {
	for _, suite := range []string{SuiteCouch, SuiteCouch20, SuiteCloudant} {
		RegisterTest(suite, "AllDocsCouch", false, AllDocsCouch)
	}
	// for _, suite := range []string{SuiteMinimal, SuitePouch, , SuiteKivikMemory, SuiteKivikServer} {
	// 	RegisterTest(suite, "AllDocs", false, AllDocs)
	// 	// RegisterTest(suite, "AllDocsRW", true, AllDocsRW)
	// }
}

// AllDocsCouch tests the /{db}/_all_docs endpoint for CouchDB-like backends.
func AllDocsCouch(client *kivik.Client, _ string, fail FailFunc) {
	db, err := client.DB("_replicator")
	if err != nil {
		fail("Failed to connect to database: %s", err)
	}
	docs := []interface{}{}
	offset, total, _, err := db.AllDocs(&docs, nil)
	if err != nil {
		fail("Failed to fetch AllDocs: %s", err)
		return
	}
	if offset != 0 {
		fail("Expected offset of 0, got %d", offset)
	}
	if total < 1 {
		fail("Expected total >= 1, got %d", total)
	}
}
