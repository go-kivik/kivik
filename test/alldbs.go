package test

import (
	"net/http"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
)

func init() {
	for _, suite := range []string{SuitePouchLocal, SuiteCouch16, SuiteCouch20, SuiteKivikMemory, SuiteCloudant, SuiteKivikServer} {
		RegisterTest(suite, "AllDBs", false, AllDBs)
		RegisterTest(suite, "AllDBsRW", true, AllDBsRW)
	}
	// Without auth, all we can do is RO tests
	for _, suite := range []string{SuiteCouch16NoAuth, SuiteCouch20NoAuth} {
		RegisterTest(suite, "AllDBs", false, AllDBs)
	}
	// Cloudant rejects unauthorized _all_dbs queries by default.
	RegisterTest(SuiteCloudantNoAuth, "AllDBsFailNoAuth", false, AllDBsFailNoAuth)
}

// AllDBsFailNoAuth tests unauthorized clients
func AllDBsFailNoAuth(client *kivik.Client, suite string, fail FailFunc) {
	_, err := client.AllDBs()
	switch errors.StatusCode(err) {
	case 0:
		fail("AllDBs(): should have failed for %s", suite)
	case http.StatusUnauthorized:
	default:
		fail("AllDBs(): Expected 401/Unauthorized, got: %s", err)
	}
	return
}

// AllDBs tests the '/_all_dbs' endpoint.
func AllDBs(client *kivik.Client, suite string, fail FailFunc) {
	var expected []string

	switch suite {
	case SuitePouchRemote, SuiteCouch16, SuiteCloudant, SuiteCouch20:
		expected = []string{"_replicator", "_users"}
	}
	allDBs, err := client.AllDBs()
	if err != nil {
		fail("Failed to get all DBs: %s", err)
		return
	}
	if len(expected) == 0 {
		return
	}
	dblist := make(map[string]struct{})
	for _, db := range allDBs {
		dblist[db] = struct{}{}
	}
	for _, exp := range expected {
		if _, ok := dblist[exp]; !ok {
			fail("Database '%s' missing from allDBs result", exp)
		}
	}
}

// AllDBsRW tests the '/_all_dbs' endpoint in RW mode.
func AllDBsRW(client *kivik.Client, suite string, fail FailFunc) {
	testDB := testDBName()
	if err := client.CreateDB(testDB); err != nil {
		fail("Failed to create test DB '%s': %s", testDB, err)
		return
	}
	defer client.DestroyDB(testDB)
	allDBs, err := client.AllDBs()
	if err != nil {
		fail("Failed to get all DBs: %s", err)
		return
	}
	for _, db := range allDBs {
		if db == testDB {
			return
		}
	}
	fail("Test database '%s' missing from allDbs result", testDB)
}
