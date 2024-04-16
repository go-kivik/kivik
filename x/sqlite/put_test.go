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

type attachmentRow struct {
	DocID    string
	Rev      int
	RevID    string
	Filename string
	Digest   string
	RevPos   int
	Stub     bool
}

func TestDBPut(t *testing.T) {
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
	tests.Add("create new document", test{
		docID: "foo",
		doc: map[string]string{
			"foo": "bar",
		},
		wantRev: "1-.*",
		wantRevs: []leaf{
			{
				ID:  "foo",
				Rev: 1,
			},
		},
	})
	tests.Add("doc rev & option rev mismatch", test{
		docID: "foo",
		doc: map[string]interface{}{
			"_rev": "1-1234567890abcdef1234567890abcdef",
			"foo":  "bar",
		},
		options:    driver.Options(kivik.Rev("2-1234567890abcdef1234567890abcdef")),
		wantStatus: http.StatusBadRequest,
		wantErr:    "document rev and option have different values",
	})
	tests.Add("attempt to create doc with rev in doc should conflict", test{
		docID: "foo",
		doc: map[string]interface{}{
			"_rev": "1-1234567890abcdef1234567890abcdef",
			"foo":  "bar",
		},
		wantStatus: http.StatusConflict,
		wantErr:    "document update conflict",
	})
	tests.Add("attempt to create doc with rev in params should conflict", test{
		docID: "foo",
		doc: map[string]interface{}{
			"foo": "bar",
		},
		options:    kivik.Rev("1-1234567890abcdef1234567890abcdef"),
		wantStatus: http.StatusConflict,
		wantErr:    "document update conflict",
	})
	tests.Add("attempt to update doc without rev should conflict", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]string{"foo": "bar"})

		return test{
			db:    db,
			docID: "foo",
			doc: map[string]interface{}{
				"foo": "bar",
			},
			wantStatus: http.StatusConflict,
			wantErr:    "document update conflict",
		}
	})
	tests.Add("attempt to update doc with wrong rev should conflict", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]string{"foo": "bar"})

		return test{
			db:    db,
			docID: "foo",
			doc: map[string]interface{}{
				"_rev": "2-1234567890abcdef1234567890abcdef",
				"foo":  "bar",
			},
			wantStatus: http.StatusConflict,
			wantErr:    "document update conflict",
		}
	})
	tests.Add("update doc with correct rev", func(t *testing.T) interface{} {
		db := newDB(t)
		rev := db.tPut("foo", map[string]string{"foo": "bar"})

		return test{
			db:    db,
			docID: "foo",
			doc: map[string]interface{}{
				"_rev": rev,
				"foo":  "baz",
			},
			wantRev: "2-.*",
			wantRevs: []leaf{
				{
					ID:  "foo",
					Rev: 1,
				},
				{
					ID:        "foo",
					Rev:       2,
					ParentRev: &[]int{1}[0],
				},
			},
		}
	})
	tests.Add("update doc with new_edits=false, no existing doc", test{
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
	})
	tests.Add("update doc with new_edits=false, no rev", test{
		docID: "foo",
		doc: map[string]interface{}{
			"foo": "baz",
		},
		options:    kivik.Param("new_edits", false),
		wantStatus: http.StatusBadRequest,
		wantErr:    "When `new_edits: false`, the document needs `_rev` or `_revisions` specified",
	})
	tests.Add("update doc with new_edits=false, existing doc", func(t *testing.T) interface{} {
		db := newDB(t)
		rev := db.tPut("foo", map[string]string{"foo": "bar"})

		r, _ := parseRev(rev)

		return test{
			db:    db,
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
					RevID: r.id,
				},
				{
					ID:    "foo",
					Rev:   1,
					RevID: "asdf",
				},
			},
		}
	})
	tests.Add("update doc with new_edits=false, existing doc and rev", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("foo", map[string]string{"foo": "bar"})

		r, _ := parseRev(rev)

		return test{
			db:    d,
			docID: "foo",
			doc: map[string]interface{}{
				"_rev": rev,
				"foo":  "baz",
			},
			options: kivik.Param("new_edits", false),
			wantRev: rev,
			check: func(t *testing.T) {
				var doc string
				err := d.underlying().QueryRow(`
					SELECT doc
					FROM test
					WHERE id='foo'
						AND rev=1
						AND rev_id=$1`, r.id).Scan(&doc)
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
					RevID: r.id,
				},
			},
		}
	})
	tests.Add("doc id in url and doc differ", test{
		docID: "foo",
		doc: map[string]interface{}{
			"_id": "bar",
			"foo": "baz",
		},
		wantStatus: http.StatusBadRequest,
		wantErr:    "Document ID must match _id in document",
	})
	tests.Add("set _deleted=true", func(t *testing.T) interface{} {
		d := newDB(t)

		return test{
			db:    d,
			docID: "foo",
			doc: map[string]interface{}{
				"_deleted": true,
				"foo":      "bar",
			},
			wantRev: "1-.*",
			wantRevs: []leaf{
				{
					ID:  "foo",
					Rev: 1,
				},
			},
			check: func(t *testing.T) {
				var deleted bool
				err := d.underlying().QueryRow(`
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
		}
	})
	tests.Add("set _deleted=false", func(t *testing.T) interface{} {
		d := newDB(t)

		return test{
			db:    d,
			docID: "foo",
			doc: map[string]interface{}{
				"_deleted": false,
				"foo":      "bar",
			},
			wantRev: "1-.*",
			wantRevs: []leaf{
				{
					ID:  "foo",
					Rev: 1,
				},
			},
			check: func(t *testing.T) {
				var deleted bool
				err := d.underlying().QueryRow(`
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
		}
	})
	tests.Add("set _deleted=true and new_edits=false", func(t *testing.T) interface{} {
		d := newDB(t)

		return test{
			db:    d,
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
			check: func(t *testing.T) {
				var deleted bool
				err := d.underlying().QueryRow(`
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
		}
	})
	tests.Add("new_edits=false, with _revisions", test{
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
	})
	tests.Add("new_edits=false, with _revisions and _rev, _revisions wins", test{
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
	})
	tests.Add("new_edits=false, with _revisions and query rev, conflict", test{
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
		wantStatus: http.StatusBadRequest,
		wantErr:    "document rev and option have different values",
	})
	tests.Add("new_edits=false, with _revisions replayed", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]interface{}{
			"_revisions": map[string]interface{}{
				"ids":   []string{"ghi", "def", "abc"},
				"start": 3,
			},
			"foo": "bar",
		}, kivik.Param("new_edits", false))

		return test{
			db:    db,
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
		}
	})
	tests.Add("new_edits=false, with _revisions and some revs already exist without parents", func(t *testing.T) interface{} {
		dbc := newDB(t)
		_, err := dbc.underlying().Exec(`
		INSERT INTO test_revs (id, rev, rev_id)
		VALUES ('foo', 1, 'abc'), ('foo', 2, 'def')
	`)
		if err != nil {
			t.Fatal(err)
		}

		return test{
			db:    dbc,
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
		}
	})
	tests.Add("new_edits=false, with _revisions and some revs already exist with docs", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]interface{}{
			"_rev": "2-def",
			"moo":  "oink",
		}, kivik.Param("new_edits", false))

		return test{
			db:    db,
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
		}
	})
	tests.Add("new_edits=true, with _revisions should conflict for new doc", test{
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
		wantErr:    "document update conflict",
	})
	tests.Add("new_edits=true, with _revisions should conflict for wrong rev", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]interface{}{
			"foo": "bar",
		})

		return test{
			db:    db,
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
			wantErr:    "document update conflict",
		}
	})
	tests.Add("new_edits=true, with _revisions should succeed for correct rev", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]interface{}{
			"foo":  "bar",
			"_rev": "1-abc",
		}, kivik.Param("new_edits", false))

		return test{
			db:    db,
			docID: "foo",
			doc: map[string]interface{}{
				"_revisions": map[string]interface{}{
					"ids":   []string{"abc"},
					"start": 1,
				},
				"foo": "bar",
			},
			options: kivik.Param("new_edits", true),
			wantRev: "2-.*",
			wantRevs: []leaf{
				{
					ID:    "foo",
					Rev:   1,
					RevID: "abc",
				},
				{
					ID:          "foo",
					Rev:         2,
					ParentRev:   &[]int{1}[0],
					ParentRevID: &[]string{"abc"}[0],
				},
			},
		}
	})
	tests.Add("new_edits=true, with _revisions should succeed for correct history", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]interface{}{
			"foo": "bar",
			"_revisions": map[string]interface{}{
				"ids":   []string{"ghi", "def", "abc"},
				"start": 3,
			},
		}, kivik.Param("new_edits", false))

		return test{
			db:    db,
			docID: "foo",
			doc: map[string]interface{}{
				"_revisions": map[string]interface{}{
					"ids":   []string{"ghi", "def", "abc"},
					"start": 3,
				},
				"foo": "bar",
			},
			options: kivik.Param("new_edits", true),
			wantRev: "4-.*",
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
					ParentRev:   &[]int{3}[0],
					ParentRevID: &[]string{"ghi"}[0],
				},
			},
		}
	})
	tests.Add("new_edits=true, with _revisions should fail for wrong history", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]interface{}{
			"foo": "bar",
			"_revisions": map[string]interface{}{
				"ids":   []string{"ghi", "def", "abc"},
				"start": 3,
			},
		}, kivik.Param("new_edits", false))

		return test{
			db:    db,
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
			wantErr:    "document update conflict",
		}
	})
	tests.Add("with attachment, no data", test{
		docID: "foo",
		doc: map[string]interface{}{
			"_attachments": map[string]interface{}{
				"foo.txt": map[string]interface{}{},
			},
			"foo": "bar",
		},
		wantStatus: http.StatusBadRequest,
		wantErr:    `invalid attachment data for "foo.txt"`,
	})
	tests.Add("with attachment, data is not base64", test{
		docID: "foo",
		doc: map[string]interface{}{
			"_attachments": map[string]interface{}{
				"foo.txt": map[string]interface{}{
					"data": "This is not base64",
				},
			},
			"foo": "bar",
		},
		wantStatus: http.StatusBadRequest,
		wantErr:    `invalid attachment data for "foo.txt": illegal base64 data at input byte 4`,
	})
	tests.Add("with attachment, data is not a string", test{
		docID: "foo",
		doc: map[string]interface{}{
			"_attachments": map[string]interface{}{
				"foo.txt": map[string]interface{}{
					"data": 1234,
				},
			},
			"foo": "bar",
		},
		wantStatus: http.StatusBadRequest,
		wantErr:    `invalid attachment data for "foo.txt": json: cannot unmarshal number into Go value of type []uint8`,
	})
	tests.Add("with attachment", test{
		docID: "foo",
		doc: map[string]interface{}{
			"_attachments": newAttachments().add("foo.txt", "This is a base64 encoding"),
			"foo":          "bar",
		},
		wantRev: "1-.*",
		wantRevs: []leaf{
			{
				ID:  "foo",
				Rev: 1,
			},
		},
		wantAttachments: []attachmentRow{
			{
				DocID:    "foo",
				RevPos:   1,
				Rev:      1,
				Filename: "foo.txt",
				Digest:   "md5-TmfHxaRgUrE9l3tkAn4s0Q==",
			},
		},
	})
	tests.Add("with attachment, no content-type", test{
		docID: "foo",
		doc: map[string]interface{}{
			"_attachments": map[string]interface{}{
				"foo.txt": map[string]interface{}{
					"data": "VGhpcyBpcyBhIGJhc2U2NCBlbmNvZGluZw==",
				},
			},
			"foo": "bar",
		},
		wantRev: "1-.*",
		wantRevs: []leaf{
			{
				ID:  "foo",
				Rev: 1,
			},
		},
		wantAttachments: []attachmentRow{
			{
				DocID:    "foo",
				RevPos:   1,
				Rev:      1,
				Filename: "foo.txt",
				Digest:   "md5-TmfHxaRgUrE9l3tkAn4s0Q==",
			},
		},
	})
	tests.Add("update doc with attachments without deleting them", func(t *testing.T) interface{} {
		db := newDB(t)
		rev := db.tPut("foo", map[string]interface{}{
			"foo":          "bar",
			"_attachments": newAttachments().add("foo.txt", "This is a base64 encoding"),
		})

		return test{
			db:    db,
			docID: "foo",
			doc: map[string]interface{}{
				"_rev": rev,
				"foo":  "baz",
				"_attachments": map[string]interface{}{
					"foo.txt": map[string]interface{}{
						"stub": true,
					},
				},
			},
			wantRev: "2-.*",
			wantRevs: []leaf{
				{
					ID:  "foo",
					Rev: 1,
				},
				{
					ID:        "foo",
					Rev:       2,
					ParentRev: &[]int{1}[0],
				},
			},
			wantAttachments: []attachmentRow{
				{
					DocID:    "foo",
					RevPos:   1,
					Rev:      1,
					Filename: "foo.txt",
					Digest:   "md5-TmfHxaRgUrE9l3tkAn4s0Q==",
				},
				{
					DocID:    "foo",
					RevPos:   1,
					Rev:      2,
					Filename: "foo.txt",
					Digest:   "md5-TmfHxaRgUrE9l3tkAn4s0Q==",
				},
			},
		}
	})
	tests.Add("update doc with attachments, delete one", func(t *testing.T) interface{} {
		db := newDB(t)
		rev := db.tPut("foo", map[string]interface{}{
			"foo": "bar",
			"_attachments": newAttachments().
				add("foo.txt", "This is a base64 encoding").
				add("bar.txt", "This is a base64 encoding"),
		})

		return test{
			db:    db,
			docID: "foo",
			doc: map[string]interface{}{
				"_rev": rev,
				"foo":  "baz",
				"_attachments": map[string]interface{}{
					"foo.txt": map[string]interface{}{
						"stub": true,
					},
				},
			},
			wantRev: "2-.*",
			wantRevs: []leaf{
				{
					ID:  "foo",
					Rev: 1,
				},
				{
					ID:        "foo",
					Rev:       2,
					ParentRev: &[]int{1}[0],
				},
			},
			wantAttachments: []attachmentRow{
				{
					DocID:    "foo",
					RevPos:   1,
					Rev:      1,
					Filename: "bar.txt",
					Digest:   "md5-TmfHxaRgUrE9l3tkAn4s0Q==",
				},
				{
					DocID:    "foo",
					RevPos:   1,
					Rev:      1,
					Filename: "foo.txt",
					Digest:   "md5-TmfHxaRgUrE9l3tkAn4s0Q==",
				},
				{
					DocID:    "foo",
					RevPos:   1,
					Rev:      2,
					Filename: "foo.txt",
					Digest:   "md5-TmfHxaRgUrE9l3tkAn4s0Q==",
				},
			},
		}
	})
	tests.Add("put with invalid attachment stub returns 412", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("foo", map[string]interface{}{
			"foo": "aaa",
			"_attachments": newAttachments().
				add("att.txt", "att.txt"),
		})

		return test{
			db:      d,
			docID:   "foo",
			options: kivik.Rev(rev),
			doc: map[string]interface{}{
				"_attachments": newAttachments().addStub("invalid.png"),
			},
			wantStatus: http.StatusPreconditionFailed,
			wantErr:    "invalid attachment stub in foo for invalid.png",
		}
	})
	tests.Add("update to conflicting leaf updates the proper branch", func(t *testing.T) interface{} {
		d := newDB(t)
		rev1 := d.tPut("foo", map[string]interface{}{
			"cat": "meow",
		})
		rev2 := d.tPut("foo", map[string]interface{}{
			"dog": "woof",
		}, kivik.Rev(rev1))

		r1, _ := parseRev(rev1)
		r2, _ := parseRev(rev2)

		// Create a conflict
		_ = d.tPut("foo", map[string]interface{}{
			"pig": "oink",
			"_revisions": map[string]interface{}{
				"start": 3,
				"ids":   []string{"abc", "def", r1.id},
			},
		}, kivik.Params(map[string]interface{}{"new_edits": false}))

		return test{
			db:      d,
			docID:   "foo",
			options: kivik.Rev(rev2),
			doc: map[string]interface{}{
				"cow": "moo",
			},
			wantRev: "3-.*",
			wantRevs: []leaf{
				{
					ID:    "foo",
					Rev:   1,
					RevID: r1.id,
				},
				{
					ID:          "foo",
					Rev:         2,
					RevID:       r2.id,
					ParentRev:   &[]int{1}[0],
					ParentRevID: &r1.id,
				},
				{
					ID:          "foo",
					Rev:         2,
					RevID:       "def",
					ParentRev:   &[]int{1}[0],
					ParentRevID: &r1.id,
				},
				{
					ID:          "foo",
					Rev:         3,
					RevID:       "abc",
					ParentRev:   &[]int{2}[0],
					ParentRevID: &[]string{"def"}[0],
				},
				{
					ID:          "foo",
					Rev:         3,
					RevID:       "f99110ca1be121ebf5653ee1ec34610c",
					ParentRev:   &[]int{2}[0],
					ParentRevID: &r2.id,
				},
			},
		}
	})
	tests.Add("new_edits=false with an attachment", test{
		docID: "foo",
		doc: map[string]interface{}{
			"_rev":         "1-abc",
			"_attachments": newAttachments().add("foo.txt", "This is a base64 encoding"),
			"foo":          "bar",
		},
		options: kivik.Param("new_edits", false),
		wantRev: "1-.*",
		wantRevs: []leaf{
			{
				ID:  "foo",
				Rev: 1,
			},
		},
		wantAttachments: []attachmentRow{
			{
				DocID:    "foo",
				RevPos:   1,
				Rev:      1,
				Filename: "foo.txt",
				Digest:   "md5-TmfHxaRgUrE9l3tkAn4s0Q==",
			},
		},
	})
	tests.Add("new_edits=false with an attachment stub and no parent rev results in 412", test{
		docID: "foo",
		doc: map[string]interface{}{
			"_rev":         "1-abc",
			"_attachments": newAttachments().addStub("foo.txt"),
			"foo":          "bar",
		},
		options:    kivik.Param("new_edits", false),
		wantStatus: http.StatusPreconditionFailed,
		wantErr:    "invalid attachment stub in foo for foo.txt",
	})
	tests.Add("new_edits=false with attachment stub and parent in _revisions works", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("foo", map[string]interface{}{
			"_attachments": newAttachments().add("foo.txt", "This is a base64 encoding"),
		})

		r, _ := parseRev(rev)

		return test{
			db:    d,
			docID: "foo",
			doc: map[string]interface{}{
				"_revisions": map[string]interface{}{
					"ids":   []string{"ghi", "def", r.id},
					"start": 3,
				},
				"_attachments": newAttachments().addStub("foo.txt"),
			},
			options: kivik.Param("new_edits", false),
			wantRev: "3-.*",
			wantRevs: []leaf{
				{
					ID:    "foo",
					Rev:   1,
					RevID: r.id,
				},
				{
					ID:          "foo",
					Rev:         2,
					RevID:       "def",
					ParentRev:   &[]int{1}[0],
					ParentRevID: &r.id,
				},
				{
					ID:          "foo",
					Rev:         3,
					RevID:       "ghi",
					ParentRev:   &[]int{2}[0],
					ParentRevID: &[]string{"def"}[0],
				},
			},
			wantAttachments: []attachmentRow{
				{
					DocID:    "foo",
					RevPos:   1,
					Rev:      1,
					Filename: "foo.txt",
					Digest:   "md5-TmfHxaRgUrE9l3tkAn4s0Q==",
				},
				{
					DocID:    "foo",
					RevPos:   1,
					Rev:      3,
					Filename: "foo.txt",
					Digest:   "md5-TmfHxaRgUrE9l3tkAn4s0Q==",
				},
			},
		}
	})
	tests.Add("new_edits=false with attachment stub and no parent in _revisions returns 412", func(t *testing.T) interface{} {
		d := newDB(t)
		_ = d.tPut("foo", map[string]interface{}{
			"_attachments": newAttachments().add("foo.txt", "This is a base64 encoding"),
		})

		return test{
			db:    d,
			docID: "foo",
			doc: map[string]interface{}{
				"_revisions": map[string]interface{}{
					"ids":   []string{"ghi", "def"},
					"start": 6,
				},
				"_attachments": newAttachments().addStub("foo.txt"),
			},
			options:    kivik.Param("new_edits", false),
			wantStatus: http.StatusPreconditionFailed,
			wantErr:    "invalid attachment stub in foo for foo.txt",
		}
	})

	/*
		TODO:
		- Encoding/compression?
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

func checkLeaves(t *testing.T, want []leaf, d *sql.DB) {
	t.Helper()
	if len(want) == 0 {
		t.Errorf("No leaves to check")
	}
	leaves := readRevisions(t, d)
	for i, r := range want {
		// allow tests to omit RevID or ParentRevID
		if r.RevID == "" {
			leaves[i].RevID = ""
		}
		if r.ParentRevID == nil {
			leaves[i].ParentRevID = nil
		}
	}
	if d := cmp.Diff(want, leaves); d != "" {
		t.Errorf("Unexpected leaves: %s", d)
	}
}

func checkAttachments(t *testing.T, d *sql.DB, want []attachmentRow) {
	t.Helper()
	rows, err := d.Query(`
		SELECT b.id, b.rev, b.rev_id, a.rev_pos, a.filename, a.digest
		FROM test_attachments AS a
		JOIN test_attachments_bridge AS b ON b.pk=a.pk
	`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got []attachmentRow
	for rows.Next() {
		var att attachmentRow
		var digest md5sum
		if err := rows.Scan(&att.DocID, &att.Rev, &att.RevID, &att.RevPos, &att.Filename, &digest); err != nil {
			t.Fatal(err)
		}
		att.Digest = digest.Digest()
		got = append(got, att)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	for i, w := range want {
		// allow tests to omit RevID
		if w.RevID == "" {
			got[i].RevID = ""
		}
	}
	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("Unexpected attachments: %s", d)
	}
}

func TestDBPut_updating_a_doc_should_produce_new_rev_id(t *testing.T) {
	t.Parallel()

	d := newDB(t)

	rev := d.tPut("foo", map[string]string{"foo": "bar"})
	rev2 := d.tPut("foo", map[string]string{"foo": "baz"}, kivik.Rev(rev))

	r, _ := parseRev(rev)
	r2, _ := parseRev(rev2)
	if r.id == r2.id {
		t.Fatalf("rev(%s) and rev2(%s) should have different rev ids", rev, rev2)
	}
}
