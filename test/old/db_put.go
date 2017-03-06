// +build ignore

package test

import (
	"net/http"
	"testing"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
)

func init() {
	for _, suite := range []string{SuiteCouch16, SuiteCouch20, SuiteCloudant} {
		RegisterTest(suite, "Put", true, Put)
	}
}

type testDoc struct {
	ID   string `json:"_id"`
	Rev  string `json:"_rev,omitempty"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// Put tests creating and updating documents.
func Put(clients *Clients, suite string, t *testing.T) {
	t.Run("Admin", func(t *testing.T) {
		t.Parallel()
		testDB := testDBName()
		defer clients.Admin.DestroyDB(testDB)
		if err := clients.Admin.CreateDB(testDB); err != nil {
			t.Errorf("Failed to create database %s: %s", testDB, err)
			return
		}
		testPut(clients.Admin, testDB, kivik.StatusNoError, t)
	})
	if clients.NoAuth == nil {
		return
	}
	t.Run("NoAuth", func(t *testing.T) {
		t.Parallel()
		testDB := testDBName()
		defer clients.Admin.DestroyDB(testDB)
		if err := clients.Admin.CreateDB(testDB); err != nil {
			t.Errorf("Failed to create database %s: %s", testDB, err)
			return
		}
		if suite == SuiteCloudant {
			testPut(clients.NoAuth, testDB, http.StatusUnauthorized, t)
		} else {
			testPut(clients.NoAuth, testDB, kivik.StatusNoError, t)
		}
	})
	// t.Run("Forbidden", func(t *testing.T) {
	// 	t.Parallel()
	// 	testDB := testDBName()
	// 	defer clients.Admin.DestroyDB(testDB)
	// 	clients.Admin.Put()
	// })
}

func testPut(client *kivik.Client, dbName string, status int, t *testing.T) {
	db, err := client.DB(dbName)
	if err != nil {
		t.Errorf("Failed to connect to test database %s: %s", dbName, err)
		return
	}
	doc := &testDoc{
		ID:   "bob",
		Name: "Robert",
		Age:  32,
	}
	if putDoc(db, doc, status, t) {
		return
	}
	if status != kivik.StatusNoError {
		// We expected a failure, and we got it, so return
		return
	}
	doc.Age = 33
	if putDoc(db, doc, status, t) {
		return
	}
}

func putDoc(db *kivik.DB, doc *testDoc, status int, t *testing.T) bool {
	rev, err := db.Put(doc.ID, doc)
	switch errors.StatusCode(err) {
	case status:
		// expected
	case 0:
		t.Errorf("Expected failure %d/%s", status, http.StatusText(status))
		return true
	default:
		t.Errorf("Failed to put doc '%s': %s", doc.ID, err)
		return true
	}
	doc.Rev = rev
	return false
}
