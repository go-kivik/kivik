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

package chttp

import (
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/couchdb/v4/internal"
)

func TestFullCommit(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected bool
		status   int
		err      string
	}{
		{
			name:     "new",
			input:    map[string]interface{}{internal.OptionFullCommit: true},
			expected: true,
		},
		{
			name:   "new error",
			input:  map[string]interface{}{internal.OptionFullCommit: 123},
			status: http.StatusBadRequest,
			err:    "kivik: option 'X-Couch-Full-Commit' must be bool, not int",
		},
		{
			name:     "none",
			input:    nil,
			expected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := fullCommit(test.input)
			testy.StatusError(t, test.err, test.status, err)
			if result != test.expected {
				t.Errorf("Unexpected result: %v", result)
			}
			if _, ok := test.input[internal.OptionFullCommit]; ok {
				t.Errorf("%s still set in options", internal.OptionFullCommit)
			}
		})
	}
}

func TestIfNoneMatch(t *testing.T) {
	tests := []struct {
		name     string
		opts     map[string]interface{}
		expected string
		status   int
		err      string
	}{
		{
			name:     "nil",
			opts:     nil,
			expected: "",
		},
		{
			name:     "inm not set",
			opts:     map[string]interface{}{"foo": "bar"},
			expected: "",
		},
		{
			name:   "wrong type",
			opts:   map[string]interface{}{internal.OptionIfNoneMatch: 123},
			status: http.StatusBadRequest,
			err:    "kivik: option 'If-None-Match' must be string, not int",
		},
		{
			name:     "valid",
			opts:     map[string]interface{}{internal.OptionIfNoneMatch: "foo"},
			expected: `"foo"`,
		},
		{
			name:     "valid, pre-quoted",
			opts:     map[string]interface{}{internal.OptionIfNoneMatch: `"foo"`},
			expected: `"foo"`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ifNoneMatch(test.opts)
			testy.StatusError(t, test.err, test.status, err)
			if result != test.expected {
				t.Errorf("Unexpected result: %s", result)
			}
			if _, ok := test.opts[internal.OptionIfNoneMatch]; ok {
				t.Errorf("%s still set in options", internal.OptionIfNoneMatch)
			}
		})
	}
}
