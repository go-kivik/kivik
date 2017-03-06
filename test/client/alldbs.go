package client

import (
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("AllDBs", allDBs)
}

func allDBs(clients *kt.Clients, conf kt.SuiteConfig, t *testing.T) {
	t.Run("Admin", func(t *testing.T) {
		testAllDBs(clients.Admin, conf, t)
	})
}

func testAllDBs(client *kivik.Client, conf kt.SuiteConfig, t *testing.T) {
	t.Parallel()
	allDBs, err := client.AllDBs()
	if err != nil {
		t.Errorf("Failed to get all DBs: %s", err)
		return
	}
	expected := conf.StringSlice(t, "expected")
	if len(allDBs) != len(expected) {
		d := diff.TextSlices(expected, allDBs)
		t.Errorf("Found %d databases, expected %d:\n%s\n", len(allDBs), len(expected), d)
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
