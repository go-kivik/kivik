package test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/flimzy/kivik/errors"
)

func init() {
	for _, suite := range []string{SuitePouchLocal, SuitePouchRemote, SuiteCouch16, SuiteCouch20, SuiteKivikMemory, SuiteCloudant} { //FIXME: SuiteKivikServer
		RegisterTest(suite, "CreateDB", true, CreateDB)
	}
}

// CreateDB tests database creation.
func CreateDB(clients *Clients, suite string, t *testing.T) {
	client := clients.Admin
	testDB := testDBName()
	fmt.Printf("testDB = %s\n", testDB)
	// defer client.DestroyDB(testDB)
	err := client.CreateDB(testDB)
	if strings.Contains(suite, "NoAuth") {
		switch errors.StatusCode(err) {
		case 0:
			t.Errorf("CreateDB: Should fail for unauthenticated session")
		case http.StatusUnauthorized:
			// Expected
		default:
			t.Errorf("CreateDB: Expected 401/Unauthorized, Got: %s", err)
		}
		return
	}
	if err != nil {
		t.Errorf("Failed to create database '%s': %s", testDB, err)
	}
}
