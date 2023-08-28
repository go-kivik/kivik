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
	"testing"
)

func TestEncodeDocID(t *testing.T) {
	tests := []struct {
		Input    string
		Expected string
	}{
		{Input: "foo", Expected: "foo"},
		{Input: "foo/bar", Expected: "foo%2Fbar"},
		{Input: "_design/foo", Expected: "_design/foo"},
		{Input: "_design/foo/bar", Expected: "_design/foo%2Fbar"},
		{Input: "foo@bar.com", Expected: "foo%40bar.com"},
		{Input: "foo+bar@baz.com", Expected: "foo%2Bbar%40baz.com"},
		{Input: "Is this a valid ID?", Expected: "Is%20this%20a%20valid%20ID%3F"},
		{Input: "nón-English-çharacters", Expected: "n%C3%B3n-English-%C3%A7haracters"},
		{Input: "foo+bar & páces?!*,", Expected: "foo%2Bbar%20%26%20p%C3%A1ces%3F%21%2A%2C"},
		{Input: "kivik$1234", Expected: "kivik%241234"},
		{Input: "_users", Expected: "_users"},
	}
	for _, test := range tests {
		result := EncodeDocID(test.Input)
		if result != test.Expected {
			t.Errorf("Unexpected encoded DocID from %s\n\tExpected: %s\n\t  Actual: %s\n", test.Input, test.Expected, result)
		}
	}
}
