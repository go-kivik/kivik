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
)

func TestSecurityMarshalJSON(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		sec := Security{}
		want := "{}"
		got, _ := json.Marshal(sec)
		if string(got) != want {
			t.Errorf("Unexpected output: %s", string(got))
		}
	})
	t.Run("pointer", func(t *testing.T) {
		sec := &Security{}
		want := "{}"
		got, _ := json.Marshal(sec)
		if string(got) != want {
			t.Errorf("Unexpected output: %s", string(got))
		}
	})
	t.Run("admin name, one member role", func(t *testing.T) {
		sec := Security{
			Admins: Members{
				Names: []string{"bob"},
			},
			Members: Members{
				Roles: []string{"users"},
			},
		}
		want := `{"admins":{"names":["bob"]},"members":{"roles":["users"]}}`
		got, _ := json.Marshal(sec)
		if string(got) != want {
			t.Errorf("Unexpected output: %s", string(got))
		}
	})
}
