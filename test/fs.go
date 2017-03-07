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

		"DBExists/Admin.databases":      []string{"chicken"},
		"DBExists/Admin/chicken.exists": false,
		"DBExists/RW/Admin.exists":      true,

		"Membership.status": http.StatusNotImplemented,

		"UUIDs.counts": []int{1},
		"UUIDs.status": http.StatusNotImplemented,

		"Log.status":          http.StatusNotImplemented,
		"Log/Admin/HTTP.skip": true,
	})
}
