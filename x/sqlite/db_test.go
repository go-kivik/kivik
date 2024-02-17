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
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

type leaf struct {
	ID          string
	Rev         int
	RevID       string
	ParentRev   *int
	ParentRevID *string
}

func readRevisions(t *testing.T, db *sql.DB, id string) []leaf {
	t.Helper()
	rows, err := db.Query(`
		SELECT id, rev, rev_id, parent_rev, parent_rev_id
		FROM "test_revs"
		WHERE id=$1
		ORDER BY rev, rev_id
	`, id)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var leaves []leaf
	for rows.Next() {
		var l leaf
		if err := rows.Scan(&l.ID, &l.Rev, &l.RevID, &l.ParentRev, &l.ParentRevID); err != nil {
			t.Fatal(err)
		}
		leaves = append(leaves, l)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return leaves
}

func TestDBPut(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		setup      func(*testing.T, driver.DB)
		docID      string
		doc        interface{}
		options    driver.Options
		check      func(*testing.T, driver.DB)
		wantRev    string
		wantRevs   []leaf
		wantStatus int
		wantErr    string
	}{
		{
			name:  "create new document",
			docID: "foo",
			doc: map[string]string{
				"foo": "bar",
			},
			wantRev: "1-9bb58f26192e4ba00f01e2e7b136bbd8",
			wantRevs: []leaf{
				{
					ID:    "foo",
					Rev:   1,
					RevID: "9bb58f26192e4ba00f01e2e7b136bbd8",
				},
			},
		},
		{
			name:  "doc rev & option rev mismatch",
			docID: "foo",
			doc: map[string]interface{}{
				"_rev": "1-1234567890abcdef1234567890abcdef",
				"foo":  "bar",
			},
			options:    driver.Options(kivik.Rev("2-1234567890abcdef1234567890abcdef")),
			wantStatus: http.StatusBadRequest,
			wantErr:    "Document rev and option have different values",
		},
		{
			name:  "attempt to create doc with rev should conflict",
			docID: "foo",
			doc: map[string]interface{}{
				"_rev": "1-1234567890abcdef1234567890abcdef",
				"foo":  "bar",
			},
			wantStatus: http.StatusConflict,
			wantErr:    "conflict",
		},
		{
			name:  "attempt to create doc with rev should conflict",
			docID: "foo",
			doc: map[string]interface{}{
				"foo": "bar",
			},
			options:    kivik.Rev("1-1234567890abcdef1234567890abcdef"),
			wantStatus: http.StatusConflict,
			wantErr:    "conflict",
		},
		{
			name: "attempt to update doc without rev should conflict",
			setup: func(t *testing.T, d driver.DB) {
				if _, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption); err != nil {
					t.Fatal(err)
				}
			},
			docID: "foo",
			doc: map[string]interface{}{
				"foo": "bar",
			},
			wantStatus: http.StatusConflict,
			wantErr:    "conflict",
		},
		{
			name: "attempt to update doc with wrong rev should conflict",
			setup: func(t *testing.T, d driver.DB) {
				if _, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption); err != nil {
					t.Fatal(err)
				}
			},
			docID: "foo",
			doc: map[string]interface{}{
				"_rev": "2-1234567890abcdef1234567890abcdef",
				"foo":  "bar",
			},
			wantStatus: http.StatusConflict,
			wantErr:    "conflict",
		},
		{
			name: "update doc with correct rev",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
				if err != nil {
					t.Fatal(err)
				}
			},
			docID: "foo",
			doc: map[string]interface{}{
				"_rev": "1-9bb58f26192e4ba00f01e2e7b136bbd8",
				"foo":  "baz",
			},
			wantRev: "2-afa7ae8a1906f4bb061be63525974f92",
			wantRevs: []leaf{
				{
					ID:    "foo",
					Rev:   1,
					RevID: "9bb58f26192e4ba00f01e2e7b136bbd8",
				},
				{
					ID:          "foo",
					Rev:         2,
					RevID:       "afa7ae8a1906f4bb061be63525974f92",
					ParentRev:   &[]int{1}[0],
					ParentRevID: &[]string{"9bb58f26192e4ba00f01e2e7b136bbd8"}[0],
				},
			},
		},
		{
			name:  "update doc with new_edits=false, no existing doc",
			docID: "foo",
			doc: map[string]interface{}{
				"_rev": "1-6fe51f74859f3579abaccc426dd5104f",
				"foo":  "baz",
			},
			options: kivik.Param("new_edits", false),
			wantRev: "1-6fe51f74859f3579abaccc426dd5104f",
			wantRevs: []leaf{
				{
					ID:    "foo",
					Rev:   1,
					RevID: "6fe51f74859f3579abaccc426dd5104f",
				},
			},
		},
		{
			name:  "update doc with new_edits=false, no rev",
			docID: "foo",
			doc: map[string]interface{}{
				"foo": "baz",
			},
			options:    kivik.Param("new_edits", false),
			wantStatus: http.StatusBadRequest,
			wantErr:    "When `new_edits: false`, the document needs `_rev` or `_revisions` specified",
		},
		{
			name: "update doc with new_edits=false, existing doc",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
				if err != nil {
					t.Fatal(err)
				}
			},
			docID: "foo",
			doc: map[string]interface{}{
				"_rev": "1-asdf",
				"foo":  "baz",
			},
			options: kivik.Param("new_edits", false),
			wantRev: "1-asdf",
			wantRevs: []leaf{
				{
					ID:    "foo",
					Rev:   1,
					RevID: "9bb58f26192e4ba00f01e2e7b136bbd8",
				},
				{
					ID:    "foo",
					Rev:   1,
					RevID: "asdf",
				},
			},
		},
		{
			name: "update doc with new_edits=false, existing doc and rev",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
				if err != nil {
					t.Fatal(err)
				}
			},
			docID: "foo",
			doc: map[string]interface{}{
				"_rev": "1-9bb58f26192e4ba00f01e2e7b136bbd8",
				"foo":  "baz",
			},
			options: kivik.Param("new_edits", false),
			wantRev: "1-9bb58f26192e4ba00f01e2e7b136bbd8",
			check: func(t *testing.T, d driver.DB) {
				var doc string
				err := d.(*db).db.QueryRow(`
					SELECT doc
					FROM test
					WHERE id='foo'
						AND rev=1
						AND rev_id='9bb58f26192e4ba00f01e2e7b136bbd8'`).Scan(&doc)
				if err != nil {
					t.Fatal(err)
				}
				if doc != `{"foo":"bar"}` {
					t.Errorf("Unexpected doc: %s", doc)
				}
			},
			wantRevs: []leaf{
				{
					ID:    "foo",
					Rev:   1,
					RevID: "9bb58f26192e4ba00f01e2e7b136bbd8",
				},
			},
		},
		{
			name:  "doc id in url and doc differ",
			docID: "foo",
			doc: map[string]interface{}{
				"_id": "bar",
				"foo": "baz",
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    "Document ID must match _id in document",
		},
		{
			name:  "set _deleted=true",
			docID: "foo",
			doc: map[string]interface{}{
				"_deleted": true,
				"foo":      "bar",
			},
			wantRev: "1-6872a0fc474ada5c46ce054b92897063",
			wantRevs: []leaf{
				{
					ID:    "foo",
					Rev:   1,
					RevID: "6872a0fc474ada5c46ce054b92897063",
				},
			},
			check: func(t *testing.T, d driver.DB) {
				var deleted bool
				err := d.(*db).db.QueryRow(`
					SELECT deleted
					FROM test
					WHERE id='foo'
					ORDER BY rev DESC, rev_id DESC
					LIMIT 1
				`).Scan(&deleted)
				if err != nil {
					t.Fatal(err)
				}
				if !deleted {
					t.Errorf("Document not marked deleted")
				}
			},
		},
		{
			name:  "set _deleted=false",
			docID: "foo",
			doc: map[string]interface{}{
				"_deleted": false,
				"foo":      "bar",
			},
			wantRev: "1-9bb58f26192e4ba00f01e2e7b136bbd8",
			wantRevs: []leaf{
				{
					ID:    "foo",
					Rev:   1,
					RevID: "9bb58f26192e4ba00f01e2e7b136bbd8",
				},
			},
			check: func(t *testing.T, d driver.DB) {
				var deleted bool
				err := d.(*db).db.QueryRow(`
					SELECT deleted
					FROM test
					WHERE id='foo'
					ORDER BY rev DESC, rev_id DESC
					LIMIT 1
				`).Scan(&deleted)
				if err != nil {
					t.Fatal(err)
				}
				if deleted {
					t.Errorf("Document marked deleted")
				}
			},
		},
		{
			name:  "set _deleted=true and new_edits=false",
			docID: "foo",
			doc: map[string]interface{}{
				"_deleted": true,
				"foo":      "bar",
				"_rev":     "1-abc",
			},
			options: kivik.Param("new_edits", false),
			wantRev: "1-abc",
			wantRevs: []leaf{
				{
					ID:    "foo",
					Rev:   1,
					RevID: "abc",
				},
			},
			check: func(t *testing.T, d driver.DB) {
				var deleted bool
				err := d.(*db).db.QueryRow(`
					SELECT deleted
					FROM test
					WHERE id='foo'
					ORDER BY rev DESC, rev_id DESC
					LIMIT 1
				`).Scan(&deleted)
				if err != nil {
					t.Fatal(err)
				}
				if !deleted {
					t.Errorf("Document not marked deleted")
				}
			},
		},
		{
			name:  "new_edits=false, with _revisions",
			docID: "foo",
			doc: map[string]interface{}{
				"_revisions": map[string]interface{}{
					"ids":   []string{"ghi", "def", "abc"},
					"start": 3,
				},
				"foo": "bar",
			},
			options: kivik.Param("new_edits", false),
			wantRev: "3-ghi",
			wantRevs: []leaf{
				{
					ID:    "foo",
					Rev:   1,
					RevID: "abc",
				},
				{
					ID:          "foo",
					Rev:         2,
					RevID:       "def",
					ParentRev:   &[]int{1}[0],
					ParentRevID: &[]string{"abc"}[0],
				},
				{
					ID:          "foo",
					Rev:         3,
					RevID:       "ghi",
					ParentRev:   &[]int{2}[0],
					ParentRevID: &[]string{"def"}[0],
				},
			},
		},
		{
			name:  "new_edits=false, with _revisions and _rev, _revisions wins",
			docID: "foo",
			doc: map[string]interface{}{
				"_revisions": map[string]interface{}{
					"ids":   []string{"ghi"},
					"start": 1,
				},
				"_rev": "1-abc",
				"foo":  "bar",
			},
			options: kivik.Param("new_edits", false),
			wantRev: "1-ghi",
			wantRevs: []leaf{
				{
					ID:    "foo",
					Rev:   1,
					RevID: "ghi",
				},
			},
		},
		{
			name:  "new_edits=false, with _revisions and query rev, conflict",
			docID: "foo",
			doc: map[string]interface{}{
				"_revisions": map[string]interface{}{
					"ids":   []string{"ghi"},
					"start": 1,
				},
				"foo": "bar",
			},
			options: kivik.Params(map[string]interface{}{
				"new_edits": false,
				"rev":       "1-abc",
			}),
			wantStatus: http.StatusConflict,
			wantErr:    "Document rev and option have different values",
		},
		{
			name: "new_edits=false, with _revisions replayed",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]interface{}{
					"_revisions": map[string]interface{}{
						"ids":   []string{"ghi", "def", "abc"},
						"start": 3,
					},
					"foo": "bar",
				}, kivik.Param("new_edits", false))
				if err != nil {
					t.Fatal(err)
				}
			},
			docID: "foo",
			doc: map[string]interface{}{
				"_revisions": map[string]interface{}{
					"ids":   []string{"ghi", "def", "abc"},
					"start": 3,
				},
				"foo": "bar",
			},
			options: kivik.Param("new_edits", false),
			wantRev: "3-ghi",
			wantRevs: []leaf{
				{
					ID:    "foo",
					Rev:   1,
					RevID: "abc",
				},
				{
					ID:          "foo",
					Rev:         2,
					RevID:       "def",
					ParentRev:   &[]int{1}[0],
					ParentRevID: &[]string{"abc"}[0],
				},
				{
					ID:          "foo",
					Rev:         3,
					RevID:       "ghi",
					ParentRev:   &[]int{2}[0],
					ParentRevID: &[]string{"def"}[0],
				},
			},
		},
		{
			name: "new_edits=false, with _revisions and some revs already exist without parents",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.(*db).db.Exec(`
					INSERT INTO test_revs (id, rev, rev_id)
					VALUES ('foo', 1, 'abc'), ('foo', 2, 'def')
				`)
				if err != nil {
					t.Fatal(err)
				}
			},
			docID: "foo",
			doc: map[string]interface{}{
				"_revisions": map[string]interface{}{
					"ids":   []string{"ghi", "def", "abc"},
					"start": 3,
				},
				"foo": "bar",
			},
			options: kivik.Param("new_edits", false),
			wantRev: "3-ghi",
			wantRevs: []leaf{
				{
					ID:    "foo",
					Rev:   1,
					RevID: "abc",
				},
				{
					ID:          "foo",
					Rev:         2,
					RevID:       "def",
					ParentRev:   &[]int{1}[0],
					ParentRevID: &[]string{"abc"}[0],
				},
				{
					ID:          "foo",
					Rev:         3,
					RevID:       "ghi",
					ParentRev:   &[]int{2}[0],
					ParentRevID: &[]string{"def"}[0],
				},
			},
		},
		{
			name: "new_edits=false, with _revisions and some revs already exist with docs",
			setup: func(t *testing.T, d driver.DB) {
				if _, err := d.Put(context.Background(), "foo", map[string]interface{}{
					"_rev": "2-def",
					"moo":  "oink",
				}, kivik.Param("new_edits", false)); err != nil {
					t.Fatal(err)
				}
			},
			docID: "foo",
			doc: map[string]interface{}{
				"_revisions": map[string]interface{}{
					"ids":   []string{"ghi", "def", "abc"},
					"start": 3,
				},
				"foo": "bar",
			},
			options: kivik.Param("new_edits", false),
			wantRev: "3-ghi",
			wantRevs: []leaf{
				{
					ID:    "foo",
					Rev:   1,
					RevID: "abc",
				},
				{
					ID:          "foo",
					Rev:         2,
					RevID:       "def",
					ParentRev:   &[]int{1}[0],
					ParentRevID: &[]string{"abc"}[0],
				},
				{
					ID:          "foo",
					Rev:         3,
					RevID:       "ghi",
					ParentRev:   &[]int{2}[0],
					ParentRevID: &[]string{"def"}[0],
				},
			},
		},
		{
			name:  "new_edits=true, with _revisions should conflict for new doc",
			docID: "foo",
			doc: map[string]interface{}{
				"_revisions": map[string]interface{}{
					"ids":   []string{"ghi", "def", "abc"},
					"start": 3,
				},
				"foo": "bar",
			},
			options:    kivik.Param("new_edits", true),
			wantStatus: http.StatusConflict,
			wantErr:    "conflict",
		},
		{
			name: "new_edits=true, with _revisions should conflict for wrong rev",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]interface{}{
					"foo": "bar",
				}, mock.NilOption)
				if err != nil {
					t.Fatal(err)
				}
			},
			docID: "foo",
			doc: map[string]interface{}{
				"_revisions": map[string]interface{}{
					"ids":   []string{"ghi"},
					"start": 1,
				},
				"foo": "bar",
			},
			options:    kivik.Param("new_edits", true),
			wantStatus: http.StatusConflict,
			wantErr:    "conflict",
		},
		{
			name: "new_edits=true, with _revisions should succeed for correct rev",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]interface{}{
					"foo":  "bar",
					"_rev": "1-abc",
				}, kivik.Param("new_edits", false))
				if err != nil {
					t.Fatal(err)
				}
			},
			docID: "foo",
			doc: map[string]interface{}{
				"_revisions": map[string]interface{}{
					"ids":   []string{"abc"},
					"start": 1,
				},
				"foo": "bar",
			},
			options: kivik.Param("new_edits", true),
			wantRev: "2-9bb58f26192e4ba00f01e2e7b136bbd8",
			wantRevs: []leaf{
				{
					ID:    "foo",
					Rev:   1,
					RevID: "abc",
				},
				{
					ID:          "foo",
					Rev:         2,
					RevID:       "9bb58f26192e4ba00f01e2e7b136bbd8",
					ParentRev:   &[]int{1}[0],
					ParentRevID: &[]string{"abc"}[0],
				},
			},
		},
		{
			name: "new_edits=true, with _revisions should succeed for correct history",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]interface{}{
					"foo": "bar",
					"_revisions": map[string]interface{}{
						"ids":   []string{"ghi", "def", "abc"},
						"start": 3,
					},
				}, kivik.Param("new_edits", false))
				if err != nil {
					t.Fatal(err)
				}
			},
			docID: "foo",
			doc: map[string]interface{}{
				"_revisions": map[string]interface{}{
					"ids":   []string{"ghi", "def", "abc"},
					"start": 3,
				},
				"foo": "bar",
			},
			options: kivik.Param("new_edits", true),
			wantRev: "4-9bb58f26192e4ba00f01e2e7b136bbd8",
			wantRevs: []leaf{
				{
					ID:    "foo",
					Rev:   1,
					RevID: "abc",
				},
				{
					ID:          "foo",
					Rev:         2,
					RevID:       "def",
					ParentRev:   &[]int{1}[0],
					ParentRevID: &[]string{"abc"}[0],
				},
				{
					ID:          "foo",
					Rev:         3,
					RevID:       "ghi",
					ParentRev:   &[]int{2}[0],
					ParentRevID: &[]string{"def"}[0],
				},
				{
					ID:          "foo",
					Rev:         4,
					RevID:       "9bb58f26192e4ba00f01e2e7b136bbd8",
					ParentRev:   &[]int{3}[0],
					ParentRevID: &[]string{"ghi"}[0],
				},
			},
		},
		{
			name: "new_edits=true, with _revisions should fail for wrong history",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]interface{}{
					"foo": "bar",
					"_revisions": map[string]interface{}{
						"ids":   []string{"ghi", "def", "abc"},
						"start": 3,
					},
				}, kivik.Param("new_edits", false))
				if err != nil {
					t.Fatal(err)
				}
			},
			docID: "foo",
			doc: map[string]interface{}{
				"_revisions": map[string]interface{}{
					"ids":   []string{"ghi", "xyz", "abc"},
					"start": 3,
				},
				"foo": "bar",
			},
			options:    kivik.Param("new_edits", true),
			wantStatus: http.StatusConflict,
			wantErr:    "conflict",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dbc := newDB(t)
			if tt.setup != nil {
				tt.setup(t, dbc)
			}
			opts := tt.options
			if opts == nil {
				opts = mock.NilOption
			}
			rev, err := dbc.Put(context.Background(), tt.docID, tt.doc, opts)
			if !testy.ErrorMatches(tt.wantErr, err) {
				t.Errorf("Unexpected error: %s", err)
			}
			if tt.check != nil {
				tt.check(t, dbc)
			}
			if err != nil {
				return
			}
			if rev != tt.wantRev {
				t.Errorf("Unexpected rev: %s, want %s", rev, tt.wantRev)
			}
			if len(tt.wantRevs) == 0 {
				t.Errorf("No leaves to check")
			}
			leaves := readRevisions(t, dbc.(*db).db, tt.docID)
			if d := cmp.Diff(tt.wantRevs, leaves); d != "" {
				t.Errorf("Unexpected leaves: %s", d)
			}
		})
	}
}

