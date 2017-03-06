// +build ignore

package test

import (
	"reflect"
	"testing"
)

func init() {
	for _, suite := range []string{SuiteCouch16, SuiteCouch20, SuiteCloudant} {
		RegisterTest(suite, "Get", true, Get)
	}
}

// Get tests fetching of documents.
func Get(clients *Clients, suite string, t *testing.T) {
	admin := clients.Admin
	// noAuth := clients.NoAuth
	testDB := testDBName()
	defer admin.DestroyDB(testDB)
	if err := admin.CreateDB(testDB); err != nil {
		t.Errorf("Failed to create database '%s': %s", testDB, err)
		return
	}
	adminDB, err := admin.DB(testDB)
	if err != nil {
		t.Errorf("Failed to connect to database: %s", err)
		return
	}
	doc := &testDoc{
		ID:   "bob",
		Name: "Robert",
		Age:  32,
	}
	rev, err := adminDB.Put(doc.ID, doc)
	if err != nil {
		t.Errorf("Failed to create document: %s", err)
		return
	}
	doc.Rev = rev
	recvDoc := &testDoc{}
	if err = adminDB.Get(doc.ID, recvDoc, nil); err != nil {
		t.Errorf("Failed to fetch document: %s", err)
		return
	}
	if !reflect.DeepEqual(doc, recvDoc) {
		t.Errorf("Retrieved document does not match")
	}
}
