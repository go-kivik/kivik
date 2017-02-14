package test

import "github.com/flimzy/kivik"

func init() {
	// For these variants, we can do a read-only test, checking for '_users'.
	for _, suite := range []string{SuiteCouch, SuiteCouch20, SuiteCloudant} {
		RegisterTest(suite, "DBExists", false, DBExists)
	}
	// For the rest, the only way to be sure a db exists is to create it first
	for _, suite := range []string{SuiteMinimal, SuitePouch, SuitePouchRemote, SuiteKivikMemory} { //FIXME: SuiteKivikServer
		RegisterTest(suite, "DBExistsRW", true, DBExistsRW)
	}
	// For all of them, except local PouchDB, we can check for non-existence without writing
	for _, suite := range []string{SuiteMinimal, SuitePouchRemote, SuiteCouch, SuiteCouch20, SuiteKivikMemory, SuiteCloudant} { //FIXME: SuiteKivikServer
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
	checkDBExists(client, testDB, true, fail)
}

// DBExists checks for the existence of the '_users' system database
func DBExists(client *kivik.Client, suite string, fail FailFunc) {
	checkDBExists(client, "_users", true, fail)
}

// DBNotExists checks that a database does not exist
func DBNotExists(client *kivik.Client, suite string, fail FailFunc) {
	checkDBExists(client, testDBName(), false, fail)
}

func checkDBExists(client *kivik.Client, dbName string, expected bool, fail FailFunc) {
	exists, err := client.DBExists(dbName)
	if err != nil {
		fail("Failed to check existence of '%s': %s", dbName, err)
	}
	if exists != expected {
		fail("DBExists() returned %t for '%s', expected %t", exists, dbName, expected)
	}
}
