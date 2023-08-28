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
	"testing"

	"gitlab.com/flimzy/testy"
)

func TestSequenceIDUnmarshal(t *testing.T) {
	tests := []struct {
		name  string
		input string

		expected sequenceID
		err      string
	}{
		{
			name:     "Couch 1.6",
			input:    "123",
			expected: "123",
		},
		{
			name:     "Couch 2.0",
			input:    `"1-seqfoo"`,
			expected: "1-seqfoo",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var seq sequenceID
			err := json.Unmarshal([]byte(test.input), &seq)
			testy.Error(t, test.err, err)
			if seq != test.expected {
				t.Errorf("Unexpected result: %s", seq)
			}
		})
	}
}
