// +build js

package test

import (
	"os"
	"testing"

	"github.com/gopherjs/gopherjs/js"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/pouchdb"
)

func init() {
	kivik.Register("memdown", &pouchdb.Driver{
		Defaults: map[string]interface{}{
			"db": js.Global.Call("require", "memdown"),
		},
	})
}

func TestPouch(t *testing.T) {
	client, err := kivik.New("memdown", "")
	if err != nil {
		t.Errorf("Failed to connect to PouchDB/memdown driver: %s", err)
		return
	}
	RunSubtests(client, true, []string{SuitePouch}, t)
}

func TestPouchRemote(t *testing.T) {
	dsn := os.Getenv("KIVIK_COUCH16_DSN")
	if dsn == "" {
		t.Skip("KIVIK_COUCH16_DSN: Couch 1.6 DSN not set; skipping remote PouchDB tests")
	}
	client, err := kivik.New("pouch", dsn)
	if err != nil {
		t.Errorf("Failed to connect to remote PouchDB: %s\n", err)
		return
	}
	RunSubtests(client, true, []string{SuitePouchRemote}, t)
}
