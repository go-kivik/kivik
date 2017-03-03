package test

import (
	"testing"

	"github.com/flimzy/kivik"
)

func init() {
	for _, suite := range []string{SuitePouchRemote, SuiteCouch16, SuiteCouch20, SuiteKivikMemory, SuiteCloudant} { //FIXME: SuiteKivikServer
		RegisterTest(suite, "DestroyDB", true, DestroyDB)
		RegisterTest(suite, "NotDestroyDB", true, NotDestroyDB)
	}
	// Local Pouch will never fail to destroy a DB, so skip NotDestroyDB for it.
	RegisterTest(SuitePouchLocal, "DestroyDB", true, DestroyDB)
}

// DestroyDB tests database destruction
func DestroyDB(clients *Clients, suite string, t *testing.T) {
	client := clients.Admin
	testDB := testDBName()
	if err := client.CreateDB(testDB); err != nil {
		t.Errorf("Failed to create database '%s': %s", testDB, err)
	}
	if err := client.DestroyDB(testDB); err != nil {
		t.Errorf("Failed to destroy database '%s': %s", testDB, err)
	}
}

// NotDestroyDB tests that database destruction fails if the db doesn't exist
func NotDestroyDB(clients *Clients, suite string, t *testing.T) {
	client := clients.Admin
	testDB := testDBName()
	err := client.DestroyDB(testDB)
	if err == nil {
		t.Errorf("Database destruction should have failed for non-existent database")
		return
	}
	if !kivik.ErrNotFound(err) {
		t.Errorf("Database destruction should have indicated NotFound, but instead: %s", err)
	}
}
