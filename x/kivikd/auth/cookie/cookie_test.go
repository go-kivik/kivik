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

package cookie

import (
	"net/http/httptest"
	"testing"
)

type redirTest struct {
	Name     string
	Input    string
	Expected string
	Err      string
}

func TestRedirectURL(t *testing.T) {
	tests := []redirTest{
		{Name: "NoURL", Input: "-"},
		{Name: "EmptyValue", Input: "", Err: "redirection url must be relative to server root"},
		{Name: "Absolute", Input: "http://google.com/", Err: "redirection url must be relative to server root"},
		{Name: "InvalidURL", Input: "://google.com/", Err: "redirection url must be relative to server root"},
		{Name: "NoSlash", Input: "foobar", Err: "redirection url must be relative to server root"},
		{Name: "Relative", Input: "/_session", Expected: "/_session"},
		{Name: "InvalidRelative", Input: "/session%25%26%26", Err: "invalid redirection url"},
		{Name: "Schemaless", Input: "//evil.org", Err: "invalid redirection url"},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			url := "/"
			if test.Input != "-" {
				url += "?next=" + test.Input
			}
			r := httptest.NewRequest("GET", url, nil)
			result, err := redirectURL(r)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if test.Err != errMsg {
				t.Errorf("Unexpected error result. Expected '%s', got '%s'", test.Err, errMsg)
			}
			if test.Expected != result {
				t.Errorf("Unexpected result. Expected '%s', got '%s'", test.Expected, result)
			}
		})
	}
}
