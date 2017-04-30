// +build !js

package client

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func replicationOptions(ctx *kt.Context, client *kivik.Client, target, source, repID string, in map[string]interface{}) map[string]interface{} {
	if in == nil {
		in = make(map[string]interface{})
	}
	return in
}
