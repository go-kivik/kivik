package test

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	RegisterSuite(SuiteKivikMemory, kt.SuiteConfig{
		// Unsupported features
		"Flush.skip": true,

		"AllDBs.expected": []string{},

		"Config.status": kivik.StatusNotImplemented,

		"CreateDB/RW/NoAuth.status":         kivik.StatusUnauthorized,
		"CreateDB/RW/Admin/Recreate.status": kivik.StatusPreconditionFailed,

		"AllDocs/Admin.databases":  []string{"foo"},
		"AllDocs/Admin/foo.status": kivik.StatusNotFound,
		"AllDocs/RW.skip":          true, // FIXME: Update this when the memory driver can create documents

		"DBExists/Admin.databases":       []string{"chicken"},
		"DBExists/Admin/chicken.exists":  false,
		"DBExists/RW/group/Admin.exists": true,

		"DestroyDB/RW/Admin/NonExistantDB.status": kivik.StatusNotFound,

		"Membership.status": kivik.StatusNotImplemented,

		"UUIDs.counts":               []int{-1, 0, 1, 10},
		"UUIDs/Admin/-1Count.status": kivik.StatusBadRequest,
		// "UUIDs.status": kivik.StatusNotImplemented,

		"Log.status":          kivik.StatusNotImplemented,
		"Log/Admin/HTTP.skip": true,

		"ServerInfo.version":        `^0\.0\.1$`,
		"ServerInfo.vendor":         `^Kivik Memory Adaptor$`,
		"ServerInfo.vendor_version": `^0\.0\.1$`,

		"Get.skip":       true, // FIXME: Unimplemented
		"Delete.skip":    true, // FIXME: Unimplemented
		"DBInfo.skip":    true, // FIXME: Unimplemented
		"CreateDoc.skip": true, // FIXME: Unimplemented
	})
}
