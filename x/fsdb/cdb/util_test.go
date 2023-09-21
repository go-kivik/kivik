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

package cdb

import (
	"testing"

	"gitlab.com/flimzy/testy"
)

func TestEscape(t *testing.T) {
	type tt struct {
		in   string
		want string
	}
	tests := testy.NewTable()
	tests.Add("simple", tt{"simple", "simple"})
	tests.Add("non-ascii", tt{"fóò", "fóò"})
	tests.Add("ddoc", tt{"_design/foo", "_design%2Ffoo"})
	tests.Add("percent", tt{"100%", "100%"})
	tests.Add("escaped slash", tt{"foo%2fbar", "foo%252fbar"})
	tests.Add("empty", tt{"", ""})

	tests.Run(t, func(t *testing.T, tt tt) {
		got := EscapeID(tt.in)
		if got != tt.want {
			t.Errorf("Unexpected escape output: %s", got)
		}
		final := UnescapeID(got)
		if final != tt.in {
			t.Errorf("Unexpected unescape output: %s", final)
		}
	})
}