func TestDBGet(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		setup      func(*testing.T, driver.DB)
		id         string
		options    driver.Options
		wantDoc    interface{}
		wantStatus int
		wantErr    string
	}{
		{
			name:       "not found",
			id:         "foo",
			wantStatus: http.StatusNotFound,
			wantErr:    "not found",
		},
		{
			name: "success",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
				if err != nil {
					t.Fatal(err)
				}
			},
			id: "foo",
			wantDoc: map[string]string{
				"_id":  "foo",
				"_rev": "1-9bb58f26192e4ba00f01e2e7b136bbd8",
				"foo":  "bar",
			},
		},
		{
			name: "get specific rev",
			setup: func(t *testing.T, d driver.DB) {
				rev, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "baz"}, kivik.Rev(rev))
				if err != nil {
					t.Fatal(err)
				}
			},
			id:      "foo",
			options: kivik.Rev("1-9bb58f26192e4ba00f01e2e7b136bbd8"),
			wantDoc: map[string]string{
				"_id":  "foo",
				"_rev": "1-9bb58f26192e4ba00f01e2e7b136bbd8",
				"foo":  "bar",
			},
		},
		{
			name:       "specific rev not found",
			id:         "foo",
			options:    kivik.Rev("1-9bb58f26192e4ba00f01e2e7b136bbd8"),
			wantStatus: http.StatusNotFound,
			wantErr:    "not found",
		},
		{
			name: "include conflicts",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, kivik.Params(map[string]interface{}{
					"new_edits": false,
					"rev":       "1-abc",
				}))
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "baz"}, kivik.Params(map[string]interface{}{
					"new_edits": false,
					"rev":       "1-xyz",
				}))
				if err != nil {
					t.Fatal(err)
				}
			},
			id:      "foo",
			options: kivik.Param("conflicts", true),
			wantDoc: map[string]interface{}{
				"_id":        "foo",
				"_rev":       "1-xyz",
				"foo":        "baz",
				"_conflicts": []string{"1-abc"},
			},
		},
		{
			name: "include only leaf conflicts",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, kivik.Params(map[string]interface{}{
					"new_edits": false,
					"rev":       "1-abc",
				}))
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "baz"}, kivik.Params(map[string]interface{}{
					"new_edits": false,
					"rev":       "1-xyz",
				}))
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "qux"}, kivik.Rev("1-xyz"))
				if err != nil {
					t.Fatal(err)
				}
			},
			id:      "foo",
			options: kivik.Param("conflicts", true),
			wantDoc: map[string]interface{}{
				"_id":        "foo",
				"_rev":       "2-8ecd3d54a4d763ebc0b6e6666d9af066",
				"foo":        "qux",
				"_conflicts": []string{"1-abc"},
			},
		},
		{
			name: "deleted document",
			setup: func(t *testing.T, d driver.DB) {
				rev, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Delete(context.Background(), "foo", kivik.Rev(rev))
				if err != nil {
					t.Fatal(err)
				}
			},
			id:         "foo",
			wantStatus: http.StatusNotFound,
			wantErr:    "not found",
		},
		{
			name: "deleted document by rev",
			setup: func(t *testing.T, d driver.DB) {
				rev, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Delete(context.Background(), "foo", kivik.Rev(rev))
				if err != nil {
					t.Fatal(err)
				}
			},
			id:      "foo",
			options: kivik.Rev("2-df2a4fe30cde39c357c8d1105748d1b9"),
			wantDoc: map[string]interface{}{
				"_id":      "foo",
				"_rev":     "2-df2a4fe30cde39c357c8d1105748d1b9",
				"_deleted": true,
			},
		},
		{
			name: "deleted document with data by rev",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]interface{}{"_deleted": true, "foo": "bar"}, mock.NilOption)
				if err != nil {
					t.Fatal(err)
				}
			},
			id:      "foo",
			options: kivik.Rev("1-6872a0fc474ada5c46ce054b92897063"),
			wantDoc: map[string]interface{}{
				"_id":      "foo",
				"_rev":     "1-6872a0fc474ada5c46ce054b92897063",
				"_deleted": true,
				"foo":      "bar",
			},
		},
		{
			name: "include conflicts, skip deleted conflicts",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]interface{}{"foo": "moo", "_deleted": true}, kivik.Params(map[string]interface{}{
					"new_edits": false,
					"rev":       "1-qwe",
				}))
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, kivik.Params(map[string]interface{}{
					"new_edits": false,
					"rev":       "1-abc",
				}))
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "baz"}, kivik.Params(map[string]interface{}{
					"new_edits": false,
					"rev":       "1-xyz",
				}))
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "qux"}, kivik.Rev("1-xyz"))
				if err != nil {
					t.Fatal(err)
				}
			},
			id:      "foo",
			options: kivik.Param("conflicts", true),
			wantDoc: map[string]interface{}{
				"_id":        "foo",
				"_rev":       "2-8ecd3d54a4d763ebc0b6e6666d9af066",
				"foo":        "qux",
				"_conflicts": []string{"1-abc"},
			},
		},
		{
			name: "include deleted conflicts",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]interface{}{"foo": "moo", "_deleted": true}, kivik.Params(map[string]interface{}{
					"new_edits": false,
					"rev":       "1-qwe",
				}))
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, kivik.Params(map[string]interface{}{
					"new_edits": false,
					"rev":       "1-abc",
				}))
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "baz"}, kivik.Params(map[string]interface{}{
					"new_edits": false,
					"rev":       "1-xyz",
				}))
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "qux"}, kivik.Rev("1-xyz"))
				if err != nil {
					t.Fatal(err)
				}
			},
			id:      "foo",
			options: kivik.Param("deleted_conflicts", true),
			wantDoc: map[string]interface{}{
				"_id":                "foo",
				"_rev":               "2-8ecd3d54a4d763ebc0b6e6666d9af066",
				"foo":                "qux",
				"_deleted_conflicts": []string{"1-qwe"},
			},
		},
		{
			name: "include all conflicts",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]interface{}{"foo": "moo", "_deleted": true}, kivik.Params(map[string]interface{}{
					"new_edits": false,
					"rev":       "1-qwe",
				}))
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, kivik.Params(map[string]interface{}{
					"new_edits": false,
					"rev":       "1-abc",
				}))
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "baz"}, kivik.Params(map[string]interface{}{
					"new_edits": false,
					"rev":       "1-xyz",
				}))
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "qux"}, kivik.Rev("1-xyz"))
				if err != nil {
					t.Fatal(err)
				}
			},
			id: "foo",
			options: kivik.Params(map[string]interface{}{
				"conflicts":         true,
				"deleted_conflicts": true,
			}),
			wantDoc: map[string]interface{}{
				"_id":                "foo",
				"_rev":               "2-8ecd3d54a4d763ebc0b6e6666d9af066",
				"foo":                "qux",
				"_deleted_conflicts": []string{"1-qwe"},
				"_conflicts":         []string{"1-abc"},
			},
		},
		{
			name: "include revs_info",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]interface{}{"foo": "moo", "_deleted": true}, kivik.Params(map[string]interface{}{
					"new_edits": false,
					"rev":       "1-qwe",
				}))
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, kivik.Params(map[string]interface{}{
					"new_edits": false,
					"rev":       "1-abc",
				}))
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "baz"}, kivik.Params(map[string]interface{}{
					"new_edits": false,
					"rev":       "1-xyz",
				}))
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "qux"}, kivik.Rev("1-xyz"))
				if err != nil {
					t.Fatal(err)
				}
			},
			id: "foo",
			options: kivik.Params(map[string]interface{}{
				"revs_info": true,
			}),
			wantDoc: map[string]interface{}{
				"_id":  "foo",
				"_rev": "2-8ecd3d54a4d763ebc0b6e6666d9af066",
				"foo":  "qux",
				"_revs_info": []map[string]string{
					{"rev": "2-8ecd3d54a4d763ebc0b6e6666d9af066", "status": "available"},
					{"rev": "1-xyz", "status": "available"},
				},
			},
		},
		{
			name: "include meta",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]interface{}{"foo": "moo", "_deleted": true}, kivik.Params(map[string]interface{}{
					"new_edits": false,
					"rev":       "1-qwe",
				}))
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, kivik.Params(map[string]interface{}{
					"new_edits": false,
					"rev":       "1-abc",
				}))
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "baz"}, kivik.Params(map[string]interface{}{
					"new_edits": false,
					"rev":       "1-xyz",
				}))
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "qux"}, kivik.Rev("1-xyz"))
				if err != nil {
					t.Fatal(err)
				}
			},
			id:      "foo",
			options: kivik.Param("meta", true),
			wantDoc: map[string]interface{}{
				"_id":  "foo",
				"_rev": "2-8ecd3d54a4d763ebc0b6e6666d9af066",
				"foo":  "qux",
				"_revs_info": []map[string]string{
					{"rev": "2-8ecd3d54a4d763ebc0b6e6666d9af066", "status": "available"},
					{"rev": "1-xyz", "status": "available"},
				},
				"_conflicts":         []string{"1-abc"},
				"_deleted_conflicts": []string{"1-qwe"},
			},
		},
		{
			name: "get latest winning leaf",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]interface{}{"foo": "aaa", "_rev": "1-aaa"}, kivik.Params(map[string]interface{}{
					"new_edits": false,
				}))
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]interface{}{
					"foo": "bbb",
					"_revisions": map[string]interface{}{
						"ids":   []string{"bbb", "aaa"},
						"start": 2,
					},
				}, kivik.Params(map[string]interface{}{
					"new_edits": false,
				}))
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]interface{}{
					"foo": "ddd",
					"_revisions": map[string]interface{}{
						"ids":   []string{"yyy", "aaa"},
						"start": 2,
					},
				}, kivik.Params(map[string]interface{}{
					"new_edits": false,
				}))
				if err != nil {
					t.Fatal(err)
				}
			},
			id: "foo",
			options: kivik.Params(map[string]interface{}{
				"latest": true,
				"rev":    "1-aaa",
			}),
			wantDoc: map[string]interface{}{
				"_id":  "foo",
				"_rev": "2-yyy",
				"foo":  "ddd",
			},
		},
		{
			name: "get latest non-winning leaf",
			setup: func(t *testing.T, d driver.DB) {
				// common root doc
				_, err := d.Put(context.Background(), "foo", map[string]interface{}{"foo": "aaa", "_rev": "1-aaa"}, kivik.Params(map[string]interface{}{
					"new_edits": false,
				}))
				if err != nil {
					t.Fatal(err)
				}
				// losing branch
				_, err = d.Put(context.Background(), "foo", map[string]interface{}{
					"foo": "bbb",
					"_revisions": map[string]interface{}{
						"ids":   []string{"ccc", "bbb", "aaa"},
						"start": 3,
					},
				}, kivik.Params(map[string]interface{}{
					"new_edits": false,
				}))
				if err != nil {
					t.Fatal(err)
				}

				// winning branch
				_, err = d.Put(context.Background(), "foo", map[string]interface{}{
					"foo": "ddd",
					"_revisions": map[string]interface{}{
						"ids":   []string{"xxx", "yyy", "aaa"},
						"start": 3,
					},
				}, kivik.Params(map[string]interface{}{
					"new_edits": false,
				}))
				if err != nil {
					t.Fatal(err)
				}
			},
			id: "foo",
			options: kivik.Params(map[string]interface{}{
				"latest": true,
				"rev":    "2-bbb",
			}),
			wantDoc: map[string]interface{}{
				"_id":  "foo",
				"_rev": "3-ccc",
				"foo":  "bbb",
			},
		},
		{
			name: "get latest rev with deleted leaf, reverts to the winning branch",
			setup: func(t *testing.T, d driver.DB) {
				// common root doc
				_, err := d.Put(context.Background(), "foo", map[string]interface{}{"foo": "aaa", "_rev": "1-aaa"}, kivik.Params(map[string]interface{}{
					"new_edits": false,
				}))
				if err != nil {
					t.Fatal(err)
				}
				// losing branch
				_, err = d.Put(context.Background(), "foo", map[string]interface{}{
					"foo": "bbb",
					"_revisions": map[string]interface{}{
						"ids":   []string{"ccc", "bbb", "aaa"},
						"start": 3,
					},
				}, kivik.Params(map[string]interface{}{
					"new_edits": false,
				}))
				if err != nil {
					t.Fatal(err)
				}
				// now delete the losing leaf
				_, err = d.Delete(context.Background(), "foo", kivik.Rev("3-ccc"))
				if err != nil {
					t.Fatal(err)
				}

				// winning branch
				_, err = d.Put(context.Background(), "foo", map[string]interface{}{
					"foo": "ddd",
					"_revisions": map[string]interface{}{
						"ids":   []string{"xxx", "yyy", "aaa"},
						"start": 3,
					},
				}, kivik.Params(map[string]interface{}{
					"new_edits": false,
				}))
				if err != nil {
					t.Fatal(err)
				}
			},
			id: "foo",
			options: kivik.Params(map[string]interface{}{
				"latest": true,
				"rev":    "2-bbb",
			}),
			wantDoc: map[string]interface{}{
				"_id":  "foo",
				"_rev": "3-xxx",
				"foo":  "ddd",
			},
		},
		{
			name: "revs=true, losing leaf",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]interface{}{"foo": "aaa", "_rev": "1-aaa"}, kivik.Params(map[string]interface{}{
					"new_edits": false,
				}))
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]interface{}{
					"foo": "bbb",
					"_revisions": map[string]interface{}{
						"ids":   []string{"bbb", "aaa"},
						"start": 2,
					},
				}, kivik.Params(map[string]interface{}{
					"new_edits": false,
				}))
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Put(context.Background(), "foo", map[string]interface{}{
					"foo": "ddd",
					"_revisions": map[string]interface{}{
						"ids":   []string{"yyy", "aaa"},
						"start": 2,
					},
				}, kivik.Params(map[string]interface{}{
					"new_edits": false,
				}))
				if err != nil {
					t.Fatal(err)
				}
			},
			id: "foo",
			options: kivik.Params(map[string]interface{}{
				"revs": true,
				"rev":  "2-bbb",
			}),
			wantDoc: map[string]interface{}{
				"_id":  "foo",
				"_rev": "2-bbb",
				"foo":  "bbb",
				"_revisions": map[string]interface{}{
					"start": 2,
					"ids":   []string{"bbb", "aaa"},
				},
			},
		},
		/*
			TODO:
			attachments = true
			att_encoding_info = true
			atts_since = [revs]
			local_seq = true
			open_revs = []
		*/
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := newDB(t)
			if tt.setup != nil {
				tt.setup(t, db)
			}
			opts := tt.options
			if opts == nil {
				opts = mock.NilOption
			}
			doc, err := db.Get(context.Background(), tt.id, opts)
			if !testy.ErrorMatches(tt.wantErr, err) {
				t.Errorf("Unexpected error: %s", err)
			}
			if status := kivik.HTTPStatus(err); status != tt.wantStatus {
				t.Errorf("Unexpected status: %d", status)
			}
			if err != nil {
				return
			}
			var gotDoc interface{}
			if err := json.NewDecoder(doc.Body).Decode(&gotDoc); err != nil {
				t.Fatal(err)
			}
			if d := testy.DiffAsJSON(tt.wantDoc, gotDoc); d != nil {
				t.Errorf("Unexpected doc: %s", d)
			}
		})
	}
}

