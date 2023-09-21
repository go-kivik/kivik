// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package cmd

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"gitlab.com/flimzy/testy"
)

func Test_get_config_RunE(t *testing.T) {
	tests := testy.NewTable()

	tests.Add("get config", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"ETag":         []string{"1-xxx"},
			},
			Body: io.NopCloser(strings.NewReader(`{"uuids":{"algorithm":"sequential","max_count":"1000"},"cluster":{"n":"1","q":"8"},"cors":{"credentials":"false"},"chttpd":{"backlog":"512","bind_address":"0.0.0.0","docroot":"./share/www","max_db_number_for_dbs_info_req":"100","port":"5984","prefer_minimal":"Cache-Control, Content-Length, Content-Range, Content-Type, ETag, Server, Transfer-Encoding, Vary","require_valid_user":"false","server_options":"[{recbuf, undefined}]","socket_options":"[{sndbuf, 262144}, {nodelay, true}]"},"attachments":{"compressible_types":"text/*, application/javascript, application/json, application/xml","compression_level":"8"},"query_server_config":{"os_process_limit":"100","reduce_limit":"true"},"vendor":{"name":"The Apache Software Foundation"},"replicator":{"connection_timeout":"30000","http_connections":"20","interval":"60000","max_churn":"20","max_jobs":"500","retries_per_request":"5","socket_options":"[{keepalive, true}, {nodelay, false}]","ssl_certificate_max_depth":"3","startup_jitter":"5000","verify_ssl_certificates":"false","worker_batch_size":"500","worker_processes":"4"},"ssl":{"port":"6984"},"log":{"file":"/var/log/couchdb/couchdb.log","level":"info","writer":"file"},"indexers":{"couch_mrview":"true"},"view_compaction":{"keyvalue_buffer_size":"2097152"},"features":{"pluggable-storage-engines":"true","scheduler":"true"},"couch_peruser":{"database_prefix":"userdb-","delete_dbs":"false","enable":"false"},"httpd":{"allow_jsonp":"false","authentication_handlers":"{couch_httpd_auth, cookie_authentication_handler}, {couch_httpd_auth, default_authentication_handler}","bind_address":"127.0.0.1","enable_cors":"false","enable_xframe_options":"false","max_http_request_size":"4294967296","port":"5986","secure_rewrites":"true","socket_options":"[{sndbuf, 262144}]"},"database_compaction":{"checkpoint_after":"5242880","doc_buffer_size":"524288"},"csp":{"enable":"true"},"couch_httpd_auth":{"allow_persistent_cookies":"true","auth_cache_size":"50","authentication_db":"_users","authentication_redirect":"/_utils/session.html","iterations":"10","require_valid_user":"false","timeout":"600"},"couchdb_engines":{"couch":"couch_bt_engine"},"couchdb":{"attachment_stream_buffer_size":"4096","changes_doc_ids_optimization_threshold":"100","database_dir":"./data","default_engine":"couch","default_security":"admin_local","delayed_commits":"false","file_compression":"snappy","max_dbs_open":"500","os_process_timeout":"5000","uuid":"0ae5d1a72d60e4e1370a444f1cf7ce7c","view_index_dir":"./data"},"compactions":{"_default":"[{db_fragmentation, \"70%\"}, {view_fragmentation, \"60%\"}]"},"compaction_daemon":{"check_interval":"3600","min_file_size":"131072"}}`)),
		}, func(t *testing.T, req *http.Request) {
			if req.URL.Path != "/_node/_local/_config" {
				t.Errorf("unexpected path: %s", req.URL.Path)
			}
		})

		return cmdTest{
			args: []string{"get", "config", s.URL},
		}
	})
	tests.Add("named node", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"ETag":         []string{"1-xxx"},
			},
			Body: io.NopCloser(strings.NewReader(`{}`)),
		}, func(t *testing.T, req *http.Request) {
			if req.URL.Path != "/_node/foo/_config" {
				t.Errorf("unexpected path: %s", req.URL.Path)
			}
		})

		return cmdTest{
			args: []string{"get", "config", s.URL, "--node", "foo"},
		}
	})
	tests.Add("section", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"ETag":         []string{"1-xxx"},
			},
			Body: io.NopCloser(strings.NewReader(`{"max_db_number_for_dbs_info_req":"100","port":"5984","prefer_minimal":"Cache-Control, Content-Length, Content-Range, Content-Type, ETag, Server, Transfer-Encoding, Vary","backlog":"512","docroot":"./share/www","socket_options":"[{sndbuf, 262144}, {nodelay, true}]","require_valid_user":"false","server_options":"[{recbuf, undefined}]","bind_address":"0.0.0.0"}`)),
		}, func(t *testing.T, req *http.Request) {
			if req.URL.Path != "/_node/_local/_config/chttpd" {
				t.Errorf("unexpected path: %s", req.URL.Path)
			}
		})

		return cmdTest{
			args: []string{"get", "config", s.URL, "--key", "chttpd"},
		}
	})
	tests.Add("key", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"ETag":         []string{"1-xxx"},
			},
			Body: io.NopCloser(strings.NewReader(`"512"`)),
		}, func(t *testing.T, req *http.Request) {
			if req.URL.Path != "/_node/_local/_config/chttpd/backlog" {
				t.Errorf("unexpected path: %s", req.URL.Path)
			}
		})

		return cmdTest{
			args: []string{"get", "config", s.URL, "--key", "chttpd/backlog"},
		}
	})

	tests.Run(t, func(t *testing.T, tt cmdTest) {
		tt.Test(t)
	})
}

func Test_configFromDSN(t *testing.T) {
	type tt struct {
		dsn       string
		node, key string
		ok        bool
	}

	tests := testy.NewTable()
	tests.Add("no path", tt{
		dsn: "http://foo.com/",
	})
	tests.Add("short path", tt{
		dsn: "http://foo.com/_node/foo",
	})
	tests.Add("config", tt{
		dsn:  "http://foo.com/_node/foo/_config",
		node: "foo",
		ok:   true,
	})
	tests.Add("config key", tt{
		dsn:  "http://foo.com/_node/foo/_config/foo/bar",
		node: "foo",
		key:  "foo/bar",
		ok:   true,
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		url, err := url.Parse(tt.dsn)
		if err != nil {
			t.Fatal(err)
		}
		node, key, ok := configFromDSN(url)
		if node != tt.node || key != tt.key || ok != tt.ok {
			t.Errorf("Unexpected result: node:%s, key:%s, ok:%T", node, key, ok)
		}
	})
}
