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
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

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
						AND rev_id=1
						AND rev='9bb58f26192e4ba00f01e2e7b136bbd8'`).Scan(&doc)
				if err != nil {
					t.Fatal(err)
				}
				if doc != `{"foo":"bar"}` {
					t.Errorf("Unexpected doc: %s", doc)
				}
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
			rev, err := db.Put(context.Background(), tt.docID, tt.doc, opts)
			if !testy.ErrorMatches(tt.wantErr, err) {
				t.Errorf("Unexpected error: %s", err)
			}
			if tt.check != nil {
				tt.check(t, db)
			}
			if err != nil {
				return
			}
			if rev != tt.wantRev {
				t.Errorf("Unexpected rev: %s, want %s", rev, tt.wantRev)
			}
		})
	}
}
