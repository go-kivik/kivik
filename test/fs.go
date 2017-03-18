package test

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	RegisterSuite(SuiteKivikFS, kt.SuiteConfig{
		"AllDBs.expected": []string{},

		"Config.status": kivik.StatusNotImplemented,

		"CreateDB/RW/NoAuth.status":         kivik.StatusUnauthorized,
		"CreateDB/RW/Admin/Recreate.status": kivik.StatusPreconditionFailed,

		"AllDocs.skip": true, // FIXME: Not yet implemented
		// "AllDocs/Admin.databases":  []string{"foo"},
		// "AllDocs/Admin/foo.status": kivik.StatusNotFound,

		"DBExists/Admin.databases":       []string{"chicken"},
		"DBExists/Admin/chicken.exists":  false,
		"DBExists/RW/group/Admin.exists": true,

		"DestroyDB/RW/Admin/NonExistantDB.status": kivik.StatusNotFound,

		"Membership.status": kivik.StatusNotImplemented,

		"UUIDs.counts": []int{1},
		"UUIDs.status": kivik.StatusNotImplemented,

		"Log.status":          kivik.StatusNotImplemented,
		"Log/Admin/HTTP.skip": true,

		"ServerInfo.version":        `^0\.0\.1$`,
		"ServerInfo.vendor":         "Kivik",
		"ServerInfo.vendor_version": `^0\.0\.1$`,

		"Get.skip":       true, // FIXME: Unimplemented
		"Flush.skip":     true, // FIXME: Unimplemented
		"Delete.skip":    true, // FIXME: Unimplemented
		"DBInfo.skip":    true, // FIXME: Unimplemented
		"CreateDoc.skip": true, // FIXME: Unimplemented
		"Compact.skip":   true, // FIXME: Unimplemented
		"Security.skip":  true, // FIXME: Unimplemented
		"RevsLimit.skip": true, // FIXME: Unimplemented
	})
}
