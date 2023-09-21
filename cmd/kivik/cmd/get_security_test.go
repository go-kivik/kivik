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

func Test_get_sec_RunE(t *testing.T) {
	tests := testy.NewTable()

	tests.Add("missing database", cmdTest{
		args:   []string{"get", "sec"},
		status: errors.ErrUsage,
	})
	tests.Add("success", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"Server":       []string{"CouchDB/2.3.1 (Erlang OTP/20)"},
			},
			Body: io.NopCloser(strings.NewReader(`{
				"admins": {
					"names": [
						"superuser"
					],
					"roles": [
						"admins"
					]
				},
				"members": {
					"names": [
						"user1",
						"user2"
					],
					"roles": [
						"developers"
					]
				}
			}`)),
		})

		return cmdTest{
			args: []string{"get", "sec", s.URL + "/db"},
		}
	})

	tests.Run(t, func(t *testing.T, tt cmdTest) {
		tt.Test(t)
	})
}
