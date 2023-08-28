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

package couchdb

import (
	"encoding/json"
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"
)

func TestDeJSONify(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
		status   int
		err      string
	}{
		{
			name:     "string",
			input:    `{"foo":"bar"}`,
			expected: map[string]interface{}{"foo": "bar"},
		},
		{
			name:     "[]byte",
			input:    []byte(`{"foo":"bar"}`),
			expected: map[string]interface{}{"foo": "bar"},
		},
		{
			name:     "json.RawMessage",
			input:    json.RawMessage(`{"foo":"bar"}`),
			expected: map[string]interface{}{"foo": "bar"},
		},
		{
			name:     "map",
			input:    map[string]string{"foo": "bar"},
			expected: map[string]string{"foo": "bar"},
		},
		{
			name:   "invalid JSON sring",
			input:  `{"foo":"\C"}`,
			status: http.StatusBadRequest,
			err:    "invalid character 'C' in string escape code",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := deJSONify(test.input)
			testy.StatusError(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}
