package test

import (
	"net/http"
	"testing"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
)

func init() {
	// For these variants, we can do a read-only test, checking for '_users'.
	for _, suite := range []string{SuiteCouch16, SuiteCouch20, SuiteCloudant} {
		RegisterTest(suite, "DBExists", false, DBExists)
	}
	// For the rest, the only way to be sure a db exists is to create it first
	for _, suite := range []string{SuitePouchLocal, SuitePouchRemote, SuiteKivikMemory, SuiteKivikServer} {
		RegisterTest(suite, "DBExistsRW", true, DBExistsRW)
	}
	// For all of them, except local PouchDB, we can check for non-existence without writing
	for _, suite := range []string{SuitePouchRemote, SuiteCouch16, SuiteCouch20, SuiteKivikMemory, SuiteCloudant, SuiteKivikServer} {
		RegisterTest(suite, "DBNotExists", false, DBNotExists)
	}
}

// DBExistsRW creates a test database to check for its existence
func DBExistsRW(clients *Clients, suite string, t *testing.T) {
	client := clients.Admin
	testDB := testDBName()
	defer client.DestroyDB(testDB)
	if err := client.CreateDB(testDB); err != nil {
		t.Errorf("Failed to create testDB '%s': %s", testDB, err)
		return
	}
	t.Run("Admin", func(t *testing.T) {
		checkDBExists(clients.Admin, testDB, true, 0, t)
	})
	if clients.NoAuth == nil {
		return
	}
	t.Run("NoAuth", func(t *testing.T) {
		checkDBExists(clients.NoAuth, testDB, true, 0, t)
	})
}

// DBExists checks for the existence of the '_users' system database
func DBExists(clients *Clients, suite string, t *testing.T) {
	t.Run("Admin", func(t *testing.T) {
		checkDBExists(clients.Admin, "_users", true, 0, t)
	})
	if clients.NoAuth == nil {
		return
	}
	t.Run("NoAuth", func(t *testing.T) {
		if suite == SuiteCloudant {
			checkDBExists(clients.NoAuth, "_users", false, http.StatusUnauthorized, t)
		} else {
			checkDBExists(clients.NoAuth, "_users", true, 0, t)
		}
	})
}

// DBNotExists checks that a database does not exist
func DBNotExists(clients *Clients, suite string, t *testing.T) {
	t.Run("Admin", func(t *testing.T) {
		checkDBExists(clients.Admin, testDBName(), false, 0, t)
	})
	if clients.NoAuth == nil {
		return
	}
	t.Run("NoAuth", func(t *testing.T) {
		checkDBExists(clients.NoAuth, testDBName(), false, 0, t)
	})
}

func checkDBExists(client *kivik.Client, dbName string, expected bool, expectedStatus int, t *testing.T) {
	exists, err := client.DBExists(dbName)
	status := errors.StatusCode(err)
	if status == expectedStatus && exists == expected {
		return
	}
	if exists != expected {
		t.Errorf("Returned %t for '%s', expected %t", exists, dbName, expected)
	}
	if status != expectedStatus {
		t.Errorf("Failed to check existence of '%s'.\n\tExpected: %d/%s\n\t  Actual: %d/%s", dbName, expectedStatus, http.StatusText(expectedStatus), status, err)
	}
}
