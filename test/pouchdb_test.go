// +build js

package test

import (
	"context"
	"testing"

	"github.com/gopherjs/gopherjs/js"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/pouchdb"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	memPouch := js.Global.Get("PouchDB").Call("defaults", map[string]interface{}{
		"db": js.Global.Call("require", "memdown"),
	})
	js.Global.Set("PouchDB", memPouch)
}

func TestPouchLocal(t *testing.T) {
	client, err := kivik.New(context.Background(), "pouch", "")
	if err != nil {
		t.Errorf("Failed to connect to PouchDB driver: %s", err)
		return
	}
	clients := &kt.Context{
		RW:    true,
		Admin: client,
	}
	runTests(clients, SuitePouchLocal, t)
}

func TestPouchRemote(t *testing.T) {
	doTest(SuitePouchRemote, "KIVIK_TEST_DSN_COUCH20", t)
}
