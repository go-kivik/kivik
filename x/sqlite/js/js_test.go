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
	"time"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"
)

func TestMap(t *testing.T) {
	t.Parallel()

	type test struct {
		code           string
		emit           func(key, value any)
		timeout        time.Duration
		doc            any
		wantCompileErr string
		wantErr        string
	}

	tests := testy.NewTable()
	tests.Add("infinite loop times out", test{
		code:    `function(doc) { while(true) {} }`,
		emit:    func(key, value any) {},
		timeout: 50 * time.Millisecond,
		doc:     map[string]any{"_id": "foo"},
		wantErr: "timeout",
	})

	tests.Run(t, func(t *testing.T, tt test) {
		fn, err := jsMap(tt.code, tt.emit, tt.timeout)
		if !testy.ErrorMatchesRE(tt.wantCompileErr, err) {
			t.Fatalf("Map() error = %v, wantCompileErr /%s/", err, tt.wantCompileErr)
		}
		if err != nil {
			return
		}

		err = fn(tt.doc)
		if !testy.ErrorMatchesRE(tt.wantErr, err) {
			t.Fatalf("fn() error = %v, wantErr /%s/", err, tt.wantErr)
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	type test struct {
		code           string
		doc            any
		req            any
		wantNewDoc     any
		wantResp       string
		wantCompileErr string
		wantErr        string
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
		code:           `not valid javascript`,
		wantCompileErr: "failed to compile update function",
	})
	tests.Add("JS exception", test{
		code:    `function(doc, req) { throw "something went wrong"; }`,
		doc:     map[string]any{},
		req:     map[string]any{},
		wantErr: "something went wrong",
	})
	tests.Add("returns non-array", test{
		code:    `function(doc, req) { return "oops"; }`,
		doc:     map[string]any{},
		req:     map[string]any{},
		wantErr: `update function must return \[doc, response\]`,
	})
	tests.Add("returns wrong length array", test{
		code:    `function(doc, req) { return [doc]; }`,
		doc:     map[string]any{},
		req:     map[string]any{},
		wantErr: `update function must return \[doc, response\]`,
	})
	tests.Add("non-string response", test{
		// TODO: non-string response is silently lost; should coerce via fmt.Sprint
		code:       `function(doc, req) { return [doc, 42]; }`,
		doc:        map[string]any{"_id": "foo"},
		req:        map[string]any{},
		wantNewDoc: map[string]any{"_id": "foo"},
		wantResp:   "",
	})
	tests.Add("null doc input", test{
		code:       `function(doc, req) { return [{"created": true}, "created"]; }`,
		doc:        nil,
		req:        map[string]any{},
		wantNewDoc: map[string]any{"created": true},
		wantResp:   "created",
	})

	tests.Run(t, func(t *testing.T, tt test) {
		fn, err := Update(tt.code)
		if !testy.ErrorMatchesRE(tt.wantCompileErr, err) {
			t.Fatalf("Update() error = %v, wantCompileErr /%s/", err, tt.wantCompileErr)
		}
		if err != nil {
			return
		}

		gotNewDoc, gotResp, err := fn(tt.doc, tt.req)
		if !testy.ErrorMatchesRE(tt.wantErr, err) {
			t.Fatalf("fn() error = %v, wantErr /%s/", err, tt.wantErr)
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
