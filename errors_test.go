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
	"errors"
	"fmt"
	"testing"

	internal "github.com/go-kivik/kivik/v4/int/errors"
)

func TestHTTPStatus(t *testing.T) {
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
			Name:     "HTTPStatus",
			Err:      &internal.Error{Status: 400, Err: errors.New("bad request")},
			Expected: 400,
		},
		{
			Name:     "wrapped HTTPStatus",
			Err:      fmt.Errorf("foo: %w", &internal.Error{Status: 400, Err: errors.New("bad request")}),
			Expected: 400,
		},
		{
			Name: "deeply buried",
			Err: func() error {
				err := error(&internal.Error{Status: 400, Err: errors.New("bad request")})
				err = fmt.Errorf("foo:%w", err)
				err = fmt.Errorf("bar: %w", err)
				err = fmt.Errorf("foo:%w", err)
				err = fmt.Errorf("bar: %w", err)
				err = fmt.Errorf("foo:%w", err)
				err = fmt.Errorf("bar: %w", err)
				return err
			}(),
			Expected: 400,
		},
	}
	for _, test := range tests {
		func(test scTest) {
			t.Run(test.Name, func(t *testing.T) {
				result := HTTPStatus(test.Err)
				if result != test.Expected {
					t.Errorf("Unexpected result. Expected %d, got %d", test.Expected, result)
				}
			})
		}(test)
	}
}
