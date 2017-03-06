// +build js

package test

import (
	"net/http"

	"github.com/flimzy/kivik/test/kt"
)

func init() {
	RegisterSuite(SuitePouchLocal, kt.SuiteConfig{
		"AllDBs.skip":        true,
		"Config.status":      http.StatusNotImplemented,
		"Config/RW.skip":     true,
		"Config/Admin.skip":  true,
		"Config.NoAuth.skip": true,
	})
	RegisterSuite(SuitePouchRemote, kt.SuiteConfig{
		"PreCleanup.skip":    true,
		"AllDBs.skip":        true,
		"Config.status":      http.StatusNotImplemented,
		"Config/RW.skip":     true,
		"Config/Admin.skip":  true,
		"Config.NoAuth.skip": true,
	})
}
