package test

import "testing"

func init() {
	// for _, suite := range AllSuites {
	// 	RegisterTest(suite, "Get", true, Get)
	// }
}

// Get tests fetching of documents.
func Get(clients *Clients, suite string, t *testing.T) {
	testDB := testDBName()
	defer clients.Admin.DestroyDB(testDB)

}
