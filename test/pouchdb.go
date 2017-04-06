// +build js

package test

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	RegisterSuite(SuitePouchLocal, kt.SuiteConfig{
		"PreCleanup.skip": true,

		// Features which are not supported by PouchDB
		"UUIDs.skip":       true,
		"Log.skip":         true,
		"Membership.skip":  true,
		"Config.skip":      true,
		"Flush.skip":       true,
		"Security.skip":    true, // FIXME: Perhaps implement later with a plugin?
		"SetSecurity.skip": true, // FIXME: Perhaps implement later with a plugin?
		"DBUpdates.skip":   true,

		"AllDBs.skip":   true, // FIXME: Find a way to test with the plugin
		"CreateDB.skip": true, // FIXME: No way to validate if this works unless/until allDbs works
		"DBExists.skip": true, // FIXME: Maybe fix this if/when allDBs works?

		"AllDocs/Admin.databases":                        []string{},
		"AllDocs/RW/group/Admin/WithDocs/UpdateSeq.skip": true,

		"ServerInfo.version":        `^6\.\d\.\d$`,
		"ServerInfo.vendor":         `^PouchDB$`,
		"ServerInfo.vendor_version": `^6\.\d\.\d$`,

		"Get/RW/group/Admin/bogus.status":  kivik.StatusNotFound,
		"Get/RW/group/NoAuth/bogus.status": kivik.StatusNotFound,

		"Rev/RW/group/Admin/bogus.status":  kivik.StatusNotFound,
		"Rev/RW/group/NoAuth/bogus.status": kivik.StatusNotFound,

		"Delete/RW/Admin/group/MissingDoc.status":       kivik.StatusNotFound,
		"Delete/RW/Admin/group/InvalidRevFormat.status": kivik.StatusBadRequest,
		"Delete/RW/Admin/group/WrongRev.status":         kivik.StatusConflict,

		"DBInfo/Admin.skip": true, // No predefined DBs for Local PouchDB

		"RevsLimit.databases": []string{},

		"BulkDocs/RW/Admin/group/Mix/Conflict.status": kivik.StatusConflict,

		"GetAttachment/RW/group/Admin/NotFound.status": kivik.StatusNotFound,

		"PutAttachment/RW/group/Admin/Conflict.status": kivik.StatusConflict,
	})
	RegisterSuite(SuitePouchRemote, kt.SuiteConfig{
		// Features which are not supported by PouchDB
		"UUIDs.skip":       true,
		"Log.skip":         true,
		"Membership.skip":  true,
		"Config.skip":      true,
		"Flush.skip":       true,
		"Session.skip":     true,
		"Security.skip":    true, // FIXME: Perhaps implement later with a plugin?
		"SetSecurity.skip": true, // FIXME: Perhaps implement later with a plugin?
		"DBUpdates.skip":   true,

		"PreCleanup.skip": true,

		"AllDBs.skip": true, // FIXME: Perhaps a workaround can be found?

		"CreateDB/RW/NoAuth.status":         kivik.StatusUnauthorized,
		"CreateDB/RW/Admin/Recreate.status": kivik.StatusPreconditionFailed,

		"DBExists.databases":              []string{"_users", "chicken"},
		"DBExists/Admin/_users.exists":    true,
		"DBExists/Admin/chicken.exists":   false,
		"DBExists/NoAuth/_users.exists":   true,
		"DBExists/NoAuth/chicken.exists":  false,
		"DBExists/RW/group/Admin.exists":  true,
		"DBExists/RW/group/NoAuth.exists": true,

		"DestroyDB/RW/NoAuth/NonExistantDB.status": kivik.StatusNotFound,
		"DestroyDB/RW/Admin/NonExistantDB.status":  kivik.StatusNotFound,
		"DestroyDB/RW/NoAuth/ExistingDB.status":    kivik.StatusUnauthorized,

		"AllDocs.databases":                                  []string{"_replicator", "_users", "chicken"},
		"AllDocs/Admin/_replicator.expected":                 []string{"_design/_replicator"},
		"AllDocs/Admin/_replicator.offset":                   0,
		"AllDocs/Admin/_users.expected":                      []string{"_design/_auth"},
		"AllDocs/Admin/chicken.status":                       kivik.StatusNotFound,
		"AllDocs/NoAuth/_replicator.status":                  kivik.StatusForbidden,
		"AllDocs/NoAuth/_users.status":                       kivik.StatusForbidden,
		"AllDocs/NoAuth/chicken.status":                      kivik.StatusNotFound,
		"AllDocs/Admin/_replicator/WithDocs/UpdateSeq.skip":  true,
		"AllDocs/Admin/_users/WithDocs/UpdateSeq.skip":       true,
		"AllDocs/RW/group/Admin/WithDocs/UpdateSeq.skip":     true,
		"AllDocs/RW/group/Admin/WithoutDocs/UpdateSeq.skip":  true,
		"AllDocs/RW/group/NoAuth/WithDocs/UpdateSeq.skip":    true,
		"AllDocs/RW/group/NoAuth/WithoutDocs/UpdateSeq.skip": true,

		"ServerInfo.version":        `^6\.\d\.\d$`,
		"ServerInfo.vendor":         `^PouchDB$`,
		"ServerInfo.vendor_version": `^6\.\d\.\d$`,

		"Get/RW/group/Admin/bogus.status":  kivik.StatusNotFound,
		"Get/RW/group/NoAuth/bogus.status": kivik.StatusNotFound,

		"Rev/RW/group/Admin/bogus.status":  kivik.StatusNotFound,
		"Rev/RW/group/NoAuth/bogus.status": kivik.StatusNotFound,

		"Delete/RW/Admin/group/MissingDoc.status":        kivik.StatusNotFound,
		"Delete/RW/Admin/group/InvalidRevFormat.status":  kivik.StatusBadRequest,
		"Delete/RW/Admin/group/WrongRev.status":          kivik.StatusConflict,
		"Delete/RW/NoAuth/group/MissingDoc.status":       kivik.StatusNotFound,
		"Delete/RW/NoAuth/group/InvalidRevFormat.status": kivik.StatusBadRequest,
		"Delete/RW/NoAuth/group/WrongRev.status":         kivik.StatusConflict,

		"DBInfo.databases":             []string{"_users", "chicken"},
		"DBInfo/Admin/chicken.status":  kivik.StatusNotFound,
		"DBInfo/NoAuth/chicken.status": kivik.StatusNotFound,

		"RevsLimit.skip": true, // FIXME: Unsupported for remote databases. Perhaps later with a plugin?

		"BulkDocs/RW/NoAuth/group/Mix/Conflict.status": kivik.StatusConflict,
		"BulkDocs/RW/Admin/group/Mix/Conflict.status":  kivik.StatusConflict,

		"GetAttachment/RW/group/Admin/NotFound.status":  kivik.StatusNotFound,
		"GetAttachment/RW/group/NoAuth/NotFound.status": kivik.StatusNotFound,

		"PutAttachment/RW/group/Admin/Conflict.status":  kivik.StatusConflict,
		"PutAttachment/RW/group/NoAuth/Conflict.status": kivik.StatusConflict,
	})
}
