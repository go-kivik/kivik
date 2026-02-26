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

package js

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"
)

func TestUpdate(t *testing.T) {
	t.Parallel()

	type test struct {
		code       string
		doc        any
		req        any
		wantNewDoc any
		wantResp   string
		wantErr    string
	}

	tests := testy.NewTable()
	tests.Add("sets updated field and returns OK", test{
		code:       `function(doc, req) { doc.updated = true; return [doc, "OK"]; }`,
		doc:        map[string]any{"_id": "foo"},
		req:        map[string]any{},
		wantNewDoc: map[string]any{"_id": "foo", "updated": true},
		wantResp:   "OK",
	})
	tests.Add("returns null doc", test{
		code:       `function(doc, req) { return [null, "no change"]; }`,
		doc:        map[string]any{"_id": "foo"},
		req:        map[string]any{},
		wantNewDoc: nil,
		wantResp:   "no change",
	})

	tests.Add("compile error", test{
		code:    `not valid javascript`,
		wantErr: "failed to compile update function",
	})
	// TODO: JS function throws an exception
	// TODO: JS function returns non-array (e.g. a string)
	// TODO: JS function returns array with wrong length (e.g. [doc])
	// TODO: response element is non-string (e.g. numeric) — should coerce or error?
	// TODO: null doc input (doc parameter is nil)

	tests.Run(t, func(t *testing.T, tt test) {
		fn, err := Update(tt.code)
		if !testy.ErrorMatchesRE(tt.wantErr, err) {
			t.Fatalf("Update() error = %v, wantErr /%s/", err, tt.wantErr)
		}
		if err != nil {
			return
		}

		gotNewDoc, gotResp, err := fn(tt.doc, tt.req)
		if !testy.ErrorMatchesRE(tt.wantErr, err) {
			t.Errorf("unexpected error: %v, want /%s/", err, tt.wantErr)
		}
		if err != nil {
			return
		}
		if d := cmp.Diff(tt.wantResp, gotResp); d != "" {
			t.Errorf("response mismatch (-want +got):\n%s", d)
		}
		if d := cmp.Diff(tt.wantNewDoc, gotNewDoc); d != "" {
			t.Errorf("newDoc mismatch (-want +got):\n%s", d)
		}
	})
}
