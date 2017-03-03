package test

import (
	"net/http"
	"testing"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
)

func init() {
	for _, suite := range []string{SuitePouchLocal, SuitePouchRemote, SuiteCouch16, SuiteCouch20, SuiteKivikMemory, SuiteCloudant} { //FIXME: SuiteKivikServer
		RegisterTest(suite, "CreateDB", true, CreateDB)
	}
}

// CreateDB tests database creation.
func CreateDB(clients *Clients, suite string, t *testing.T) {
	t.Run("Admin", func(t *testing.T) {
		t.Parallel()
		testDB := testDBName()
		defer clients.Admin.DestroyDB(testDB)
		testCreateDB(clients.Admin, testDB, 0, t)
	})
	if clients.NoAuth == nil {
		return
	}
	t.Run("NoAuth", func(t *testing.T) {
		t.Parallel()
		testDB := testDBName()
		defer clients.NoAuth.DestroyDB(testDB) // Just in case we succeed
		testCreateDB(clients.NoAuth, testDB, http.StatusUnauthorized, t)
	})
}

func testCreateDB(client *kivik.Client, dbName string, status int, t *testing.T) {
	err := client.CreateDB(dbName)
	switch errors.StatusCode(err) {
	case status:
		// Expected
	case 0:
		t.Errorf("Expected failure creating database '%s' %d/%s", dbName, status, http.StatusText(status))
	default:
		t.Errorf("Failed to create database '%s': %s", dbName, err)
	}
}
