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
	"database/sql"
	"net/http"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

type ddoc struct {
	ID         string
	Rev        int
	RevID      string
	Lang       string
	FuncType   string
	FuncName   string
	FuncBody   string
	AutoUpdate bool
}

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
		wantDDocs       []ddoc
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
	tests.Add("design doc with view function creates .Design entries and map table", func(t *testing.T) interface{} {
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
			wantRev: "1-.*",
			wantRevs: []leaf{
				{ID: "_design/foo", Rev: 1},
			},
			wantDDocs: []ddoc{
				{
					ID:         "_design/foo",
					Rev:        1,
					Lang:       "javascript",
					FuncType:   "map",
					FuncName:   "bar",
					FuncBody:   "function(doc) { emit(doc._id, null); }",
					AutoUpdate: true,
				},
			},
			check: func(t *testing.T) {
				var viewCount int
				err := d.underlying().QueryRow(`
					SELECT COUNT(*)
					FROM sqlite_master
					WHERE type = 'table'
						AND name LIKE 'foo_map_bar_%'
				`).Scan(&viewCount)
				if err != nil {
					t.Fatal(err)
				}
				if viewCount != 1 {
					t.Errorf("Found %d view tables, expected 1", viewCount)
				}
			},
		}
	})
	/*
		TODO:
		- unsupported language? -- ignored?
		- Drop old indexes when a ddoc changes

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
		checkLeaves(t, dbc.underlying(), tt.wantRevs)
		checkAttachments(t, dbc.underlying(), tt.wantAttachments)
		checkDDocs(t, dbc.underlying(), tt.wantDDocs)
	})
}

func checkDDocs(t *testing.T, d *sql.DB, want []ddoc) {
	t.Helper()
	rows, err := d.Query(`
		SELECT id, rev, rev_id, language, func_type, func_name, func_body, auto_update
		FROM test_design
	`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got []ddoc
	for rows.Next() {
		var d ddoc
		if err := rows.Scan(&d.ID, &d.Rev, &d.RevID, &d.Lang, &d.FuncType, &d.FuncName, &d.FuncBody, &d.AutoUpdate); err != nil {
			t.Fatal(err)
		}
		got = append(got, d)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	for i, w := range want {
		if i > len(got)-1 {
			t.Errorf("Missing expected design doc: %+v", w)
			break
		}
		// allow tests to omit RevID
		if w.RevID == "" {
			got[i].RevID = ""
		}
	}
	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("Unexpected design docs:\n%s", d)
	}
}
