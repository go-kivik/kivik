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

func Test_put_attachment_RunE(t *testing.T) {
	tests := testy.NewTable()

	tests.Add("missing resource", cmdTest{
		args:   []string{"put", "att"},
		status: errors.ErrUsage,
	})
	tests.Add("full url on command line", cmdTest{
		args:   []string{"--debug", "put", "att", "http://localhost:1/foo/bar/foo.txt", "-d", "Testing"},
		status: errors.ErrUnavailable,
	})
	tests.Add("json data string", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Body: io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
		}, gunzip(checkRequest))

		return cmdTest{
			args: []string{"--debug", "put", "att", s.URL + "/foo/bar/foo.txt", "--data", `Testing`},
		}
	})
	tests.Add("stdin", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Body: io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
		}, gunzip(checkRequest))

		return cmdTest{
			args:  []string{"--debug", "put", "att", s.URL + "/foo/bar/foo.txt", "--data-file", "-"},
			stdin: `Testing`,
		}
	})
	tests.Add("data file", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Body: io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
		}, gunzip(checkRequest))

		return cmdTest{
			args:  []string{"--debug", "put", "att", s.URL + "/foo/bar/doc.json", "--data-file", "./testdata/doc.json"},
			stdin: `{"foo":"bar"}`,
		}
	})

	tests.Run(t, func(t *testing.T, tt cmdTest) {
		tt.Test(t)
	})
}
