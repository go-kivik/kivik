package test

import "testing"

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
func Put(clients *Clients, _ string, t *testing.T) {
	client := clients.Admin
	testDB := testDBName()
	defer client.DestroyDB(testDB)
	if err := client.CreateDB(testDB); err != nil {
		t.Errorf("Failed to create database %s: %s", testDB, err)
	}
	db, err := client.DB(testDB)
	if err != nil {
		t.Errorf("Failed to connect to test database %s: %s", testDB, err)
		return
	}
	doc := testDoc{
		ID:   "bob",
		Name: "Robert",
		Age:  32,
	}
	rev, err := db.Put(doc.ID, doc)
	if err != nil {
		t.Errorf("Failed to create new doc '%s': %s", doc.ID, err)
		return
	}
	doc.Rev = rev
	doc.Age = 33
	_, err = db.Put(doc.ID, doc)
	if err != nil {
		t.Errorf("Failed to update doc '%s'/'%s': %s", doc.ID, rev, err)
	}
}
