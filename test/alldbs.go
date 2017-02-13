package test

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/flimzy/kivik"
)

func init() {
	for _, suite := range AllSuites {
		RegisterTest(suite, "AllDBs", false, AllDBs)
	}
	for _, suite := range AllSuites {
		RegisterTest(suite, "AllDBsRW", true, AllDBsRW)
	}
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
}

var rnd *rand.Rand

// AllDBs tests the '/_all_dbs' endpoint.
func AllDBs(client *kivik.Client, suite string, fail FailFunc) {
	var expected []string

	switch suite {
	case SuiteCouch, SuiteCloudant, SuiteCouch20:
		expected = []string{"_replicator", "_users"}
	}
	allDBs, err := client.AllDBs()
	if err != nil {
		fail("Failed to get all DBs: %s", err)
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
			fail("Database '%s' missing from allDBs result", exp)
		}
	}
}

// AllDBsRW tests the '/_all_dbs' endpoint in RW mode.
func AllDBsRW(client *kivik.Client, suite string, fail FailFunc) {
	testDB := fmt.Sprintf("kivik$%d", rnd.Int63())
	if err := client.CreateDB(testDB); err != nil {
		fail("Failed to create test DB '%s': %s", testDB, err)
		return
	}
	defer client.DestroyDB(testDB)
	allDBs, err := client.AllDBs()
	if err != nil {
		fail("Failed to get all DBs: %s", err)
		return
	}
	for _, db := range allDBs {
		if db == testDB {
			return
		}
	}
	fail("Test database '%s' missing from allDbs result", testDB)
}
