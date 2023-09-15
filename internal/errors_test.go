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

package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"gitlab.com/flimzy/testy"
)

func TestFormatError(t *testing.T) {
	type tst struct {
		err  error
		str  string
		std  string
		full string
	}
	tests := testy.NewTable()
	tests.Add("standard error", tst{
		err:  errors.New("foo"),
		str:  "foo",
		std:  "foo",
		full: "foo",
	})
	tests.Add("not from server", tst{
		err:  &Error{Status: http.StatusNotFound, Err: errors.New("not found")},
		str:  "not found",
		std:  "not found",
		full: `404 / Not Found: not found`,
	})
	tests.Add("with message", tst{
		err:  &Error{Status: http.StatusNotFound, Message: "It's missing", Err: errors.New("not found")},
		str:  "It's missing: not found",
		std:  "It's missing: not found",
		full: `It's missing: 404 / Not Found: not found`,
	})
	tests.Add("embedded error", func() interface{} {
		_, err := json.Marshal(func() {}) //nolint:staticcheck
		return tst{
			err:  &Error{Status: http.StatusBadRequest, Err: err},
			str:  "json: unsupported type: func()",
			std:  "json: unsupported type: func()",
			full: `400 / Bad Request: json: unsupported type: func()`,
		}
	})
	tests.Add("embedded network error", func() interface{} {
		client := testy.HTTPClient(func(_ *http.Request) (*http.Response, error) {
			_, err := json.Marshal(func() {}) //nolint:staticcheck
			return nil, &Error{Status: http.StatusBadRequest, Err: err}
		})
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		_, err := client.Do(req)
		return tst{
			err:  err,
			str:  "Get /: json: unsupported type: func()",
			std:  "Get /: json: unsupported type: func()",
			full: "Get /: json: unsupported type: func()",
		}
	})

	re := testy.Replacement{
		Regexp:      regexp.MustCompile(`"/"`),
		Replacement: "/",
	}

	tests.Run(t, func(t *testing.T, test tst) {
		if d := testy.DiffText(test.str, test.err.Error(), re); d != nil {
			t.Errorf("Error():\n%s", d)
		}
		if d := testy.DiffText(test.std, fmt.Sprintf("%v", test.err), re); d != nil {
			t.Errorf("Standard:\n%s", d)
		}
		if d := testy.DiffText(test.full, fmt.Sprintf("%+v", test.err), re); d != nil {
			t.Errorf("Full:\n%s", d)
		}
	})
}
