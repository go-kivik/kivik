package test

import (
	"net/http"

	"github.com/flimzy/kivik/test/kt"
)

func init() {
	RegisterSuite(SuiteCouch16, kt.SuiteConfig{
		"AllDBs.expected": []string{"_replicator", "_users"},

		"Config/Admin/GetAll.expected_sections": []string{"admins", "attachments", "compaction_daemon", "cors", "couch_httpd_auth",
			"couch_httpd_oauth", "couchdb", "daemons", "database_compaction", "httpd", "httpd_db_handlers", "httpd_design_handlers",
			"httpd_global_handlers", "log", "query_server_config", "query_servers", "replicator", "ssl", "stats", "uuids", "vendor",
			"view_compaction"},
		"Config/Admin/GetSection.sections":                       []string{"log", "chicken"},
		"Config/Admin/GetSection/log.keys":                       []string{"file", "include_sasl", "level"},
		"Config/Admin/GetSection/chicken.keys":                   []string{},
		"Config/Admin/GetItem.items":                             []string{"log.level", "log.foobar", "logx.level"},
		"Config/Admin/GetItem/log.foobar.status":                 http.StatusNotFound,
		"Config/Admin/GetItem/logx.level.status":                 http.StatusNotFound,
		"Config/Admin/GetItem/log.level.expected":                "info",
		"Config/NoAuth/GetAll.status":                            http.StatusUnauthorized,
		"Config/NoAuth/GetSection.sections":                      []string{"log", "chicken"},
		"Config/NoAuth/GetSection.status":                        http.StatusUnauthorized,
		"Config/NoAuth/GetItem.items":                            []string{"log.level", "log.foobar", "logx.level"},
		"Config/NoAuth/GetItem.status":                           http.StatusUnauthorized,
		"Config/RW/group/NoAuth/Set.status":                      http.StatusUnauthorized,
		"Config/RW/group/NoAuth/Delete.status":                   http.StatusUnauthorized,
		"Config/RW/group/Admin/Delete/NonExistantKey.status":     http.StatusNotFound,
		"Config/RW/group/Admin/Delete/NonExistantSection.status": http.StatusNotFound,

		"CreateDB/RW/NoAuth.status":         http.StatusUnauthorized,
		"CreateDB/RW/Admin/Recreate.status": http.StatusPreconditionFailed,

		"AllDocs/Admin.databases":            []string{"_replicator", "chicken"},
		"AllDocs/Admin/_replicator.expected": []string{"_design/_replicator"},
		"AllDocs/Admin/_replicator.offset":   0,
		"AllDocs/Admin/chicken.status":       http.StatusNotFound,
		"AllDocs/NoAuth.databases":           []string{"_replicator", "chicken"},
		"AllDocs/NoAuth/_replicator.status":  http.StatusForbidden,
		"AllDocs/NoAuth/chicken.status":      http.StatusNotFound,

		"DBExists.databases":              []string{"_users", "chicken"},
		"DBExists/Admin/_users.exists":    true,
		"DBExists/Admin/chicken.exists":   false,
		"DBExists/NoAuth/_users.exists":   true,
		"DBExists/NoAuth/chicken.exists":  false,
		"DBExists/RW/group/Admin.exists":  true,
		"DBExists/RW/group/NoAuth.exists": true,

		"Membership.status": http.StatusBadRequest,

		"UUIDs.counts":                []int{-1, 0, 1, 10},
		"UUIDs/Admin/-1Count.status":  http.StatusBadRequest,
		"UUIDs/NoAuth/-1Count.status": http.StatusBadRequest,

		"Log/NoAuth.status":                   http.StatusUnauthorized,
		"Log/Admin/Offset-1000.skip":          true, // This appears to trigger a bug in CouchDB, that sometimes returns 500, and sometimes returns a log
		"Log/Admin/HTTP/NegativeBytes.status": http.StatusInternalServerError,
		"Log/Admin/HTTP/TextBytes.status":     http.StatusInternalServerError,
	})
}
