package test

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver/pouchdb"
	"github.com/gopherjs/gopherjs/js"
)

const TestServer = ""

func init() {
	kivik.Register("memdown", &pouchdb.Driver{
		Defaults: map[string]interface{}{
			"db": js.Global.Call("require", "memdown"),
		},
	})
}
