// +build js

package test

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	RegisterSuite(SuitePouchLocal, kt.SuiteConfig{
		// Features which are not supported by PouchDB
		"UUIDs.skip":      true,
		"Log.skip":        true,
		"Membership.skip": true,
		"Config.skip":     true,
		"Flush.skip":      true,

		"AllDBs.skip":   true, // FIXME: Find a way to test with the plugin
		"CreateDB.skip": true, // FIXME: No way to validate if this works unless/until allDbs works
		"DBExists.skip": true, // FIXME: Maybe fix this if/when allDBs works?

		"AllDocs/Admin.skip": true,
		"AllDocs/RW.skip":    true, // FIXME: Not sure why this is broken

		"ServerInfo.version":        `^6\.\d\.\d$`,
		"ServerInfo.vendor":         `^PouchDB$`,
		"ServerInfo.vendor_version": `^6\.\d\.\d$`,

		"Get.skip": true, // FIXME: Update this when Get is implemented

		"Delete.skip": true, // FIXME: Unimplemented
	})
	RegisterSuite(SuitePouchRemote, kt.SuiteConfig{
		// Features which are not supported by PouchDB
		"UUIDs.skip":      true,
		"Log.skip":        true,
		"Membership.skip": true,
		"Config.skip":     true,
		"Flush.skip":      true,
		"Session.skip":    true,

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
		"AllDocs/RW.skip":                    true, // FIXME: Not sure why this is broken

		"ServerInfo.version":        `^6\.\d\.\d$`,
		"ServerInfo.vendor":         `^PouchDB$`,
		"ServerInfo.vendor_version": `^6\.\d\.\d$`,

		"Get.skip": true, // FIXME: Update this when Get is implemented

		"Delete.skip": true, // FIXME: Unimplemented

		"DBInfo.skip": true, // FIXME: Implement Put() and Delete() first
	})
}
