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
		"Config/Admin/GetSection.sections":                 []string{"log", "chicken"},
		"Config/Admin/GetSection/log.keys":                 []string{"file", "include_sasl", "level"},
		"Config/Admin/GetSection/chicken.keys":             []string{},
		"Config/Admin/GetItem.items":                       []string{"log.level", "log.foobar", "logx.level"},
		"Config/Admin/GetItem/log.foobar.status":           http.StatusNotFound,
		"Config/Admin/GetItem/logx.level.status":           http.StatusNotFound,
		"Config/Admin/GetItem/log.level.expected":          "info",
		"Config/NoAuth/GetAll.status":                      http.StatusUnauthorized,
		"Config/NoAuth/GetSection.sections":                []string{"log", "chicken"},
		"Config/NoAuth/GetSection.status":                  http.StatusUnauthorized,
		"Config/NoAuth/GetItem.items":                      []string{"log.level", "log.foobar", "logx.level"},
		"Config/NoAuth/GetItem.status":                     http.StatusUnauthorized,
		"Config/RW/NoAuth/Set.status":                      http.StatusUnauthorized,
		"Config/RW/NoAuth/Delete.status":                   http.StatusUnauthorized,
		"Config/RW/Admin/Delete/NonExistantKey.status":     http.StatusNotFound,
		"Config/RW/Admin/Delete/NonExistantSection.status": http.StatusNotFound,
	})
}
