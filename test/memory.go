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

		"AllDocs.skip": true, // FIXME: Unimplemented

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

		"Get.skip":               true,                       // FIXME: Unimplemented
		"Delete.skip":            true,                       // FIXME: Unimplemented
		"DBInfo.skip":            true,                       // FIXME: Unimplemented
		"CreateDoc.skip":         true,                       // FIXME: Unimplemented
		"Compact.skip":           true,                       // FIXME: Unimplemented
		"Security.skip":          true,                       // FIXME: Unimplemented
		"RevsLimit.skip":         true,                       // FIXME: Unimplemented
		"DBUpdates.status":       kivik.StatusNotImplemented, // FIXME: Unimplemented
		"Changes.skip":           true,                       // FIXME: Unimplemented
		"Copy.skip":              true,                       // FIXME: Unimplemented, depends on Get/Put or Copy
		"BulkDocs.skip":          true,                       // FIXME: Unimplemented
		"GetAttachment.skip":     true,                       // FIXME: Unimplemented
		"GetAttachmentMeta.skip": true,                       // FIXME: Unimplemented
		"PutAttachment.skip":     true,                       // FIXME: Unimplemented
	})
}
