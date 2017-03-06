package test

import (
	"net/http"

	"github.com/flimzy/kivik/test/kt"
)

func init() {
	RegisterSuite(SuiteKivikFS, kt.SuiteConfig{
		"AllDBs.expected": []string{},

		"Config.status": http.StatusNotImplemented,

		"CreateDB/NoAuth.status":         http.StatusUnauthorized,
		"CreateDB/Admin/Recreate.status": http.StatusPreconditionFailed,

		"AllDocs.skip": true, // FIXME: Not yet implemented
		// "AllDocs/Admin.databases":  []string{"foo"},
		// "AllDocs/Admin/foo.status": http.StatusNotFound,
	})
}
