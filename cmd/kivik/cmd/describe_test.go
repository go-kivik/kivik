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
)

func Test_describe_RunE(t *testing.T) {
	tests := testy.NewTable()

	tests.Add("auto describe doc", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type":   []string{"application/json"},
				"ETag":           []string{"1-xxx"},
				"Content-Length": []string{"59"},
			},
			Body: io.NopCloser(strings.NewReader(`{
				"_id":"foo",
				"_rev":"1-xxx",
				"foo":"bar"
			}`)),
		})

		return cmdTest{
			args: []string{"describe", s.URL + "/foo/bar"},
		}
	})
	tests.Add("auto describe version", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"Server":       []string{"CouchDB/2.3.1 (Erlang OTP/20)"},
			},
			Body: io.NopCloser(strings.NewReader(`{"couchdb":"Welcome","version":"2.3.1","git_sha":"c298091a4","uuid":"0ae5d1a72d60e4e1370a444f1cf7ce7c","features":["pluggable-storage-engines","scheduler"],"vendor":{"name":"The Apache Software Foundation"}}
			`)),
		})

		return cmdTest{
			args: []string{"describe", s.URL},
		}
	})
	tests.Add("auto attachment", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type":   []string{"text/plain"},
				"Server":         []string{"CouchDB/2.3.1 (Erlang OTP/20)"},
				"ETag":           []string{`"cy5z3SF7yaYp4vmLX0k31Q==`},
				"Content-Length": []string{"7"},
			},
			Body: io.NopCloser(strings.NewReader(`Testing`)),
		})

		return cmdTest{
			args: []string{"descr", s.URL + "/db/doc/foo.txt"},
		}
	})

	tests.Run(t, func(t *testing.T, tt cmdTest) {
		tt.Test(t)
	})
}
