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

package sqlite

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"
)

func Test_marshalDoc(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantDoc string
		wantID  string
		wantRev string
		wantErr string
	}{
		{
			name: "extract id and rev",
			input: map[string]interface{}{
				"_id":  "foo",
				"_rev": "1-123",
				"foo":  "bar",
			},
			wantDoc: `{"foo":"bar"}`,
			wantID:  "foo",
			wantRev: "1-123",
		},
		{
			name: "non-string",
			input: map[string]interface{}{
				"_id":  123,
				"_rev": 456,
				"foo":  "bar",
			},
			wantErr: `json: cannot unmarshal number into Go struct field ._id of type string`,
		},
		{
			name: "unmarshalable input",
			input: map[string]interface{}{
				"foo": make(chan int),
			},
			wantErr: `json: unsupported type: chan int`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, id, revID, rev, err := marshalDoc(tt.input)
			if !testy.ErrorMatches(tt.wantErr, err) {
				t.Errorf("marshalDoc error: %v", err)
			}
			if d := cmp.Diff(string(doc), tt.wantDoc); d != "" {
				t.Errorf("doc: %s", d)
			}
			if id != tt.wantID {
				t.Errorf("id: got %s, want %s", id, tt.wantID)
			}
			var fullRev string
			if revID != "" || rev != "" {
				fullRev = revID + "-" + rev
			}
			if fullRev != tt.wantRev {
				t.Errorf("rev: got %s, want %s", fullRev, tt.wantRev)
			}
		})
	}
}
