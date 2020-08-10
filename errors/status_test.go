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

package errors

import "testing"

func TestStatusText(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected string
	}{
		{
			name:     "Network Error",
			code:     601,
			expected: "network_error",
		},
		{
			name:     "undefined",
			code:     999,
			expected: "unknown",
		},
		{
			name:     "Bad API Call",
			code:     604,
			expected: "bad_api_call",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := statusText(test.code)
			if test.expected != result {
				t.Errorf("Unexpected result: %s", result)
			}
		})
	}
}
