package test

import (
	"net/http"

	"github.com/flimzy/kivik/test/kt"
)

func init() {
	RegisterSuite(SuiteKivikMemory, kt.SuiteConfig{
		"AllDBs.expected": []string{},

		"Config.status": http.StatusNotImplemented,

		"CreateDB/NoAuth.status":         http.StatusUnauthorized,
		"CreateDB/Admin/Recreate.status": http.StatusPreconditionFailed,

		"AllDocs/Admin.databases":  []string{"foo"},
		"AllDocs/Admin/foo.status": http.StatusNotFound,

		"DBExists/Admin.databases":      []string{"chicken"},
		"DBExists/Admin/chicken.exists": false,
		"DBExists/RW/Admin.exists":      true,

		"Membership.status": http.StatusNotImplemented,

		"UUIDs.counts":               []int{-1, 0, 1, 10},
		"UUIDs/Admin/-1Count.status": http.StatusBadRequest,
		// "UUIDs.status": http.StatusNotImplemented,

		"Log.status":          http.StatusNotImplemented,
		"Log/Admin/HTTP.skip": true,

		"ServerInfo.version":        `^0\.0\.1$`,
		"ServerInfo.vendor":         `^Kivik Memory Adaptor$`,
		"ServerInfo.vendor_version": `^0\.0\.1$`,
	})
}
