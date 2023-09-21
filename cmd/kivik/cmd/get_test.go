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
	"strings"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/xkivik/v4/cmd/kivik/errors"
)

func Test_get_RunE(t *testing.T) {
	tests := testy.NewTable()

	tests.Add("missing document", cmdTest{
		args:   []string{"get"},
		status: errors.ErrUsage,
	})
	tests.Add("invalid URL on command line", cmdTest{
		args:   []string{"--debug", "get", "http://localhost:1/foo/bar/%xxx"},
		status: errors.ErrUsage,
	})
	tests.Add("invalid URL on command line, doc command", cmdTest{
		args:   []string{"--debug", "get", "document", "http://localhost:1/foo/bar/%xxx"},
		status: errors.ErrUsage,
	})
	tests.Add("full url on command line", cmdTest{
		args:   []string{"--debug", "get", "http://localhost:1/foo/bar"},
		status: errors.ErrUnavailable,
	})
	tests.Add("path only on command line", cmdTest{
		args:   []string{"--debug", "--config", "./testdata/localhost.yaml", "get", "/foo/bar"},
		status: errors.ErrUnavailable,
	})
	tests.Add("document only on command line", cmdTest{
		args:   []string{"--debug", "--config", "./testdata/localhost.yaml", "get", "bar"},
		status: errors.ErrUnavailable,
	})
	tests.Add("not found", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusNotFound,
		})

		return cmdTest{
			args:   []string{"get", s.URL},
			status: errors.ErrNotFound,
		}
	})
	tests.Add("invalid JSON response", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"ETag":         []string{"1-xxx"},
			},
			Body: io.NopCloser(strings.NewReader("invalid")),
		})

		return cmdTest{
			args:   []string{"get", s.URL},
			status: errors.ErrProtocol,
		}
	})
	tests.Add("success", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"ETag":         []string{"1-xxx"},
			},
			Body: io.NopCloser(strings.NewReader(`{
				"_id":"foo",
				"_rev":"1-xxx",
				"foo":"bar"
			}`)),
		})

		return cmdTest{
			args: []string{"get", s.URL + "/db/doc"},
		}
	})
	tests.Add("get database", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"ETag":         []string{"1-xxx"},
			},
			Body: io.NopCloser(strings.NewReader(`{"db_name":"foo","purge_seq":"0-g1AAAABPeJzLYWBgYMpgTmHgzcvPy09JdcjLz8gvLskBCeexAEmGBiD1HwiyEhlwqEtkSKqHKMgCAIT2GV4","update_seq":"0-g1AAAABPeJzLYWBgYMpgTmHgzcvPy09JdcjLz8gvLskBCeexAEmGBiD1HwiyEhlwqEtkSKqHKMgCAIT2GV4","sizes":{"file":16692,"external":0,"active":0},"props":{},"doc_del_count":0,"doc_count":0,"disk_format_version":8,"compact_running":false,"cluster":{"q":2,"n":1,"w":1,"r":1},"instance_start_time":"0"}
			`)),
		})
		return cmdTest{
			args: []string{"get", "database", s.URL + "/foo"},
		}
	})
	tests.Add("auto get database", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"ETag":         []string{"1-xxx"},
			},
			Body: io.NopCloser(strings.NewReader(`{"db_name":"foo","purge_seq":"0-g1AAAABPeJzLYWBgYMpgTmHgzcvPy09JdcjLz8gvLskBCeexAEmGBiD1HwiyEhlwqEtkSKqHKMgCAIT2GV4","update_seq":"0-g1AAAABPeJzLYWBgYMpgTmHgzcvPy09JdcjLz8gvLskBCeexAEmGBiD1HwiyEhlwqEtkSKqHKMgCAIT2GV4","sizes":{"file":16692,"external":0,"active":0},"props":{},"doc_del_count":0,"doc_count":0,"disk_format_version":8,"compact_running":false,"cluster":{"q":2,"n":1,"w":1,"r":1},"instance_start_time":"0"}
			`)),
		})
		return cmdTest{
			args: []string{"--debug", "get", s.URL + "/foo"},
		}
	})
	tests.Add("describe database", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"ETag":         []string{"1-xxx"},
			},
			Body: io.NopCloser(strings.NewReader(`{"db_name":"foo","purge_seq":"0-g1AAAABPeJzLYWBgYMpgTmHgzcvPy09JdcjLz8gvLskBCeexAEmGBiD1HwiyEhlwqEtkSKqHKMgCAIT2GV4","update_seq":"0-g1AAAABPeJzLYWBgYMpgTmHgzcvPy09JdcjLz8gvLskBCeexAEmGBiD1HwiyEhlwqEtkSKqHKMgCAIT2GV4","sizes":{"file":16692,"external":0,"active":0},"props":{},"doc_del_count":0,"doc_count":0,"disk_format_version":8,"compact_running":false,"cluster":{"q":2,"n":1,"w":1,"r":1},"instance_start_time":"0"}
			`)),
		})
		return cmdTest{
			args: []string{"describe", "database", s.URL + "/foo"},
		}
	})
	tests.Add("auto version", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"ETag":         []string{"1-xxx"},
			},
			Body: io.NopCloser(strings.NewReader(`{"couchdb":"Welcome","version":"2.3.1","git_sha":"c298091a4","uuid":"0ae5d1a72d60e4e1370a444f1cf7ce7c","features":["pluggable-storage-engines","scheduler"],"vendor":{"name":"The Apache Software Foundation"}}
			`)),
		})

		return cmdTest{
			args: []string{"get", s.URL},
		}
	})
	tests.Add("auto attachment", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"text/plain"},
				"Server":       []string{"CouchDB/2.3.1 (Erlang OTP/20)"},
				"ETag":         []string{`"cy5z3SF7yaYp4vmLX0k31Q==`},
			},
			Body: io.NopCloser(strings.NewReader(`Testing`)),
		})

		return cmdTest{
			args: []string{"get", s.URL + "/db/doc/foo.txt"},
		}
	})
	tests.Add("auto all-dbs", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"text/plain"},
				"Server":       []string{"CouchDB/2.3.1 (Erlang OTP/20)"},
				"ETag":         []string{`"cy5z3SF7yaYp4vmLX0k31Q==`},
			},
			Body: io.NopCloser(strings.NewReader(`["foo","bar"]`)),
		})

		return cmdTest{
			args: []string{"get", s.URL + "/_all_dbs"},
		}
	})
	tests.Add("auto config key", func(t *testing.T) interface{} {
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
			args: []string{"get", s.URL + "/_node/_local/_config/chttpd/backlog"},
		}
	})
	tests.Add("auto security", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"ETag":         []string{"1-xxx"},
			},
			Body: io.NopCloser(strings.NewReader(`{"admins":{"names":["bob"]}}`)),
		}, func(t *testing.T, req *http.Request) {
			if req.URL.Path != "/foo/_security" {
				t.Errorf("unexpected path: %s", req.URL.Path)
			}
		})

		return cmdTest{
			args: []string{"get", s.URL + "/foo/_security"},
		}
	})
	tests.Add("auto cluster-setup", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"ETag":         []string{"1-xxx"},
			},
			Body: io.NopCloser(strings.NewReader(`{"state":"cluster_enabled"}
}`)),
		}, func(t *testing.T, req *http.Request) {
			if req.URL.Path != clusterPath {
				t.Errorf("unexpected path: %s", req.URL.Path)
			}
		})

		return cmdTest{
			args: []string{"get", s.URL + clusterPath},
		}
	})

	tests.Run(t, func(t *testing.T, tt cmdTest) {
		tt.Test(t)
	})
}
