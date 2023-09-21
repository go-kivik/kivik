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

func Test_get_attachment_RunE(t *testing.T) {
	tests := testy.NewTable()

	tests.Add("missing resource", cmdTest{
		args:   []string{"get", "attachment"},
		status: errors.ErrUsage,
	})
	tests.Add("not found", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusNotFound,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(strings.NewReader(`{"error":"not_found","reason":"Document is missing attachment"}
			`)),
		})

		return cmdTest{
			args:   []string{"get", "attachment", s.URL + "/db/doc/foo.txt"},
			status: errors.ErrNotFound,
		}
	})
	tests.Add("success", func(t *testing.T) interface{} {
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
			args: []string{"get", "attachment", s.URL + "/db/doc/foo.txt"},
		}
	})
	tests.Add("success json", func(t *testing.T) interface{} {
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
			args: []string{"get", "attachment", s.URL + "/db/doc/foo.txt", "-f", "json"},
		}
	})

	tests.Run(t, func(t *testing.T, tt cmdTest) {
		tt.Test(t)
	})
}
