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

func Test_delete_RunE(t *testing.T) {
	tests := testy.NewTable()

	tests.Add("missing resource", cmdTest{
		args:   []string{"delete"},
		status: errors.ErrUsage,
	})
	tests.Add("auto delete doc", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"Server":       []string{"CouchDB/2.3.1 (Erlang OTP/20)"},
				"ETag":         []string{`"2-eec205a9d413992850a6e32678485900`},
			},
			Body: io.NopCloser(strings.NewReader(`{"ok":true,"id":"fe6a1fef482d660160b45165ed001740","rev":"2-eec205a9d413992850a6e32678485900"}`)),
		})

		return cmdTest{
			args: []string{"delete", s.URL + "/db/doc", "-O", "rev=1-xxx"},
		}
	})
	tests.Add("auto delete db", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"Server":       []string{"CouchDB/2.3.1 (Erlang OTP/20)"},
			},
			Body: io.NopCloser(strings.NewReader(`{"ok":true}`)),
		})

		return cmdTest{
			args: []string{"delete", s.URL + "/db"},
		}
	})
	tests.Add("auto delete attachment", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"Server":       []string{"CouchDB/2.3.1 (Erlang OTP/20)"},
				"ETag":         []string{`"2-eec205a9d413992850a6e32678485900`},
			},
			Body: io.NopCloser(strings.NewReader(`{"ok":true,"id":"fe6a1fef482d660160b45165ed001740","rev":"2-eec205a9d413992850a6e32678485900"}`)),
		})

		return cmdTest{
			args: []string{"delete", s.URL + "/db/doc/filename.txt", "-O", "rev=1-xxx"},
		}
	})
	tests.Add("auto delete config", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"Server":       []string{"CouchDB/2.3.1 (Erlang OTP/20)"},
			},
			Body: io.NopCloser(strings.NewReader(`"foo"`)),
		})

		return cmdTest{
			args: []string{"delete", s.URL + "/_node/_local/_config/foo/bar"},
		}
	})

	tests.Run(t, func(t *testing.T, tt cmdTest) {
		tt.Test(t)
	})
}