func TestDBDelete(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		setup      func(*testing.T, driver.DB)
		id         string
		options    driver.Options
		wantRev    string
		check      func(*testing.T, driver.DB)
		wantStatus int
		wantErr    string
	}{
		{
			name:       "not found",
			id:         "foo",
			wantStatus: http.StatusNotFound,
			wantErr:    "not found",
		},
		{
			name: "success",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
				if err != nil {
					t.Fatal(err)
				}
			},
			id:      "foo",
			options: kivik.Rev("1-9bb58f26192e4ba00f01e2e7b136bbd8"),
			wantRev: "2-df2a4fe30cde39c357c8d1105748d1b9",
			check: func(t *testing.T, d driver.DB) {
				var deleted bool
				err := d.(*db).db.QueryRow(`
					SELECT deleted
					FROM test
					WHERE id='foo'
					ORDER BY rev DESC, rev_id DESC
					LIMIT 1
				`).Scan(&deleted)
				if err != nil {
					t.Fatal(err)
				}
				if !deleted {
					t.Errorf("Document not marked deleted")
				}
			},
		},
		{
			name: "replay delete should conflict",
			setup: func(t *testing.T, d driver.DB) {
				rev, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Delete(context.Background(), "foo", kivik.Rev(rev))
				if err != nil {
					t.Fatal(err)
				}
			},
			id:         "foo",
			options:    kivik.Rev("1-9bb58f26192e4ba00f01e2e7b136bbd8"),
			wantStatus: http.StatusConflict,
			wantErr:    "conflict",
		},
		{
			name: "delete deleted doc should succeed",
			setup: func(t *testing.T, d driver.DB) {
				rev, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Delete(context.Background(), "foo", kivik.Rev(rev))
				if err != nil {
					t.Fatal(err)
				}
			},
			id:      "foo",
			options: kivik.Rev("2-df2a4fe30cde39c357c8d1105748d1b9"),
			wantRev: "3-df2a4fe30cde39c357c8d1105748d1b9",
		},
		{
			name: "delete without rev",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
				if err != nil {
					t.Fatal(err)
				}
			},
			id:         "foo",
			wantStatus: http.StatusConflict,
			wantErr:    "conflict",
		},
		/* _revisions */
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := newDB(t)
			if tt.setup != nil {
				tt.setup(t, db)
			}
			opts := tt.options
			if opts == nil {
				opts = mock.NilOption
			}
			rev, err := db.Delete(context.Background(), tt.id, opts)
			if !testy.ErrorMatches(tt.wantErr, err) {
				t.Errorf("Unexpected error: %s", err)
			}
			if status := kivik.HTTPStatus(err); status != tt.wantStatus {
				t.Errorf("Unexpected status: %d", status)
			}
			if err != nil {
				return
			}
			if rev != tt.wantRev {
				t.Errorf("Unexpected rev: %s", rev)
			}
			if tt.check != nil {
				tt.check(t, db)
			}
		})
	}
}
