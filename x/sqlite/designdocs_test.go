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

//go:build !js
// +build !js

package sqlite

import (
	"context"
	"net/http"
	"regexp"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

func TestDBPut_designDocs(t *testing.T) {
	t.Parallel()
	type test struct {
		db              *testDB
		docID           string
		doc             interface{}
		options         driver.Options
		check           func(*testing.T)
		wantRev         string
		wantRevs        []leaf
		wantStatus      int
		wantErr         string
		wantAttachments []attachmentRow
	}
	tests := testy.NewTable()
	tests.Add("design doc with non-string language returns 400", test{
		docID: "_design/foo",
		doc: map[string]interface{}{
			"language": 1234,
		},
		wantStatus: http.StatusBadRequest,
		wantErr:    "json: cannot unmarshal number into Go struct field designDocData.language of type string",
	})
	tests.Add("non-design doc with non-string language value is ok", test{
		docID: "foo",
		doc: map[string]interface{}{
			"language": 1234,
		},
		wantRev: "1-.*",
		wantRevs: []leaf{
			{ID: "foo", Rev: 1},
		},
	})
	tests.Add("design doc with view function creates .Design entries", func(t *testing.T) interface{} {
		d := newDB(t)
		return test{
			db:    d,
			docID: "_design/foo",
			doc: map[string]interface{}{
				"language": "javascript",
				"views": map[string]interface{}{
					"bar": map[string]interface{}{
						"map": "function(doc) { emit(doc._id, null); }",
					},
				},
			},
			check: func(t *testing.T) {
			},
			wantRev: "1-.*",
			wantRevs: []leaf{
				{ID: "_design/foo", Rev: 1},
			},
		}
	})
	/*
		TODO:
		- non-object for func map: 400
		- non-object for func map keys: 400
		- funcmap keys: views, updates, filters
		- validate_doc_update: func
		- validate_doc_update is not function: 400
		- unsupported language? -- ignored?
		- autoupdate
	*/

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		dbc := tt.db
		if dbc == nil {
			dbc = newDB(t)
		}
		opts := tt.options
		if opts == nil {
			opts = mock.NilOption
		}
		rev, err := dbc.Put(context.Background(), tt.docID, tt.doc, opts)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status: %d", status)
		}
		if tt.check != nil {
			tt.check(t)
		}
		if err != nil {
			return
		}
		if !regexp.MustCompile(tt.wantRev).MatchString(rev) {
			t.Errorf("Unexpected rev: %s, want %s", rev, tt.wantRev)
		}
		checkLeaves(t, tt.wantRevs, dbc.underlying())
		checkAttachments(t, dbc.underlying(), tt.wantAttachments)
	})
}
