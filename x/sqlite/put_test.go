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
	"github.com/go-kivik/kivik/v4/int/mock"
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
		doc             any
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
		doc: map[string]any{
			"_rev": "1-1234567890abcdef1234567890abcdef",
			"foo":  "bar",
		},
		options:    driver.Options(kivik.Rev("2-1234567890abcdef1234567890abcdef")),
		wantStatus: http.StatusBadRequest,
		wantErr:    "document rev and option have different values",
	})
	tests.Add("attempt to create doc with rev in doc should conflict", test{
		docID: "foo",
		doc: map[string]any{
			"_rev": "1-1234567890abcdef1234567890abcdef",
			"foo":  "bar",
		},
		wantStatus: http.StatusConflict,
		wantErr:    "document update conflict",
	})
	tests.Add("attempt to create doc with rev in params should conflict", test{
		docID: "foo",
		doc: map[string]any{
			"foo": "bar",
		},
		options:    kivik.Rev("1-1234567890abcdef1234567890abcdef"),
		wantStatus: http.StatusConflict,
		wantErr:    "document update conflict",
	})
	tests.Add("attempt to update doc without rev should conflict", func(t *testing.T) any {
		db := newDB(t)
		_ = db.tPut("foo", map[string]string{"foo": "bar"})

		return test{
			db:    db,
			docID: "foo",
			doc: map[string]any{
				"foo": "bar",
			},
			wantStatus: http.StatusConflict,
			wantErr:    "document update conflict",
		}
	})
	tests.Add("attempt to update doc with wrong rev should conflict", func(t *testing.T) any {
		db := newDB(t)
		_ = db.tPut("foo", map[string]string{"foo": "bar"})

		return test{
			db:    db,
			docID: "foo",
			doc: map[string]any{
				"_rev": "2-1234567890abcdef1234567890abcdef",
				"foo":  "bar",
			},
			wantStatus: http.StatusConflict,
			wantErr:    "document update conflict",
		}
	})
	tests.Add("update doc with correct rev", func(t *testing.T) any {
		db := newDB(t)
		rev := db.tPut("foo", map[string]string{"foo": "bar"})

		return test{
			db:    db,
			docID: "foo",
			doc: map[string]any{
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
		doc: map[string]any{
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
		doc: map[string]any{
			"foo": "baz",
		},
		options:    kivik.Param("new_edits", false),
		wantStatus: http.StatusBadRequest,
		wantErr:    "When `new_edits: false`, the document needs `_rev` or `_revisions` specified",
	})
	tests.Add("update doc with new_edits=false, existing doc", func(t *testing.T) any {
		db := newDB(t)
		rev := db.tPut("foo", map[string]string{"foo": "bar"})

		r, _ := parseRev(rev)

		return test{
			db:    db,
			docID: "foo",
			doc: map[string]any{
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
	tests.Add("update doc with new_edits=false, existing doc and rev", func(t *testing.T) any {
		d := newDB(t)
		rev := d.tPut("foo", map[string]string{"foo": "bar"})

		r, _ := parseRev(rev)

		return test{
			db:    d,
			docID: "foo",
			doc: map[string]any{
				"_rev": rev,
				"foo":  "baz",
			},
			options: kivik.Param("new_edits", false),
			wantRev: rev,
			check: func(t *testing.T) {
				var doc string
				err := d.underlying().QueryRow(`
					SELECT doc
					FROM "kivik$test"
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
	tests.Add("leading underscore in ID", test{
		docID: "_badid",
		doc: map[string]string{
			"_id": "_badid",
			"foo": "bar",
		},
		wantStatus: http.StatusBadRequest,
		wantErr:    "only reserved document ids may start with underscore",
	})
	tests.Add("doc id in url and doc differ", test{
		docID: "foo",
		doc: map[string]any{
			"_id": "bar",
			"foo": "baz",
		},
		wantStatus: http.StatusBadRequest,
		wantErr:    "Document ID must match _id in document",
	})
	tests.Add("set _deleted=true", func(t *testing.T) any {
		d := newDB(t)

		return test{
			db:    d,
			docID: "foo",
			doc: map[string]any{
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
					FROM "kivik$test"
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
	tests.Add("set _deleted=false", func(t *testing.T) any {
		d := newDB(t)

		return test{
			db:    d,
			docID: "foo",
			doc: map[string]any{
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
					FROM "kivik$test"
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
	tests.Add("set _deleted=true and new_edits=false", func(t *testing.T) any {
		d := newDB(t)

		return test{
			db:    d,
			docID: "foo",
			doc: map[string]any{
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
					FROM "kivik$test"
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
		doc: map[string]any{
			"_revisions": map[string]any{
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
		doc: map[string]any{
			"_revisions": map[string]any{
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
		doc: map[string]any{
			"_revisions": map[string]any{
				"ids":   []string{"ghi"},
				"start": 1,
			},
			"foo": "bar",
		},
		options: kivik.Params(map[string]any{
			"new_edits": false,
			"rev":       "1-abc",
		}),
		wantStatus: http.StatusBadRequest,
		wantErr:    "document rev and option have different values",
	})
	tests.Add("new_edits=false, with _revisions replayed", func(t *testing.T) any {
		db := newDB(t)
		_ = db.tPut("foo", map[string]any{
			"_revisions": map[string]any{
				"ids":   []string{"ghi", "def", "abc"},
				"start": 3,
			},
			"foo": "bar",
		}, kivik.Param("new_edits", false))

		return test{
			db:    db,
			docID: "foo",
			doc: map[string]any{
				"_revisions": map[string]any{
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
	tests.Add("new_edits=false, with _revisions and some revs already exist without parents", func(t *testing.T) any {
		dbc := newDB(t)
		_, err := dbc.underlying().Exec(`
		INSERT INTO "kivik$test$revs" (id, rev, rev_id)
		VALUES ('foo', 1, 'abc'), ('foo', 2, 'def')
	`)
		if err != nil {
			t.Fatal(err)
		}

		return test{
			db:    dbc,
			docID: "foo",
			doc: map[string]any{
				"_revisions": map[string]any{
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
	tests.Add("new_edits=false, with _revisions and some revs already exist with docs", func(t *testing.T) any {
		db := newDB(t)
		_ = db.tPut("foo", map[string]any{
			"_rev": "2-def",
			"moo":  "oink",
		}, kivik.Param("new_edits", false))

		return test{
			db:    db,
			docID: "foo",
			doc: map[string]any{
				"_revisions": map[string]any{
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
		doc: map[string]any{
			"_revisions": map[string]any{
				"ids":   []string{"ghi", "def", "abc"},
				"start": 3,
			},
			"foo": "bar",
		},
		options:    kivik.Param("new_edits", true),
		wantStatus: http.StatusConflict,
		wantErr:    "document update conflict",
	})
	tests.Add("new_edits=true, with _revisions should conflict for wrong rev", func(t *testing.T) any {
		db := newDB(t)
		_ = db.tPut("foo", map[string]any{
			"foo": "bar",
		})

		return test{
			db:    db,
			docID: "foo",
			doc: map[string]any{
				"_revisions": map[string]any{
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
	tests.Add("new_edits=true, with _revisions should succeed for correct rev", func(t *testing.T) any {
		db := newDB(t)
		_ = db.tPut("foo", map[string]any{
			"foo":  "bar",
			"_rev": "1-abc",
		}, kivik.Param("new_edits", false))

		return test{
			db:    db,
			docID: "foo",
			doc: map[string]any{
				"_revisions": map[string]any{
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
	tests.Add("new_edits=true, with _revisions should succeed for correct history", func(t *testing.T) any {
		db := newDB(t)
		_ = db.tPut("foo", map[string]any{
			"foo": "bar",
			"_revisions": map[string]any{
				"ids":   []string{"ghi", "def", "abc"},
				"start": 3,
			},
		}, kivik.Param("new_edits", false))

		return test{
			db:    db,
			docID: "foo",
			doc: map[string]any{
				"_revisions": map[string]any{
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
	tests.Add("new_edits=true, with _revisions should fail for wrong history", func(t *testing.T) any {
		db := newDB(t)
		_ = db.tPut("foo", map[string]any{
			"foo": "bar",
			"_revisions": map[string]any{
				"ids":   []string{"ghi", "def", "abc"},
				"start": 3,
			},
		}, kivik.Param("new_edits", false))

		return test{
			db:    db,
			docID: "foo",
			doc: map[string]any{
				"_revisions": map[string]any{
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
		doc: map[string]any{
			"_attachments": map[string]any{
				"foo.txt": map[string]any{},
			},
			"foo": "bar",
		},
		wantStatus: http.StatusBadRequest,
		wantErr:    `invalid attachment data for "foo.txt"`,
	})
	tests.Add("with attachment, data is not base64", test{
		docID: "foo",
		doc: map[string]any{
			"_attachments": map[string]any{
				"foo.txt": map[string]any{
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
		doc: map[string]any{
			"_attachments": map[string]any{
				"foo.txt": map[string]any{
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
		doc: map[string]any{
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
		doc: map[string]any{
			"_attachments": map[string]any{
				"foo.txt": map[string]any{
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
	tests.Add("update doc with attachments without deleting them", func(t *testing.T) any {
		db := newDB(t)
		rev := db.tPut("foo", map[string]any{
			"foo":          "bar",
			"_attachments": newAttachments().add("foo.txt", "This is a base64 encoding"),
		})

		return test{
			db:    db,
			docID: "foo",
			doc: map[string]any{
				"_rev": rev,
				"foo":  "baz",
				"_attachments": map[string]any{
					"foo.txt": map[string]any{
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
	tests.Add("update doc with attachments, delete one", func(t *testing.T) any {
		db := newDB(t)
		rev := db.tPut("foo", map[string]any{
			"foo": "bar",
			"_attachments": newAttachments().
				add("foo.txt", "This is a base64 encoding").
				add("bar.txt", "This is a base64 encoding"),
		})

		return test{
			db:    db,
			docID: "foo",
			doc: map[string]any{
				"_rev": rev,
				"foo":  "baz",
				"_attachments": map[string]any{
					"foo.txt": map[string]any{
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
	tests.Add("put with invalid attachment stub returns 412", func(t *testing.T) any {
		d := newDB(t)
		rev := d.tPut("foo", map[string]any{
			"foo": "aaa",
			"_attachments": newAttachments().
				add("att.txt", "att.txt"),
		})

		return test{
			db:      d,
			docID:   "foo",
			options: kivik.Rev(rev),
			doc: map[string]any{
				"_attachments": newAttachments().addStub("invalid.png"),
			},
			wantStatus: http.StatusPreconditionFailed,
			wantErr:    "invalid attachment stub in foo for invalid.png",
		}
	})
	tests.Add("update to conflicting leaf updates the proper branch", func(t *testing.T) any {
		d := newDB(t)
		rev1 := d.tPut("foo", map[string]any{
			"cat": "meow",
		})
		rev2 := d.tPut("foo", map[string]any{
			"dog": "woof",
		}, kivik.Rev(rev1))

		r1, _ := parseRev(rev1)
		r2, _ := parseRev(rev2)

		// Create a conflict
		_ = d.tPut("foo", map[string]any{
			"pig": "oink",
			"_revisions": map[string]any{
				"start": 3,
				"ids":   []string{"abc", "def", r1.id},
			},
		}, kivik.Params(map[string]any{"new_edits": false}))

		return test{
			db:      d,
			docID:   "foo",
			options: kivik.Rev(rev2),
			doc: map[string]any{
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
		doc: map[string]any{
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
		doc: map[string]any{
			"_rev":         "1-abc",
			"_attachments": newAttachments().addStub("foo.txt"),
			"foo":          "bar",
		},
		options:    kivik.Param("new_edits", false),
		wantStatus: http.StatusPreconditionFailed,
		wantErr:    "invalid attachment stub in foo for foo.txt",
	})
	tests.Add("new_edits=false with attachment stub and parent in _revisions works", func(t *testing.T) any {
		d := newDB(t)
		rev := d.tPut("foo", map[string]any{
			"_attachments": newAttachments().add("foo.txt", "This is a base64 encoding"),
		})

		r, _ := parseRev(rev)

		return test{
			db:    d,
			docID: "foo",
			doc: map[string]any{
				"_revisions": map[string]any{
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
	tests.Add("new_edits=false with attachment stub and no parent in _revisions returns 412", func(t *testing.T) any {
		d := newDB(t)
		_ = d.tPut("foo", map[string]any{
			"_attachments": newAttachments().add("foo.txt", "This is a base64 encoding"),
		})

		return test{
			db:    d,
			docID: "foo",
			doc: map[string]any{
				"_revisions": map[string]any{
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

	tests.Add("validate_doc_update rejects document", func(t *testing.T) interface{} {
		d := newDB(t)
		d.tAddValidation("_design/validation", `function(newDoc, oldDoc, userCtx, secObj) { throw({forbidden: "not allowed"}); }`)

		return test{
			db:    d,
			docID: "foo",
			doc: map[string]interface{}{
				"foo": "bar",
			},
			wantStatus: http.StatusForbidden,
			wantErr:    "not allowed",
		}
	})
	tests.Add("validate_doc_update plain string throw", func(t *testing.T) interface{} {
		d := newDB(t)
		d.tAddValidation("_design/validation", `function(newDoc, oldDoc, userCtx, secObj) { throw("plain string error"); }`)

		return test{
			db:    d,
			docID: "foo",
			doc: map[string]interface{}{
				"foo": "bar",
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    "plain string error",
		}
	})
	tests.Add("validate_doc_update unknown key throw", func(t *testing.T) interface{} {
		d := newDB(t)
		d.tAddValidation("_design/validation", `function(newDoc, oldDoc, userCtx, secObj) { throw({custom_key: "some message"}); }`)

		return test{
			db:    d,
			docID: "foo",
			doc: map[string]interface{}{
				"foo": "bar",
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    "some message",
		}
	})
	tests.Add("validate_doc_update receives oldDoc on update", func(t *testing.T) interface{} {
		d := newDB(t)
		d.tAddValidation("_design/validation", `function(newDoc, oldDoc, userCtx, secObj) { if (oldDoc && oldDoc.foo === "bar") throw({forbidden: "cannot update foo=bar docs"}); }`)
		rev := d.tPut("testdoc", map[string]interface{}{"foo": "bar"})

		return test{
			db:    d,
			docID: "testdoc",
			doc: map[string]interface{}{
				"_rev": rev,
				"foo":  "baz",
			},
			wantStatus: http.StatusForbidden,
			wantErr:    "cannot update foo=bar docs",
		}
	})

	tests.Add("validate_doc_update multiple design docs", func(t *testing.T) interface{} {
		d := newDB(t)
		d.tAddValidation("_design/val1", `function(newDoc, oldDoc, userCtx, secObj) { }`)
		d.tAddValidation("_design/val2", `function(newDoc, oldDoc, userCtx, secObj) { throw({forbidden: "blocked by second"}); }`)

		return test{
			db:    d,
			docID: "foo",
			doc: map[string]interface{}{
				"foo": "bar",
			},
			wantStatus: http.StatusForbidden,
			wantErr:    "blocked by second",
		}
	})

	tests.Add("validate_doc_update stored via design doc put is enforced", func(t *testing.T) interface{} {
		d := newDB(t)
		d.tPut("_design/test", map[string]interface{}{
			"validate_doc_update": `function(newDoc) { if(newDoc.blocked) throw({forbidden: "blocked"}); }`,
		})

		d.tPut("ok", map[string]interface{}{})

		return test{
			db:    d,
			docID: "bad",
			doc: map[string]interface{}{
				"blocked": true,
			},
			wantStatus: http.StatusForbidden,
			wantErr:    "blocked",
		}
	})

	tests.Add("validate_doc_update receives admin party userCtx", func(t *testing.T) interface{} {
		d := newDB(t)
		d.tAddValidation("_design/validation", `function(newDoc, oldDoc, userCtx, secObj) {
			if (userCtx.roles.indexOf("_admin") === -1) throw({forbidden: "not admin"});
			if (userCtx.db !== "test") throw({forbidden: "wrong db name"});
		}`)

		return test{
			db:    d,
			docID: "foo",
			doc: map[string]interface{}{
				"foo": "bar",
			},
			wantRev: "1-.*",
			wantRevs: []leaf{
				{
					ID:  "_design/validation",
					Rev: 1,
				},
				{
					ID:  "foo",
					Rev: 1,
				},
			},
		}
	})

	/*
		TODO:
		- with updates function
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
	})
}

func checkLeaves(t *testing.T, d *sql.DB, want []leaf) {
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
		FROM "kivik$test$attachments" AS a
		JOIN "kivik$test$attachments_bridge" AS b ON b.pk=a.pk
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
