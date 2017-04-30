package test

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	RegisterSuite(SuiteKivikFS, kt.SuiteConfig{
		"AllDBs.expected": []string{},

		"CreateDB/RW/NoAuth.status":         kivik.StatusUnauthorized,
		"CreateDB/RW/Admin/Recreate.status": kivik.StatusPreconditionFailed,

		"AllDocs.skip": true, // FIXME: Not yet implemented
		// "AllDocs/Admin.databases":  []string{"foo"},
		// "AllDocs/Admin/foo.status": kivik.StatusNotFound,

		"DBExists/Admin.databases":       []string{"chicken"},
		"DBExists/Admin/chicken.exists":  false,
		"DBExists/RW/group/Admin.exists": true,

		"DestroyDB/RW/Admin/NonExistantDB.status": kivik.StatusNotFound,

		"Log.status":          kivik.StatusNotImplemented,
		"Log/Admin/HTTP.skip": true,

		"Version.version":        `^0\.0\.1$`,
		"Version.vendor":         "Kivik",
		"Version.vendor_version": `^0\.0\.1$`,

		// Replications not to be implemented
		"GetReplications.skip": true,
		"Replicate.skip":       true,

		"Get.skip":               true,                       // FIXME: Unimplemented
		"Flush.skip":             true,                       // FIXME: Unimplemented
		"Delete.skip":            true,                       // FIXME: Unimplemented
		"Stats.skip":             true,                       // FIXME: Unimplemented
		"CreateDoc.skip":         true,                       // FIXME: Unimplemented
		"Compact.skip":           true,                       // FIXME: Unimplemented
		"Security.skip":          true,                       // FIXME: Unimplemented
		"DBUpdates.status":       kivik.StatusNotImplemented, // FIXME: Unimplemented
		"Changes.skip":           true,                       // FIXME: Unimplemented
		"Copy.skip":              true,                       // FIXME: Unimplemented, depends on Get/Put or Copy
		"BulkDocs.skip":          true,                       // FIXME: Unimplemented
		"GetAttachment.skip":     true,                       // FIXME: Unimplemented
		"GetAttachmentMeta.skip": true,                       // FIXME: Unimplemented
		"PutAttachment.skip":     true,                       // FIXME: Unimplemented
		"DeleteAttachment.skip":  true,                       // FIXME: Unimplemented
		"Query.skip":             true,                       // FIXME: Unimplemented
		"Find.skip":              true,                       // FIXME: Unimplemented
		"CreateIndex.skip":       true,                       // FIXME: Unimplemented
		"GetIndexes.skip":        true,                       // FIXME: Unimplemented
		"DeleteIndex.skip":       true,                       // FIXME: Unimplemented
		"Put.skip":               true,                       // FIXME: Unimplemented
		"SetSecurity.skip":       true,                       // FIXME: Unimplemented
		"ViewCleanup.skip":       true,                       // FIXME: Unimplemented
		"Rev.skip":               true,                       // FIXME: Unimplemented
	})
}
