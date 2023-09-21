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

func Test_describe_doc_RunE(t *testing.T) {
	tests := testy.NewTable()

	tests.Add("describe doc", func(t *testing.T) interface{} {
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
			args: []string{"describe", "doc", s.URL + "/foo/bar"},
		}
	})
	tests.Add("describe doc json", func(t *testing.T) interface{} {
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
			args: []string{"describe", "-f", "json", "doc", s.URL + "/foo/bar"},
		}
	})
	tests.Add("describe doc header", func(t *testing.T) interface{} {
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
			args: []string{"describe", "-H", "doc", s.URL + "/foo/bar"},
		}
	})
	tests.Add("describe doc verbose", func(t *testing.T) interface{} {
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
			args: []string{"describe", "-v", "doc", s.URL + "/foo/bar"},
		}
	})

	tests.Run(t, func(t *testing.T, tt cmdTest) {
		tt.Test(t)
	})
}
