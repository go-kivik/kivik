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

func Test_delete_config_RunE(t *testing.T) {
	tests := testy.NewTable()

	tests.Add("missing key", func(t *testing.T) interface{} {
		return cmdTest{
			args:   []string{"delete", "config", "http://example.com/foo/"},
			status: errors.ErrUsage,
		}
	})
	tests.Add("named node", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(strings.NewReader(`"foo"`)),
		}, func(t *testing.T, req *http.Request) {
			if req.Method != http.MethodDelete {
				t.Errorf("Unexpected method: %s", req.Method)
			}
			if req.URL.Path != "/_node/quack/_config/foo/bar" {
				t.Errorf("unexpected path: %s", req.URL.Path)
			}
		})

		return cmdTest{
			args: []string{"delete", "config", s.URL, "--node", "quack", "--key", "foo/bar"},
		}
	})

	tests.Run(t, func(t *testing.T, tt cmdTest) {
		tt.Test(t)
	})
}
