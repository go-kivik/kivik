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

package kivik

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	pkgerrs "github.com/pkg/errors"
	"gitlab.com/flimzy/testy"
	"golang.org/x/xerrors"
)

func TestStatusCoder(t *testing.T) {
	type scTest struct {
		Name     string
		Err      error
		Expected int
	}
	tests := []scTest{
		{
			Name:     "nil",
			Expected: 0,
		},
		{
			Name:     "Standard error",
			Err:      errors.New("foo"),
			Expected: 500,
		},
		{
			Name:     "StatusCoder",
			Err:      &Error{HTTPStatus: 400, Err: errors.New("bad request")},
			Expected: 400,
		},
		{
			Name:     "buried xerrors StatusCoder",
			Err:      xerrors.Errorf("foo: %w", &Error{HTTPStatus: 400, Err: errors.New("bad request")}),
			Expected: 400,
		},
		{
			Name:     "buried pkg/errors StatusCoder",
			Err:      pkgerrs.Wrap(&Error{HTTPStatus: 400, Err: errors.New("bad request")}, "foo"),
			Expected: 400,
		},
		{
			Name: "deeply buried",
			Err: func() error {
				err := error(&Error{HTTPStatus: 400, Err: errors.New("bad request")})
				err = pkgerrs.Wrap(err, "foo")
				err = xerrors.Errorf("bar: %w", err)
				err = pkgerrs.Wrap(err, "foo")
				err = xerrors.Errorf("bar: %w", err)
				err = pkgerrs.Wrap(err, "foo")
				err = xerrors.Errorf("bar: %w", err)
				return err
			}(),
			Expected: 400,
		},
	}
	for _, test := range tests {
		func(test scTest) {
			t.Run(test.Name, func(t *testing.T) {
				result := StatusCode(test.Err)
				if result != test.Expected {
					t.Errorf("Unexpected result. Expected %d, got %d", test.Expected, result)
				}
			})
		}(test)
	}
}

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
		err:  &Error{HTTPStatus: http.StatusNotFound, Err: errors.New("not found")},
		str:  "not found",
		std:  "not found",
		full: `404 / Not Found: not found`,
	})
	tests.Add("with message", tst{
		err:  &Error{HTTPStatus: http.StatusNotFound, Message: "It's missing", Err: errors.New("not found")},
		str:  "It's missing: not found",
		std:  "It's missing: not found",
		full: `It's missing: 404 / Not Found: not found`,
	})
	tests.Add("embedded error", func() interface{} {
		_, err := json.Marshal(func() {}) //nolint:staticcheck
		return tst{
			err:  &Error{HTTPStatus: http.StatusBadRequest, Err: err},
			str:  "json: unsupported type: func()",
			std:  "json: unsupported type: func()",
			full: `400 / Bad Request: json: unsupported type: func()`,
		}
	})
	tests.Add("embedded network error", func() interface{} {
		client := testy.HTTPClient(func(_ *http.Request) (*http.Response, error) {
			_, err := json.Marshal(func() {}) //nolint:staticcheck
			return nil, &Error{HTTPStatus: http.StatusBadRequest, Err: err}
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
