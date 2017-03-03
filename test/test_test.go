package test

import (
	"os"
	"testing"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/couchdb"
	_ "github.com/flimzy/kivik/driver/fs"
	_ "github.com/flimzy/kivik/driver/memory"
)

func TestMemory(t *testing.T) {
	client, err := kivik.New("memory", "")
	if err != nil {
		t.Errorf("Failed to connect to memory driver: %s\n", err)
		return
	}
	clients := &Clients{
		Admin: client,
	}
	RunSubtests(clients, true, SuiteKivikMemory, t)
}

func doTest(suite, envName string, requireAuth bool, t *testing.T) {
	dsn := os.Getenv(envName)
	if dsn == "" {
		t.Skip("%s: %s DSN not set; skipping tests", envName, suite)
	}
	clients, err := connectClients(driverMap[suite], dsn)
	if err != nil {
		t.Errorf("Failed to connect to %s: %s\n", suite, err)
		return
	}
	RunSubtests(clients, true, suite, t)

}

func TestCloudant(t *testing.T) {
	doTest(SuiteCloudant, "KIVIK_TEST_DSN_CLOUDANT", true, t)
}

func TestCouch16(t *testing.T) {
	doTest(SuiteCouch16, "KIVIK_TEST_DSN_COUCH16", true, t)
}

func TestCouch20(t *testing.T) {
	doTest(SuiteCouch20, "KIVIK_TEST_DSN_COUCH20", true, t)
}
