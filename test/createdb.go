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
		testCreateDB(clients.Admin, t)
	})
	if clients.NoAuth == nil {
		return
	}
	t.Run("NoAuth", func(t *testing.T) {
		testCreateDBUnauthorized(clients.NoAuth, t)
	})
}

func testCreateDBUnauthorized(client *kivik.Client, t *testing.T) {
	testDB := testDBName()
	defer client.DestroyDB(testDB) // Just in case we succeed
	err := client.CreateDB(testDB)
	switch errors.StatusCode(err) {
	case 0:
		t.Errorf("CreateDB: Should fail for unauthenticated session")
	case http.StatusUnauthorized:
		// Expected
	default:
		t.Errorf("CreateDB: Expected 401/Unauthorized, Got: %s", err)
	}
}

func testCreateDB(client *kivik.Client, t *testing.T) {
	testDB := testDBName()
	defer client.DestroyDB(testDB)
	err := client.CreateDB(testDB)
	if err != nil {
		t.Errorf("Failed to create database '%s': %s", testDB, err)
	}
}
