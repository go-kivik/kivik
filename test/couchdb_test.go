package test

import (
	"os"
	"testing"

	_ "github.com/flimzy/kivik/driver/couchdb"
	_ "github.com/flimzy/kivik/driver/fs"
	_ "github.com/flimzy/kivik/driver/memory"
)

func doTest(suite, envName string, t *testing.T) {
	dsn := os.Getenv(envName)
	if dsn == "" {
		t.Skipf("%s: %s DSN not set; skipping tests", envName, suite)
	}
	clients, err := connectClients(driverMap[suite], dsn, t)
	if err != nil {
		t.Errorf("Failed to connect to %s: %s\n", suite, err)
		return
	}
	clients.RW = true
	if clients.Admin != nil {
		count, err := doCleanup(clients.Admin, true)
		if count > 0 {
			t.Logf("Pre-cleanup removed %d databases from previous test runs", count)
		}
		if err != nil {
			t.Fatalf("Pre-cleanup failed: %s", err)
		}
	}
	runTests(clients, suite, t)
}

func TestCouch16(t *testing.T) {
	doTest(SuiteCouch16, "KIVIK_TEST_DSN_COUCH16", t)
}

func TestCloudant(t *testing.T) {
	doTest(SuiteCloudant, "KIVIK_TEST_DSN_CLOUDANT", t)
}

func TestCouch20(t *testing.T) {
	doTest(SuiteCouch20, "KIVIK_TEST_DSN_COUCH20", t)
}
