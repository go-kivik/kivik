// +build ignore

package test

import (
	"net/http"
	"testing"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
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
	admin := clients.Admin
	t.Run("Admin", func(t *testing.T) {
		t.Parallel()
		testDB := testDBName()
		if err := admin.CreateDB(testDB); err != nil {
			t.Errorf("Failed to create database '%s': %s", testDB, err)
			return
		}
		testDestroyDB(clients.Admin, testDB, 0, t)
	})
	if clients.NoAuth == nil {
		return
	}
	t.Run("NoAuth", func(t *testing.T) {
		t.Parallel()
		testDB := testDBName()
		if err := admin.CreateDB(testDB); err != nil {
			t.Errorf("Failed to create database '%s': %s", testDB, err)
			return
		}
		testDestroyDB(clients.NoAuth, testDB, http.StatusUnauthorized, t)
	})
}

func testDestroyDB(client *kivik.Client, dbName string, status int, t *testing.T) {
	err := client.DestroyDB(dbName)
	switch errors.StatusCode(err) {
	case status:
		// Expected
	case 0:
		t.Errorf("Expected failure with status %d/%s", status, http.StatusText(status))
	default:
		t.Errorf("Unexpected failure destroying database '%s'.\nExpected: %d/%s\n  Actual: %s\n", dbName, status, http.StatusText(status), err)
	}
}

// NotDestroyDB tests that database destruction fails if the db doesn't exist
func NotDestroyDB(clients *Clients, suite string, t *testing.T) {
	testDB := testDBName()
	t.Run("Admin", func(t *testing.T) {
		testDestroyDB(clients.Admin, testDB, http.StatusNotFound, t)
	})
	if clients.NoAuth == nil {
		return
	}
	t.Run("NoAuth", func(t *testing.T) {
		if suite == SuiteCloudant || suite == SuitePouchRemote {
			testDestroyDB(clients.NoAuth, testDB, http.StatusNotFound, t)
			return
		}
		testDestroyDB(clients.NoAuth, testDB, http.StatusUnauthorized, t)
	})
}
