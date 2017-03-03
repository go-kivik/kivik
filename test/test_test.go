package test

import (
	"fmt"
	"net/url"
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
	RunSubtests(client, true, SuiteKivikMemory, t)
}

func doTest(suite, envName string, requireAuth bool, t *testing.T) {
	dsn := os.Getenv(envName)
	if dsn == "" {
		t.Skip("%s: %s DSN not set; skipping tests", envName, suite)
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		panic(err)
	}
	if requireAuth {
		if parsed.User == nil {
			t.Skip("%s: %s DSN does not include auth; skipping tests", envName, suite)
		}
	} else {
		parsed.User = nil
		dsn = parsed.String()
	}
	fmt.Printf("dsn = %s\n", dsn)
	client, err := kivik.New(driverMap[suite], dsn)
	if err != nil {
		t.Errorf("Failed to connect to %s: %s\n", suite, err)
		return
	}
	RunSubtests(client, true, suite, t)

}

func TestCloudant(t *testing.T) {
	doTest(SuiteCloudant, "KIVIK_CLOUDANT_DSN", true, t)
}

func TestCloudantNoAuth(t *testing.T) {
	doTest(SuiteCloudantNoAuth, "KIVIK_CLOUDANT_DSN", false, t)
}

func TestCouch16(t *testing.T) {
	doTest(SuiteCouch16, "KIVIK_COUCH16_DSN", true, t)
}

func TestCouch16NoAuth(t *testing.T) {
	doTest(SuiteCouch16NoAuth, "KIVIK_COUCH16_DSN", false, t)
}

func TestCouch20(t *testing.T) {
	doTest(SuiteCouch20, "KIVIK_COUCH20_DSN", true, t)
}

func TestCouch20NoAuth(t *testing.T) {
	doTest(SuiteCouch20NoAuth, "KIVIK_COUCH20_DSN", false, t)
}
