package test

import "github.com/flimzy/kivik"

func init() {
	for _, suite := range []string{SuitePouchRemote, SuiteCouch, SuiteCouch20, SuiteKivikMemory, SuiteCloudant} { //FIXME: SuiteKivikServer
		RegisterTest(suite, "DestroyDB", true, DestroyDB)
		RegisterTest(suite, "NotDestroyDB", true, NotDestroyDB)
	}
	// Local Pouch will never fail to destroy a DB, so skip NotDestroyDB for it.
	RegisterTest(SuitePouch, "DestroyDB", true, DestroyDB)
}

// DestroyDB tests database destruction
func DestroyDB(client *kivik.Client, suite string, fail FailFunc) {
	testDB := testDBName()
	if err := client.CreateDB(testDB); err != nil {
		fail("Failed to create database '%s': %s", testDB, err)
	}
	if err := client.DestroyDB(testDB); err != nil {
		fail("Failed to destroy database '%s': %s", testDB, err)
	}
}

// NotDestroyDB tests that database destruction fails if the db doesn't exist
func NotDestroyDB(client *kivik.Client, suite string, fail FailFunc) {
	testDB := testDBName()
	err := client.DestroyDB(testDB)
	if err == nil {
		fail("Database destruction should have failed for non-existent database")
		return
	}
	if !kivik.ErrNotFound(err) {
		fail("Database destruction should have indicated NotFound, but instead: %s", err)
	}
}
