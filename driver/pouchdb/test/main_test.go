package test

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver/pouchdb"
	"github.com/gopherjs/gopherjs/js"
)

const TestDriver = "memdown"
const TestServer = ""
const RemoteServer = "https://kivik:K1v1k123@kivik.cloudant.com/"

func init() {
	kivik.Register("memdown", &pouchdb.Driver{
		Defaults: map[string]interface{}{
			"db": js.Global.Call("require", "memdown"),
		},
	})
}
