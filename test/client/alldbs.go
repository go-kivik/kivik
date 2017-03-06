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
	if conf.Bool(t, "skip") {
		return
	}
	if clients.RW && clients.Admin != nil {
		t.Run("RW", func(t *testing.T) {
			conf.Skip(t)
			testAllDBsRW(clients, conf, t)
		})
	}
	clients.RunAdmin(t, func(t *testing.T) {
		conf.Skip(t)
		// t.Parallel()
		testAllDBs(clients.Admin, conf, conf.StringSlice(t, "expected"), t)
	})
	clients.RunNoAuth(t, func(t *testing.T) {
		conf.Skip(t)
		// t.Parallel()
		testAllDBs(clients.NoAuth, conf, conf.StringSlice(t, "expected"), t)
	})
}

func testAllDBsRW(clients *kt.Clients, conf kt.SuiteConfig, t *testing.T) {
	admin := clients.Admin
	dbName := kt.TestDBName(t)
	defer admin.DestroyDB(dbName)
	if err := admin.CreateDB(dbName); err != nil {
		t.Errorf("Failed to create test DB '%s': %s", dbName, err)
		return
	}
	expected := append(conf.StringSlice(t, "expected"), dbName)
	clients.RunAdmin(t, func(t *testing.T) {
		conf.Skip(t)
		testAllDBs(clients.Admin, conf, expected, t)
	})
	clients.RunNoAuth(t, func(t *testing.T) {
		conf.Skip(t)
		testAllDBs(clients.NoAuth, conf, expected, t)
	})
}

func testAllDBs(client *kivik.Client, conf kt.SuiteConfig, expected []string, t *testing.T) {
	allDBs, err := client.AllDBs()
	status := conf.Int(t, "status")
	kt.IsError(err, status, t)
	if status > 0 {
		return
	}
	if d := diff.TextSlices(expected, allDBs); d != "" {
		t.Errorf("AllDBs() returned unexpected list:\n%s\n", d)
	}
	if len(expected) == 0 {
		return
	}
}
