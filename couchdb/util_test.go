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
	"fmt"
	"net/http"
	"strings"
	"testing"

	"gitlab.com/flimzy/testy"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/internal"
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
			if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

type multiOptions []kivik.Option

var _ kivik.Option = (multiOptions)(nil)

func (o multiOptions) Apply(t interface{}) {
	for _, opt := range o {
		if opt != nil {
			opt.Apply(t)
		}
	}
}

func (o multiOptions) String() string {
	parts := make([]string, 0, len(o))
	for _, opt := range o {
		if part := fmt.Sprintf("%s", opt); part != "" {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, ",")
}
