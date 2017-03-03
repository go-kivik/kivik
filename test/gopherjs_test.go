// +build js

package test

import (
	"testing"

	"github.com/gopherjs/gopherjs/js"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver/pouchdb"
)

func init() {
	kivik.Register("memdown", &pouchdb.Driver{
		Defaults: map[string]interface{}{
			"db": js.Global.Call("require", "memdown"),
		},
	})
}

func TestPouchLocal(t *testing.T) {
	client, err := kivik.New("memdown", "")
	if err != nil {
		t.Errorf("Failed to connect to PouchDB/memdown driver: %s", err)
		return
	}
	RunSubtests(client, true, SuitePouchLocal, t)
}

func TestPouchRemote(t *testing.T) {
	doTest(SuitePouchRemote, "KIVIK_COUCH16_DSN", true, t)
}

func TestPouchRemoteNoAuth(t *testing.T) {
	doTest(SuitePouchRemoteNoAuth, "KIVIK_COUCH16_DSN", false, t)
}
