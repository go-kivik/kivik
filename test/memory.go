package test

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	RegisterSuite(SuiteKivikMemory, kt.SuiteConfig{
		// Unsupported features
		"Flush.skip": true,

		"AllDBs.expected": []string{"_users"},

		"CreateDB/RW/NoAuth.status":         kivik.StatusUnauthorized,
		"CreateDB/RW/Admin/Recreate.status": kivik.StatusPreconditionFailed,

		"AllDocs.skip": true, // FIXME: Unimplemented

		"DBExists/Admin.databases":       []string{"chicken"},
		"DBExists/Admin/chicken.exists":  false,
		"DBExists/RW/group/Admin.exists": true,

		"DestroyDB/RW/Admin/NonExistantDB.status": kivik.StatusNotFound,

		"Log.status":          kivik.StatusNotImplemented,
		"Log/Admin/HTTP.skip": true,

		"Version.version":        `^0\.0\.1$`,
		"Version.vendor":         `^Kivik Memory Adaptor$`,
		"Version.vendor_version": `^0\.0\.1$`,

		// Replications not to be implemented
		"GetReplications.skip": true,
		"Replicate.skip":       true,

		"Get/RW/group/Admin/bogus.status": kivik.StatusNotFound,

		"Rev/RW/group/Admin/bogus.status": kivik.StatusNotFound,

		"Put/RW/Admin/group/LeadingUnderscoreInID.status": kivik.StatusBadRequest,
		"Put/RW/Admin/group/Conflict.status":              kivik.StatusConflict,

		"Delete/RW/Admin/group/MissingDoc.status":       kivik.StatusNotFound,
		"Delete/RW/Admin/group/InvalidRevFormat.status": kivik.StatusBadRequest,
		"Delete/RW/Admin/group/WrongRev.status":         kivik.StatusConflict,

		"Security.databases":            []string{"_users", "chicken", "_duck"},
		"Security/Admin/chicken.status": kivik.StatusNotFound,
		"Security/Admin/_duck.status":   kivik.StatusNotFound,

		"SetSecurity/RW/Admin/NotExists.status": kivik.StatusNotFound,

		"BulkDocs/RW/NoAuth/group/Mix/Conflict.status": kivik.StatusConflict,
		"BulkDocs/RW/Admin/group/Mix/Conflict.status":  kivik.StatusConflict,

		"Find.databases":                       []string{"chicken", "_duck"},
		"Find/Admin/chicken.status":            kivik.StatusNotFound,
		"Find/Admin/_duck.status":              kivik.StatusNotFound,
		"Find/NoAuth/chicken.status":           kivik.StatusNotFound,
		"Find/NoAuth/_duck.status":             kivik.StatusUnauthorized,
		"Find/RW/group/Admin/Warning.warning":  "no matching index found, create an index to optimize query time",
		"Find/RW/group/NoAuth/Warning.warning": "no matching index found, create an index to optimize query time",

		"Stats.skip":             true,                       // FIXME: Unimplemented
		"Compact.skip":           true,                       // FIXME: Unimplemented
		"DBUpdates.status":       kivik.StatusNotImplemented, // FIXME: Unimplemented
		"Changes.skip":           true,                       // FIXME: Unimplemented
		"Copy.skip":              true,                       // FIXME: Unimplemented, depends on Get/Put or Copy
		"GetAttachment.skip":     true,                       // FIXME: Unimplemented
		"GetAttachmentMeta.skip": true,                       // FIXME: Unimplemented
		"PutAttachment.skip":     true,                       // FIXME: Unimplemented
		"DeleteAttachment.skip":  true,                       // FIXME: Unimplemented
		"Query.skip":             true,                       // FIXME: Unimplemented
		"CreateIndex.skip":       true,                       // FIXME: Unimplemented
		"GetIndexes.skip":        true,                       // FIXME: Unimplemented
		"DeleteIndex.skip":       true,                       // FIXME: Unimplemented
		"SetSecurity.skip":       true,                       // FIXME: Unimplemented
		"ViewCleanup.skip":       true,                       // FIXME: Unimplemented
	})
}
