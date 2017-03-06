package test

import "github.com/flimzy/kivik/test/kt"

func init() {
	RegisterSuite(SuiteCouch16, kt.SuiteConfig{
		"AllDBs.expected": []string{"_replicator", "_users"},
		"Config/Admin/GetAll.expected_sections": []string{"admins", "attachments", "compaction_daemon", "cors", "couch_httpd_auth",
			"couch_httpd_oauth", "couchdb", "daemons", "database_compaction", "httpd", "httpd_db_handlers", "httpd_design_handlers",
			"httpd_global_handlers", "log", "query_server_config", "query_servers", "replicator", "ssl", "stats", "uuids", "vendor",
			"view_compaction"},
	})
}
