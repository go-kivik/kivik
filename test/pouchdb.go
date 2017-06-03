// +build js

package test

import (
	"net/url"
	"os"
	"strings"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
	"github.com/gopherjs/gopherjs/js"
)

func init() {
	RegisterSuite(SuitePouchLocal, kt.SuiteConfig{
		"db": map[string]interface{}{"db": js.Global.Call("require", "memdown")},

		"PreCleanup.skip": true,

		// Features which are not supported by PouchDB
		"Log.skip":         true,
		"Flush.skip":       true,
		"Security.skip":    true, // FIXME: Perhaps implement later with a plugin?
		"SetSecurity.skip": true, // FIXME: Perhaps implement later with a plugin?
		"DBUpdates.skip":   true,

		"AllDBs.skip":   true, // FIXME: Find a way to test with the plugin
		"CreateDB.skip": true, // FIXME: No way to validate if this works unless/until allDbs works
		"DBExists.skip": true, // FIXME: Maybe fix this if/when allDBs works?

		"AllDocs/Admin.databases":                        []string{},
		"AllDocs/RW/group/Admin/WithDocs/UpdateSeq.skip": true,

		"Find/Admin.databases":                []string{},
		"Find/RW/group/Admin/Warning.warning": "no matching index found, create an index to optimize query time",

		"Query/RW/group/Admin/WithDocs/UpdateSeq.skip": true,

		"Version.version":        `^6\.\d\.\d$`,
		"Version.vendor":         `^PouchDB$`,
		"Version.vendor_version": `^6\.\d\.\d$`,

		"Get/RW/group/Admin/bogus.status":  kivik.StatusNotFound,
		"Get/RW/group/NoAuth/bogus.status": kivik.StatusNotFound,

		"Rev/RW/group/Admin/bogus.status":  kivik.StatusNotFound,
		"Rev/RW/group/NoAuth/bogus.status": kivik.StatusNotFound,

		"Delete/RW/Admin/group/MissingDoc.status":       kivik.StatusNotFound,
		"Delete/RW/Admin/group/InvalidRevFormat.status": kivik.StatusBadRequest,
		"Delete/RW/Admin/group/WrongRev.status":         kivik.StatusConflict,

		"Stats/Admin.skip": true, // No predefined DBs for Local PouchDB

		"BulkDocs/RW/Admin/group/Mix/Conflict.status": kivik.StatusConflict,

		"GetAttachment/RW/group/Admin/foo/NotFound.status": kivik.StatusNotFound,

		"GetAttachmentMeta/RW/group/Admin/foo/NotFound.status": kivik.StatusNotFound,

		"PutAttachment/RW/group/Admin/Conflict.status": kivik.StatusConflict,

		// "DeleteAttachment/RW/group/Admin/NotFound.status": kivik.StatusNotFound, // https://github.com/pouchdb/pouchdb/issues/6409
		"DeleteAttachment/RW/group/Admin/NoDoc.status": kivik.StatusNotFound,

		"Put/RW/Admin/group/LeadingUnderscoreInID.status": kivik.StatusBadRequest,
		"Put/RW/Admin/group/Conflict.status":              kivik.StatusConflict,

		"CreateIndex/RW/Admin/group/EmptyIndex.status":   kivik.StatusInternalServerError,
		"CreateIndex/RW/Admin/group/BlankIndex.status":   kivik.StatusBadRequest,
		"CreateIndex/RW/Admin/group/InvalidIndex.status": kivik.StatusInternalServerError,
		"CreateIndex/RW/Admin/group/NilIndex.status":     kivik.StatusInternalServerError,
		"CreateIndex/RW/Admin/group/InvalidJSON.status":  kivik.StatusBadRequest,

		"GetIndexes.databases": []string{},

		"DeleteIndex/RW/Admin/group/NotFoundDdoc.status": kivik.StatusNotFound,
		"DeleteIndex/RW/Admin/group/NotFoundName.status": kivik.StatusNotFound,

		"Replicate.skip": true, // No need to do this for both Local and Remote

		"Query/RW/group/Admin/WithoutDocs/ScanDoc.status": kivik.StatusBadRequest,
	})
	RegisterSuite(SuitePouchRemote, kt.SuiteConfig{
		// Features which are not supported by PouchDB
		"Log.skip":         true,
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
		"AllDocs/NoAuth/_replicator.status":                  kivik.StatusUnauthorized,
		"AllDocs/NoAuth/_users.status":                       kivik.StatusUnauthorized,
		"AllDocs/NoAuth/chicken.status":                      kivik.StatusNotFound,
		"AllDocs/Admin/_replicator/WithDocs/UpdateSeq.skip":  true,
		"AllDocs/Admin/_users/WithDocs/UpdateSeq.skip":       true,
		"AllDocs/RW/group/Admin/WithDocs/UpdateSeq.skip":     true,
		"AllDocs/RW/group/Admin/WithoutDocs/UpdateSeq.skip":  true,
		"AllDocs/RW/group/NoAuth/WithDocs/UpdateSeq.skip":    true,
		"AllDocs/RW/group/NoAuth/WithoutDocs/UpdateSeq.skip": true,

		"Find.skip":        true, // Find doesn't work with CouchDB 1.6, which we use for these tests
		"CreateIndex.skip": true, // Find doesn't work with CouchDB 1.6, which we use for these tests
		"GetIndexes.skip":  true, // Find doesn't work with CouchDB 1.6, which we use for these tests
		"DeleteIndex.skip": true, // Find doesn't work with CouchDB 1.6, which we use for these tests

		"Query/RW/group/Admin/WithDocs/UpdateSeq.skip":  true,
		"Query/RW/group/NoAuth/WithDocs/UpdateSeq.skip": true,

		"Version.version":        `^6\.\d\.\d$`,
		"Version.vendor":         `^PouchDB$`,
		"Version.vendor_version": `^6\.\d\.\d$`,

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
		"Delete/RW/NoAuth/group/DesignDoc.status":        kivik.StatusUnauthorized,

		"Stats.databases":             []string{"_users", "chicken"},
		"Stats/Admin/chicken.status":  kivik.StatusNotFound,
		"Stats/NoAuth/chicken.status": kivik.StatusNotFound,

		"BulkDocs/RW/NoAuth/group/Mix/Conflict.status": kivik.StatusConflict,
		"BulkDocs/RW/Admin/group/Mix/Conflict.status":  kivik.StatusConflict,

		"GetAttachment/RW/group/Admin/foo/NotFound.status":  kivik.StatusNotFound,
		"GetAttachment/RW/group/NoAuth/foo/NotFound.status": kivik.StatusNotFound,

		"GetAttachmentMeta/RW/group/Admin/foo/NotFound.status":  kivik.StatusNotFound,
		"GetAttachmentMeta/RW/group/NoAuth/foo/NotFound.status": kivik.StatusNotFound,

		"PutAttachment/RW/group/Admin/Conflict.status":         kivik.StatusInternalServerError, // Couch 2.0 bug
		"PutAttachment/RW/group/NoAuth/Conflict.status":        kivik.StatusInternalServerError, // Couch 2.0 bug
		"PutAttachment/RW/group/NoAuth/UpdateDesignDoc.status": kivik.StatusUnauthorized,
		"PutAttachment/RW/group/NoAuth/CreateDesignDoc.status": kivik.StatusUnauthorized,

		// "DeleteAttachment/RW/group/Admin/NotFound.status":  kivik.StatusNotFound, // COUCHDB-3362
		// "DeleteAttachment/RW/group/NoAuth/NotFound.status": kivik.StatusNotFound, // COUCHDB-3362
		"DeleteAttachment/RW/group/Admin/NoDoc.status":      kivik.StatusInternalServerError, // Couch 2.0 bug
		"DeleteAttachment/RW/group/NoAuth/NoDoc.status":     kivik.StatusInternalServerError, // Couch 2.0 bug
		"DeleteAttachment/RW/group/NoAuth/DesignDoc.status": kivik.StatusUnauthorized,

		"Put/RW/Admin/group/LeadingUnderscoreInID.status":  kivik.StatusBadRequest,
		"Put/RW/Admin/group/Conflict.status":               kivik.StatusConflict,
		"Put/RW/NoAuth/group/DesignDoc.status":             kivik.StatusUnauthorized,
		"Put/RW/NoAuth/group/LeadingUnderscoreInID.status": kivik.StatusBadRequest,
		"Put/RW/NoAuth/group/Conflict.status":              kivik.StatusConflict,

		"Replicate.NotFoundDB": func() string {
			var dsn string
			for _, env := range []string{"KIVIK_TEST_DSN_COUCH16", "KIVIK_TEST_DSN_COUCH20", "KIVIK_TEST_DSN_CLOUDANT"} {
				dsn = os.Getenv(env)
				if dsn != "" {
					break
				}
			}
			parsed, _ := url.Parse(dsn)
			parsed.User = nil
			return strings.TrimSuffix(parsed.String(), "/") + "/doesntexist"

		}(),
		"Replicate.prefix":                                       "none",
		"Replicate.timeoutSeconds":                               5,
		"Replicate.mode":                                         "pouchdb",
		"Replicate/RW/Admin/group/MissingSource/Results.status":  kivik.StatusUnauthorized,
		"Replicate/RW/Admin/group/MissingTarget/Results.status":  kivik.StatusUnauthorized,
		"Replicate/RW/NoAuth/group/MissingSource/Results.status": kivik.StatusUnauthorized,
		"Replicate/RW/NoAuth/group/MissingTarget/Results.status": kivik.StatusUnauthorized,

		"Query/RW/group/Admin/WithoutDocs/ScanDoc.status":  kivik.StatusBadRequest,
		"Query/RW/group/NoAuth/WithoutDocs/ScanDoc.status": kivik.StatusBadRequest,
	})
}
