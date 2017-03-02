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
	RunSubtests(client, true, []string{SuiteKivikMemory}, t)
}

func doTest(suite, envName string, t *testing.T) {
	dsn := os.Getenv(envName)
	if dsn == "" {
		t.Skip("%s: %s DSN not set; skipping tests", envName, suite)
	}
	client, err := kivik.New(driverMap[suite], dsn)
	if err != nil {
		t.Errorf("Failed to connect to %s: %s\n", suite, err)
		return
	}
	RunSubtests(client, true, []string{suite}, t)

}

func TestCloudant(t *testing.T) {
	doTest(SuiteCloudant, "KIVIK_CLOUDANT_DSN", t)
}

func TestCouch16(t *testing.T) {
	doTest(SuiteCouch16, "KIVIK_COUCH16_DSN", t)
}

func TestCouch20(t *testing.T) {
	doTest(SuiteCouch20, "KIVIK_COUCH20_DSN", t)
}
