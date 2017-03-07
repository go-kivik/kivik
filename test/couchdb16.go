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

		"AllDocs/Admin.databases":            []string{"_replicator", "chicken"},
		"AllDocs/Admin/_replicator.expected": []string{"_design/_replicator"},
		"AllDocs/Admin/_replicator.offset":   0,
		"AllDocs/Admin/chicken.status":       kivik.StatusNotFound,
		"AllDocs/NoAuth.databases":           []string{"_replicator", "chicken"},
		"AllDocs/NoAuth/_replicator.status":  kivik.StatusForbidden,
		"AllDocs/NoAuth/chicken.status":      kivik.StatusNotFound,

		"DBExists.databases":              []string{"_users", "chicken"},
		"DBExists/Admin/_users.exists":    true,
		"DBExists/Admin/chicken.exists":   false,
		"DBExists/NoAuth/_users.exists":   true,
		"DBExists/NoAuth/chicken.exists":  false,
		"DBExists/RW/group/Admin.exists":  true,
		"DBExists/RW/group/NoAuth.exists": true,

		"Membership.status": kivik.StatusBadRequest,

		"UUIDs.counts":                []int{-1, 0, 1, 10},
		"UUIDs/Admin/-1Count.status":  kivik.StatusBadRequest,
		"UUIDs/NoAuth/-1Count.status": kivik.StatusBadRequest,

		"Log/NoAuth.status":                   kivik.StatusUnauthorized,
		"Log/Admin/Offset-1000.skip":          true, // This appears to trigger a bug in CouchDB, that sometimes returns 500, and sometimes returns a log
		"Log/Admin/HTTP/NegativeBytes.status": kivik.StatusInternalServerError,
		"Log/Admin/HTTP/TextBytes.status":     kivik.StatusInternalServerError,

		"ServerInfo.version":        `^1\.6\.1$`,
		"ServerInfo.vendor":         `^The Apache Software Foundation$`,
		"ServerInfo.vendor_version": `^1\.6\.1$`,

		"Get/RW/group/Admin/bogus.status":  kivik.StatusNotFound,
		"Get/RW/group/NoAuth/bogus.status": kivik.StatusNotFound,
	})
}
