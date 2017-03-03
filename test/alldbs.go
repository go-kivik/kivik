package test

import "testing"

func init() {
	for _, suite := range []string{SuitePouchLocal, SuiteCouch16, SuiteCouch20, SuiteKivikMemory, SuiteCloudant, SuiteKivikServer} {
		RegisterTest(suite, "AllDBs", false, AllDBs)
		RegisterTest(suite, "AllDBsRW", true, AllDBsRW)
	}
}

// AllDBs tests the '/_all_dbs' endpoint.
func AllDBs(clients *Clients, suite string, t *testing.T) {
	client := clients.Admin
	var expected []string

	switch suite {
	case SuitePouchRemote, SuiteCouch16, SuiteCloudant, SuiteCouch20:
		expected = []string{"_replicator", "_users"}
	}
	allDBs, err := client.AllDBs()
	if err != nil {
		t.Errorf("Failed to get all DBs: %s", err)
		return
	}
	if len(expected) == 0 {
		return
	}
	dblist := make(map[string]struct{})
	for _, db := range allDBs {
		dblist[db] = struct{}{}
	}
	for _, exp := range expected {
		if _, ok := dblist[exp]; !ok {
			t.Errorf("Database '%s' missing from allDBs result", exp)
		}
	}
}

// AllDBsRW tests the '/_all_dbs' endpoint in RW mode.
func AllDBsRW(clients *Clients, suite string, t *testing.T) {
	client := clients.Admin
	testDB := testDBName()
	if err := client.CreateDB(testDB); err != nil {
		t.Errorf("Failed to create test DB '%s': %s", testDB, err)
		return
	}
	defer client.DestroyDB(testDB)
	allDBs, err := client.AllDBs()
	if err != nil {
		t.Errorf("Failed to get all DBs: %s", err)
		return
	}
	for _, db := range allDBs {
		if db == testDB {
			return
		}
	}
	t.Errorf("Test database '%s' missing from allDbs result", testDB)
}
