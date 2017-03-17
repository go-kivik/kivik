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
		"UUIDs.skip":      true,
		"Log.skip":        true,
		"Membership.skip": true,
		"Config.skip":     true,
		"Flush.skip":      true,
		"Security.skip":   true, // FIXME: Perhaps implement later with a plugin?

		"AllDBs.skip":   true, // FIXME: Find a way to test with the plugin
		"CreateDB.skip": true, // FIXME: No way to validate if this works unless/until allDbs works
		"DBExists.skip": true, // FIXME: Maybe fix this if/when allDBs works?

		"AllDocs/Admin.databases": []string{},

		"ServerInfo.version":        `^6\.\d\.\d$`,
		"ServerInfo.vendor":         `^PouchDB$`,
		"ServerInfo.vendor_version": `^6\.\d\.\d$`,

		"Get/RW/group/Admin/bogus.status":  kivik.StatusNotFound,
		"Get/RW/group/NoAuth/bogus.status": kivik.StatusNotFound,

		"Delete/RW/Admin/group/MissingDoc.status":       kivik.StatusNotFound,
		"Delete/RW/Admin/group/InvalidRevFormat.status": kivik.StatusBadRequest,
		"Delete/RW/Admin/group/WrongRev.status":         kivik.StatusConflict,

		"DBInfo/Admin.skip": true, // No predefined DBs for Local PouchDB
	})
	RegisterSuite(SuitePouchRemote, kt.SuiteConfig{
		// Features which are not supported by PouchDB
		"UUIDs.skip":      true,
		"Log.skip":        true,
		"Membership.skip": true,
		"Config.skip":     true,
		"Flush.skip":      true,
		"Session.skip":    true,
		"Security.skip":   true, // FIXME: Perhaps implement later with a plugin?

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

		"AllDocs/Admin.databases":            []string{"_replicator"},
		"AllDocs/Admin/_replicator.expected": []string{"_design/_replicator"},
		"AllDocs/Admin/_replicator.offset":   0,
		"AllDocs/NoAuth.databases":           []string{"_replicator"},
		"AllDocs/NoAuth/_replicator.status":  kivik.StatusForbidden,

		"ServerInfo.version":        `^6\.\d\.\d$`,
		"ServerInfo.vendor":         `^PouchDB$`,
		"ServerInfo.vendor_version": `^6\.\d\.\d$`,

		"Get/RW/group/Admin/bogus.status":  kivik.StatusNotFound,
		"Get/RW/group/NoAuth/bogus.status": kivik.StatusNotFound,

		"Delete/RW/Admin/group/MissingDoc.status":        kivik.StatusNotFound,
		"Delete/RW/Admin/group/InvalidRevFormat.status":  kivik.StatusBadRequest,
		"Delete/RW/Admin/group/WrongRev.status":          kivik.StatusConflict,
		"Delete/RW/NoAuth/group/MissingDoc.status":       kivik.StatusNotFound,
		"Delete/RW/NoAuth/group/InvalidRevFormat.status": kivik.StatusBadRequest,
		"Delete/RW/NoAuth/group/WrongRev.status":         kivik.StatusConflict,

		"DBInfo.databases": []string{"_users"},
	})
}
