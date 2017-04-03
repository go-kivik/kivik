package test

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	RegisterSuite(SuiteCouch16, kt.SuiteConfig{
		"AllDBs.expected": []string{"_replicator", "_users"},

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
		"AllDocs/Admin/_duck.status":         kivik.StatusBadRequest,
		"AllDocs/NoAuth/_replicator.status":  kivik.StatusForbidden,
		"AllDocs/NoAuth/chicken.status":      kivik.StatusNotFound,
		"AllDocs/NoAuth/_duck.status":        kivik.StatusBadRequest,

		"DBExists.databases":              []string{"_users", "chicken", "_duck"},
		"DBExists/Admin/_users.exists":    true,
		"DBExists/Admin/chicken.exists":   false,
		"DBExists/Admin/_duck.status":     kivik.StatusBadRequest,
		"DBExists/NoAuth/_users.exists":   true,
		"DBExists/NoAuth/chicken.exists":  false,
		"DBExists/NoAuth/_duck.status":    kivik.StatusBadRequest,
		"DBExists/RW/group/Admin.exists":  true,
		"DBExists/RW/group/NoAuth.exists": true,

		"Membership.status": kivik.StatusNotImplemented,

		"UUIDs.counts":                []int{-1, 0, 1, 10},
		"UUIDs/Admin/-1Count.status":  kivik.StatusBadRequest,
		"UUIDs/NoAuth/-1Count.status": kivik.StatusBadRequest,

		"Log/NoAuth.status":                   kivik.StatusUnauthorized,
		"Log/NoAuth/Offset-1000.status":       kivik.StatusBadRequest,
		"Log/Admin/Offset-1000.status":        kivik.StatusBadRequest,
		"Log/Admin/HTTP/NegativeBytes.status": kivik.StatusInternalServerError,
		"Log/Admin/HTTP/TextBytes.status":     kivik.StatusInternalServerError,

		"ServerInfo.version":        `^1\.6\.1$`,
		"ServerInfo.vendor":         `^The Apache Software Foundation$`,
		"ServerInfo.vendor_version": `^1\.6\.1$`,

		"Get/RW/group/Admin/bogus.status":  kivik.StatusNotFound,
		"Get/RW/group/NoAuth/bogus.status": kivik.StatusNotFound,

		"Rev/RW/group/Admin/bogus.status":  kivik.StatusNotFound,
		"Rev/RW/group/NoAuth/bogus.status": kivik.StatusNotFound,

		"Flush.databases":                     []string{"_users", "chicken", "_duck"},
		"Flush/Admin/chicken/DoFlush.status":  kivik.StatusNotFound,
		"Flush/Admin/_duck/DoFlush.status":    kivik.StatusBadRequest,
		"Flush/NoAuth/chicken/DoFlush.status": kivik.StatusNotFound,
		"Flush/NoAuth/_duck/DoFlush.status":   kivik.StatusBadRequest,

		"Delete/RW/Admin/group/MissingDoc.status":        kivik.StatusNotFound,
		"Delete/RW/Admin/group/InvalidRevFormat.status":  kivik.StatusBadRequest,
		"Delete/RW/Admin/group/WrongRev.status":          kivik.StatusConflict,
		"Delete/RW/NoAuth/group/MissingDoc.status":       kivik.StatusNotFound,
		"Delete/RW/NoAuth/group/InvalidRevFormat.status": kivik.StatusBadRequest,
		"Delete/RW/NoAuth/group/WrongRev.status":         kivik.StatusConflict,

		"Session/Get/Admin.info.authentication_handlers":  "oauth,cookie,default",
		"Session/Get/Admin.info.authentication_db":        "_users",
		"Session/Get/Admin.info.authenticated":            "cookie",
		"Session/Get/Admin.userCtx.roles":                 "_admin",
		"Session/Get/Admin.ok":                            "true",
		"Session/Get/NoAuth.info.authentication_handlers": "oauth,cookie,default",
		"Session/Get/NoAuth.info.authentication_db":       "_users",
		"Session/Get/NoAuth.info.authenticated":           "",
		"Session/Get/NoAuth.userCtx.roles":                "",
		"Session/Get/NoAuth.ok":                           "true",

		"Session/Post/EmptyJSON.status":                             kivik.StatusBadRequest,
		"Session/Post/BogusTypeJSON.status":                         kivik.StatusUnauthorized,
		"Session/Post/BogusTypeForm.status":                         kivik.StatusUnauthorized,
		"Session/Post/EmptyForm.status":                             kivik.StatusUnauthorized,
		"Session/Post/BadJSON.status":                               kivik.StatusBadRequest,
		"Session/Post/BadForm.status":                               kivik.StatusUnauthorized,
		"Session/Post/MeaninglessJSON.status":                       kivik.StatusInternalServerError,
		"Session/Post/MeaninglessForm.status":                       kivik.StatusUnauthorized,
		"Session/Post/GoodJSON.status":                              kivik.StatusUnauthorized,
		"Session/Post/BadQueryParam.status":                         kivik.StatusUnauthorized,
		"Session/Post/BadCredsJSON.status":                          kivik.StatusUnauthorized,
		"Session/Post/BadCredsForm.status":                          kivik.StatusUnauthorized,
		"Session/Post/GoodCredsJSONRemoteRedirHeaderInjection.skip": true, // CouchDB allows header injection
		"Session/Post/GoodCredsJSONRemoteRedirInvalidURL.skip":      true, // CouchDB doesn't sanitize the Location value, so sends unparseable headers.

		"DBInfo.databases":             []string{"_users", "chicken", "_duck"},
		"DBInfo/Admin/chicken.status":  kivik.StatusNotFound,
		"DBInfo/Admin/_duck.status":    kivik.StatusBadRequest,
		"DBInfo/NoAuth/chicken.status": kivik.StatusNotFound,
		"DBInfo/NoAuth/_duck.status":   kivik.StatusBadRequest,

		"Compact/RW/NoAuth.status": kivik.StatusUnauthorized,

		"ViewCleanup/RW/NoAuth.status": kivik.StatusUnauthorized,

		"Security.databases":              []string{"_replicator", "_users", "chicken", "_duck"},
		"Security/Admin/chicken.status":   kivik.StatusNotFound,
		"Security/Admin/_duck.status":     kivik.StatusBadRequest,
		"Security/NoAuth/chicken.status":  kivik.StatusNotFound,
		"Security/NoAuth/_duck.status":    kivik.StatusBadRequest,
		"Security/RW/group/NoAuth.status": kivik.StatusUnauthorized,

		"SetSecurity/RW/Admin/NotExists.status":  kivik.StatusNotFound,
		"SetSecurity/RW/NoAuth/NotExists.status": kivik.StatusNotFound,
		"SetSecurity/RW/NoAuth/Exists.status":    kivik.StatusUnauthorized,

		"RevsLimit.databases":                     []string{"_replicator", "_users", "chicken", "_duck"},
		"RevsLimit.revs_limit":                    1000,
		"RevsLimit/Admin/chicken.status":          kivik.StatusNotFound,
		"RevsLimit/Admin/_duck.status":            kivik.StatusBadRequest,
		"RevsLimit/NoAuth/_global_changes.status": kivik.StatusUnauthorized,
		"RevsLimit/NoAuth/chicken.status":         kivik.StatusNotFound,
		"RevsLimit/NoAuth/_duck.status":           kivik.StatusBadRequest,
		"RevsLimit/RW/NoAuth/Set.status":          kivik.StatusUnauthorized,

		"DBUpdates/RW/NoAuth.status": kivik.StatusUnauthorized,

		"BulkDocs/RW/NoAuth/group/Mix/Conflict.status": kivik.StatusConflict,
		"BulkDocs/RW/Admin/group/Mix/Conflict.status":  kivik.StatusConflict,
	})
}
