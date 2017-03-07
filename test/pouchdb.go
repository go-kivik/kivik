// +build js

package test

import (
	"net/http"

	"github.com/flimzy/kivik/test/kt"
)

func init() {
	RegisterSuite(SuitePouchLocal, kt.SuiteConfig{
		"AllDBs.skip": true,

		"Config.status":      http.StatusNotImplemented,
		"Config/RW.skip":     true,
		"Config/Admin.skip":  true,
		"Config.NoAuth.skip": true,

		"AllDocs/Admin.skip":  true,
		"AllDocs/NoAuth.skip": true,
		"AllDocs/RW.skip":     true, // FIXME: Not sure why this is broken

		"ServerInfo.version":        `^1\.6\.1$`,
		"ServerInfo.vendor":         "Kivik",
		"ServerInfo.vendor_version": `^0\.0\.1$`,
	})
	RegisterSuite(SuitePouchRemote, kt.SuiteConfig{
		"PreCleanup.skip": true,

		"AllDBs.skip": true,

		"Config.status":      http.StatusNotImplemented,
		"Config/RW.skip":     true,
		"Config/Admin.skip":  true,
		"Config.NoAuth.skip": true,

		"CreateDB/NoAuth.status":         http.StatusUnauthorized,
		"CreateDB/Admin/Recreate.status": http.StatusPreconditionFailed,

		"AllDocs/Admin.databases":            []string{"_replicator"},
		"AllDocs/Admin/_replicator.expected": []string{"_design/_replicator"},
		"AllDocs/Admin/_replicator.offset":   0,
		"AllDocs/NoAuth.databases":           []string{"_replicator"},
		"AllDocs/NoAuth/_replicator.status":  http.StatusForbidden,
		"AllDocs/RW.skip":                    true, // FIXME: Not sure why this is broken

		"ServerInfo.version":        `^1\.6\.1$`,
		"ServerInfo.vendor":         "Kivik",
		"ServerInfo.vendor_version": `^0\.0\.1$`,
	})
}
