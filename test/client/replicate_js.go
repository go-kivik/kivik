// +build js

package client

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver/pouchdb/bindings"
	"github.com/flimzy/kivik/test/kt"
)

func replicationOptions(ctx *kt.Context, client *kivik.Client, target, source, repID string, in map[string]interface{}) map[string]interface{} {
	if in == nil {
		in = make(map[string]interface{})
	}
	if ctx.String("mode") != "pouchdb" {
		in["_id"] = repID
		return in
	}
	in["source"] = bindings.GlobalPouchDB().New(source, ctx.Options("db"))
	in["target"] = bindings.GlobalPouchDB().New(target, ctx.Options("db"))
	return in
}
