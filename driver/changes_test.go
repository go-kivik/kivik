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

package driver

import (
	"encoding/json"
	"testing"

	"gitlab.com/flimzy/testy"
)

func TestChangesUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ChangedRevs
		err      string
	}{
		{
			name:  "invalid JSON",
			input: `{"foo":"bar"}`,
			err:   `json: cannot unmarshal object into Go value of type []struct { Rev string "json:\"rev\"" }`,
		},
		{
			name: "success",
			input: `[
                    {"rev": "6-460637e73a6288cb24d532bf91f32969"},
                    {"rev": "5-eeaa298781f60b7bcae0c91bdedd1b87"}
                ]`,
			expected: ChangedRevs{"6-460637e73a6288cb24d532bf91f32969", "5-eeaa298781f60b7bcae0c91bdedd1b87"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result ChangedRevs
			err := json.Unmarshal([]byte(test.input), &result)
			testy.Error(t, test.err, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}
