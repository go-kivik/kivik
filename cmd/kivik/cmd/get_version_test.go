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

func Test_get_version_RunE(t *testing.T) {
	tests := testy.NewTable()

	tests.Add("get version", func(t *testing.T) interface{} {
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
			args: []string{"get", "version", s.URL},
		}
	})
	tests.Add("get version json", func(t *testing.T) interface{} {
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
			args: []string{"get", "version", s.URL, "-f", "json"},
		}
	})
	tests.Add("get version no url", func(t *testing.T) interface{} {
		return cmdTest{
			args:   []string{"get", "version"},
			status: errors.ErrUsage,
		}
	})

	tests.Run(t, func(t *testing.T, tt cmdTest) {
		tt.Test(t)
	})
}
