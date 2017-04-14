package test

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	RegisterSuite(SuiteCouch20, kt.SuiteConfig{
		"AllDBs.expected": []string{"_global_changes", "_replicator", "_users"},

		"Config.skip": true, // FIXME: CouchDB config has moved
		"Config/Admin/GetAll.expected_sections": []string{"admins", "attachments", "compaction_daemon", "cors", "couch_httpd_auth",
			"couch_httpd_oauth", "couchdb", "daemons", "database_compaction", "httpd", "httpd_db_handlers", "httpd_design_handlers",
			"httpd_global_handlers", "log", "query_server_config", "query_servers", "replicator", "ssl", "stats", "uuids", "vendor",
			"view_compaction"},
		"Config/Admin/GetSection.sections":                             []string{"log", "chicken"},
		"Config/Admin/GetSection/log.keys":                             []string{"file", "include_sasl", "level"},
		"Config/Admin/GetSection/chicken.keys":                         []string{},
		"Config/Admin/GetItem.items":                                   []string{"log.level", "log.foobar", "logx.level"},
		"Config/Admin/GetItem/log.foobar.status":                       kivik.StatusNotFound,
		"Config/Admin/GetItem/logx.level.status":                       kivik.StatusNotFound,
		"Config/Admin/GetItem/log.level.expected":                      "info",
		"Config/NoAuth/GetAll.status":                                  kivik.StatusUnauthorized,
		"Config/NoAuth/GetSection.sections":                            []string{"log", "chicken"},
		"Config/NoAuth/GetSection.status":                              kivik.StatusUnauthorized,
		"Config/NoAuth/GetItem.items":                                  []string{"log.level", "log.foobar", "logx.level"},
		"Config/NoAuth/GetItem.status":                                 kivik.StatusUnauthorized,
		"Config/RW/group/NoAuth/Set.status":                            kivik.StatusUnauthorized,
		"Config/RW/group/NoAuth/Delete/group.status":                   kivik.StatusUnauthorized,
		"Config/RW/group/Admin/Delete/group/NonExistantKey.status":     kivik.StatusNotFound,
		"Config/RW/group/Admin/Delete/group/NonExistantSection.status": kivik.StatusNotFound,

		"CreateDB/RW/NoAuth.status":         kivik.StatusUnauthorized,
		"CreateDB/RW/Admin/Recreate.status": kivik.StatusPreconditionFailed,

		"DestroyDB/RW/NoAuth.status":              kivik.StatusUnauthorized,
		"DestroyDB/RW/Admin/NonExistantDB.status": kivik.StatusNotFound,

		"AllDocs.databases":                  []string{"_replicator", "chicken", "_duck"},
		"AllDocs/Admin/_replicator.expected": []string{"_design/_replicator"},
		"AllDocs/Admin/_replicator.offset":   0,
		"AllDocs/Admin/chicken.status":       kivik.StatusNotFound,
		"AllDocs/Admin/_duck.status":         kivik.StatusNotFound,
		"AllDocs/NoAuth/_replicator.status":  kivik.StatusUnauthorized,
		"AllDocs/NoAuth/chicken.status":      kivik.StatusNotFound,
		"AllDocs/NoAuth/_duck.status":        kivik.StatusUnauthorized,

		"Find.databases":                   []string{"_replicator", "chicken", "_duck"},
		"Find/Admin/_replicator.expected":  []string{"_design/_replicator"},
		"Find/Admin/chicken.status":        kivik.StatusNotFound,
		"Find/Admin/_duck.status":          kivik.StatusNotFound,
		"Find/NoAuth/_replicator.expected": []string{"_design/_replicator"},
		"Find/NoAuth/chicken.status":       kivik.StatusNotFound,
		"Find/NoAuth/_duck.status":         kivik.StatusUnauthorized,

		"DBExists.databases":              []string{"_users", "chicken", "_duck"},
		"DBExists/Admin/_users.exists":    true,
		"DBExists/Admin/chicken.exists":   false,
		"DBExists/Admin/_duck.exists":     false,
		"DBExists/NoAuth/_users.exists":   true,
		"DBExists/NoAuth/chicken.exists":  false,
		"DBExists/NoAuth/_duck.status":    kivik.StatusUnauthorized,
		"DBExists/RW/group/Admin.exists":  true,
		"DBExists/RW/group/NoAuth.exists": true,

		"Membership/Admin.all":      []string{"nonode@nohost"},
		"Membership/NoAuth.all":     []string{"nonode@nohost"},
		"Membership/Admin.cluster":  []string{"nonode@nohost"},
		"Membership/NoAuth.cluster": []string{"nonode@nohost"},

		"UUIDs.counts":                []int{-1, 0, 1, 10},
		"UUIDs/Admin/-1Count.status":  kivik.StatusBadRequest,
		"UUIDs/NoAuth/-1Count.status": kivik.StatusBadRequest,

		"Log.skip": true, // This was removed in CouchDB 2.0

		"ServerInfo.version":        `^2\.0\.0$`,
		"ServerInfo.vendor":         `^The Apache Software Foundation$`,
		"ServerInfo.vendor_version": ``, // CouchDB 2.0 no longer has a vendor version

		"Get/RW/group/Admin/bogus.status":  kivik.StatusNotFound,
		"Get/RW/group/NoAuth/bogus.status": kivik.StatusNotFound,

		"Rev/RW/group/Admin/bogus.status":  kivik.StatusNotFound,
		"Rev/RW/group/NoAuth/bogus.status": kivik.StatusNotFound,

		"Flush.databases":                            []string{"_users", "chicken", "_duck"},
		"Flush/NoAuth/chicken/DoFlush.status":        kivik.StatusNotFound,
		"Flush/Admin/chicken/DoFlush.status":         kivik.StatusNotFound,
		"Flush/Admin/_users/DoFlush/Timestamp.skip":  true, // CouchDB 2.0 always returns 0?
		"Flush/Admin/_duck/DoFlush.status":           kivik.StatusNotFound,
		"Flush/NoAuth/_users/DoFlush/Timestamp.skip": true, // CouchDB 2.0 always returns 0?
		"Flush/NoAuth/_duck/DoFlush.status":          kivik.StatusUnauthorized,

		"Delete/RW/Admin/group/MissingDoc.status":        kivik.StatusNotFound,
		"Delete/RW/Admin/group/InvalidRevFormat.status":  kivik.StatusBadRequest,
		"Delete/RW/Admin/group/WrongRev.status":          kivik.StatusConflict,
		"Delete/RW/NoAuth/group/MissingDoc.status":       kivik.StatusNotFound,
		"Delete/RW/NoAuth/group/InvalidRevFormat.status": kivik.StatusBadRequest,
		"Delete/RW/NoAuth/group/WrongRev.status":         kivik.StatusConflict,
		"Delete/RW/NoAuth/group/DesignDoc.status":        kivik.StatusUnauthorized,

		"Session/Get/Admin.info.authentication_handlers":  "cookie,default",
		"Session/Get/Admin.info.authentication_db":        "_users",
		"Session/Get/Admin.info.authenticated":            "cookie",
		"Session/Get/Admin.userCtx.roles":                 "_admin",
		"Session/Get/Admin.ok":                            "true",
		"Session/Get/NoAuth.info.authentication_handlers": "cookie,default",
		"Session/Get/NoAuth.info.authentication_db":       "_users",
		"Session/Get/NoAuth.info.authenticated":           "",
		"Session/Get/NoAuth.userCtx.roles":                "",
		"Session/Get/NoAuth.ok":                           "true",

		"Session/Post/EmptyJSON.status":                             kivik.StatusBadRequest,
		"Session/Post/BogusTypeJSON.status":                         kivik.StatusBadRequest,
		"Session/Post/BogusTypeForm.status":                         kivik.StatusBadRequest,
		"Session/Post/EmptyForm.status":                             kivik.StatusBadRequest,
		"Session/Post/BadJSON.status":                               kivik.StatusBadRequest,
		"Session/Post/BadForm.status":                               kivik.StatusBadRequest,
		"Session/Post/MeaninglessJSON.status":                       kivik.StatusInternalServerError,
		"Session/Post/MeaninglessForm.status":                       kivik.StatusBadRequest,
		"Session/Post/GoodJSON.status":                              kivik.StatusUnauthorized,
		"Session/Post/BadQueryParam.status":                         kivik.StatusUnauthorized,
		"Session/Post/BadCredsJSON.status":                          kivik.StatusUnauthorized,
		"Session/Post/BadCredsForm.status":                          kivik.StatusUnauthorized,
		"Session/Post/GoodCredsJSONRemoteRedirHeaderInjection.skip": true, // CouchDB allows header injection
		"Session/Post/GoodCredsJSONRemoteRedirInvalidURL.skip":      true, // CouchDB doesn't sanitize the Location value, so sends unparseable headers.

		"DBInfo.databases":             []string{"_users", "chicken", "_duck"},
		"DBInfo/Admin/chicken.status":  kivik.StatusNotFound,
		"DBInfo/Admin/_duck.status":    kivik.StatusNotFound,
		"DBInfo/NoAuth/chicken.status": kivik.StatusNotFound,
		"DBInfo/NoAuth/_duck.status":   kivik.StatusUnauthorized,

		"Compact.skip": false,

		"Security.databases":                     []string{"_replicator", "_users", "_global_changes", "chicken", "_duck"},
		"Security/Admin/chicken.status":          kivik.StatusNotFound,
		"Security/Admin/_duck.status":            kivik.StatusNotFound,
		"Security/NoAuth/_global_changes.status": kivik.StatusUnauthorized,
		"Security/NoAuth/chicken.status":         kivik.StatusNotFound,
		"Security/NoAuth/_duck.status":           kivik.StatusUnauthorized,
		"Security/RW/group/NoAuth.status":        kivik.StatusUnauthorized,

		"SetSecurity/RW/Admin/NotExists.status":  kivik.StatusNotFound,
		"SetSecurity/RW/NoAuth/NotExists.status": kivik.StatusNotFound,
		"SetSecurity/RW/NoAuth/Exists.status":    kivik.StatusInternalServerError, // Can you say WTF?

		"RevsLimit.databases":                     []string{"_replicator", "_users", "_global_changes", "chicken", "_duck"},
		"RevsLimit.revs_limit":                    1000,
		"RevsLimit/Admin/chicken.status":          kivik.StatusNotFound,
		"RevsLimit/Admin/_duck.status":            kivik.StatusNotFound,
		"RevsLimit/NoAuth/_global_changes.status": kivik.StatusUnauthorized,
		"RevsLimit/NoAuth/chicken.status":         kivik.StatusNotFound,
		"RevsLimit/NoAuth/_duck.status":           kivik.StatusUnauthorized,
		"RevsLimit/RW/NoAuth/Set.status":          kivik.StatusInternalServerError, // Stupid bug in Couch 2.0

		"DBUpdates/RW/NoAuth.status": kivik.StatusUnauthorized,

		"BulkDocs/RW/NoAuth/group/Mix/Conflict.status": kivik.StatusConflict,
		"BulkDocs/RW/Admin/group/Mix/Conflict.status":  kivik.StatusConflict,

		"GetAttachment/RW/group/Admin/foo/NotFound.status":  kivik.StatusNotFound,
		"GetAttachment/RW/group/NoAuth/foo/NotFound.status": kivik.StatusNotFound,

		"GetAttachmentMeta/RW/group/Admin/foo/NotFound.status":  kivik.StatusNotFound,
		"GetAttachmentMeta/RW/group/NoAuth/foo/NotFound.status": kivik.StatusNotFound,

		"PutAttachment/RW/group/Admin/Conflict.status":         kivik.StatusInternalServerError, // COUCHDB-3361
		"PutAttachment/RW/group/NoAuth/Conflict.status":        kivik.StatusInternalServerError, // COUCHDB-3361
		"PutAttachment/RW/group/NoAuth/UpdateDesignDoc.status": kivik.StatusUnauthorized,
		"PutAttachment/RW/group/NoAuth/CreateDesignDoc.status": kivik.StatusUnauthorized,

		// "DeleteAttachment/RW/group/Admin/NotFound.status":  kivik.StatusNotFound, // COUCHDB-3362
		// "DeleteAttachment/RW/group/NoAuth/NotFound.status": kivik.StatusNotFound, // COUCHDB-3362
		"DeleteAttachment/RW/group/Admin/NoDoc.status":      kivik.StatusInternalServerError,
		"DeleteAttachment/RW/group/NoAuth/NoDoc.status":     kivik.StatusInternalServerError,
		"DeleteAttachment/RW/group/NoAuth/DesignDoc.status": kivik.StatusUnauthorized,

		"Put/RW/Admin/group/LeadingUnderscoreInID.status":  kivik.StatusBadRequest,
		"Put/RW/Admin/group/Conflict.status":               kivik.StatusConflict,
		"Put/RW/NoAuth/group/LeadingUnderscoreInID.status": kivik.StatusBadRequest,
		"Put/RW/NoAuth/group/DesignDoc.status":             kivik.StatusUnauthorized,
		"Put/RW/NoAuth/group/Conflict.status":              kivik.StatusConflict,

		"CreateIndex/RW/Admin/group/EmptyIndex.status":    kivik.StatusBadRequest,
		"CreateIndex/RW/Admin/group/BlankIndex.status":    kivik.StatusBadRequest,
		"CreateIndex/RW/Admin/group/InvalidIndex.status":  kivik.StatusBadRequest,
		"CreateIndex/RW/Admin/group/NilIndex.status":      kivik.StatusBadRequest,
		"CreateIndex/RW/Admin/group/InvalidJSON.status":   kivik.StatusBadRequest,
		"CreateIndex/RW/NoAuth/group/EmptyIndex.status":   kivik.StatusBadRequest,
		"CreateIndex/RW/NoAuth/group/BlankIndex.status":   kivik.StatusBadRequest,
		"CreateIndex/RW/NoAuth/group/InvalidIndex.status": kivik.StatusBadRequest,
		"CreateIndex/RW/NoAuth/group/NilIndex.status":     kivik.StatusBadRequest,
		"CreateIndex/RW/NoAuth/group/InvalidJSON.status":  kivik.StatusBadRequest,
		"CreateIndex/RW/NoAuth/group/Valid.status":        kivik.StatusInternalServerError, // COUCHDB-3374

		"GetIndexes.databases":                      []string{"_replicator", "_users", "_global_changes", "chicken", "_duck"},
		"GetIndexes/Admin/_replicator.indexes":      []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/Admin/_users.indexes":           []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/Admin/_global_changes.indexes":  []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/Admin/chicken.status":           kivik.StatusNotFound,
		"GetIndexes/Admin/_duck.status":             kivik.StatusNotFound,
		"GetIndexes/NoAuth/_replicator.indexes":     []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/NoAuth/_users.indexes":          []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/NoAuth/_global_changes.indexes": []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/NoAuth/chicken.status":          kivik.StatusNotFound,
		"GetIndexes/NoAuth/_duck.status":            kivik.StatusNotFound,

		"DeleteIndex/RW/Admin/group/NotFoundDdoc.status":  kivik.StatusNotFound,
		"DeleteIndex/RW/Admin/group/NotFoundName.status":  kivik.StatusNotFound,
		"DeleteIndex/RW/NoAuth/group/NotFoundDdoc.status": kivik.StatusNotFound,
		"DeleteIndex/RW/NoAuth/group/NotFoundName.status": kivik.StatusNotFound,
	})
}
