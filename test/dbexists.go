package test

import (
	"net/http"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
)

func init() {
	// For these variants, we can do a read-only test, checking for '_users'.
	for _, suite := range []string{SuiteCouch16, SuiteCouch20, SuiteCloudant, SuiteCouch16NoAuth, SuiteCouch20NoAuth} {
		RegisterTest(suite, "DBExists", false, DBExists)
	}
	// For unauthorized Cloudant, we expect an Unauthorized status for this check
	RegisterTest(SuiteCloudantNoAuth, "DBExistsUnauthorized", false, DBExistsUnauthorized)
	// For the rest, the only way to be sure a db exists is to create it first
	for _, suite := range []string{SuitePouchLocal, SuitePouchRemote, SuiteKivikMemory, SuiteKivikServer} {
		RegisterTest(suite, "DBExistsRW", true, DBExistsRW)
	}
	// For all of them, except local PouchDB, we can check for non-existence without writing
	for _, suite := range []string{SuitePouchRemote, SuiteCouch16, SuiteCouch20, SuiteKivikMemory, SuiteCloudant, SuiteKivikServer,
		SuitePouchRemoteNoAuth, SuiteCouch16NoAuth, SuiteCouch20NoAuth, SuiteCloudantNoAuth} {
		RegisterTest(suite, "DBNotExists", false, DBNotExists)
	}
}

// DBExistsRW creates a test database to check for its existence
func DBExistsRW(client *kivik.Client, suite string, fail FailFunc) {
	testDB := testDBName()
	defer client.DestroyDB(testDB)
	if err := client.CreateDB(testDB); err != nil {
		fail("Failed to create testDB '%s': %s", testDB, err)
		return
	}
	checkDBExists(client, testDB, true, 0, fail)
}

// DBExists checks for the existence of the '_users' system database
func DBExists(client *kivik.Client, suite string, fail FailFunc) {
	checkDBExists(client, "_users", true, 0, fail)
}

// DBExistsUnauthorized checks for the existence of the '_users' system database,
// but expects an unauthorized response
func DBExistsUnauthorized(client *kivik.Client, suite string, fail FailFunc) {
	checkDBExists(client, "_users", false, 401, fail)
}

// DBNotExists checks that a database does not exist
func DBNotExists(client *kivik.Client, suite string, fail FailFunc) {
	checkDBExists(client, testDBName(), false, 0, fail)
}

func checkDBExists(client *kivik.Client, dbName string, expected bool, expectedStatus int, fail FailFunc) {
	exists, err := client.DBExists(dbName)
	status := errors.StatusCode(err)
	if status == expectedStatus && exists == expected {
		return
	}
	if exists != expected {
		fail("Returned %t for '%s', expected %t", exists, dbName, expected)
	}
	if status != expectedStatus {
		fail("Failed to check existence of '%s'.\n\tExpected: %d/%s\n\t  Actual: %d/%s", dbName, expectedStatus, http.StatusText(expectedStatus), status, err)
	}
}
