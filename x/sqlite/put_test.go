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

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

type attachmentRow struct {
	DocID        string
	Rev          int
	RevID        string
	Filename     string
	ContentType  string
	Digest       string
	Length       int64
	RevPos       int
	Stub         bool
	Data         string
	DeletedRev   *int
	DeletedRevID *string
}

func TestDBPut(t *testing.T) {
	t.Parallel()
	type test struct {
		setup           func(*testing.T, driver.DB)
		docID           string
		doc             interface{}
		options         driver.Options
		check           func(*testing.T, driver.DB)
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
		wantRev: "1-9bb58f26192e4ba00f01e2e7b136bbd8",
		wantRevs: []leaf{
			{
				ID:    "foo",
				Rev:   1,
				RevID: "9bb58f26192e4ba00f01e2e7b136bbd8",
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
		wantErr:    "Document rev and option have different values",
	})
	tests.Add("attempt to create doc with rev in doc should conflict", test{
		docID: "foo",
		doc: map[string]interface{}{
			"_rev": "1-1234567890abcdef1234567890abcdef",
			"foo":  "bar",
		},
		wantStatus: http.StatusConflict,
		wantErr:    "conflict",
	})
	tests.Add("attempt to create doc with rev in params should conflict", test{
		docID: "foo",
		doc: map[string]interface{}{
			"foo": "bar",
		},
		options:    kivik.Rev("1-1234567890abcdef1234567890abcdef"),
		wantStatus: http.StatusConflict,
		wantErr:    "conflict",
	})
	tests.Add("attempt to update doc without rev should conflict", test{
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
	})
	tests.Add("attempt to update doc with wrong rev should conflict", test{
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
	})
	tests.Add("update doc with correct rev", test{
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
	tests.Add("update doc with new_edits=false, existing doc", test{
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
	})
	tests.Add("update doc with new_edits=false, existing doc and rev", test{
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
	tests.Add("set _deleted=true", test{
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
	})
	tests.Add("set _deleted=false", test{
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
	})
	tests.Add("set _deleted=true and new_edits=false", test{
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
		wantErr:    "Document rev and option have different values",
	})
	tests.Add("new_edits=false, with _revisions replayed", test{
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
	})
	tests.Add("new_edits=false, with _revisions and some revs already exist without parents", test{
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
	})
	tests.Add("new_edits=false, with _revisions and some revs already exist with docs", test{
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
		wantErr:    "conflict",
	})
	tests.Add("new_edits=true, with _revisions should conflict for wrong rev", test{
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
	})
	tests.Add("new_edits=true, with _revisions should succeed for correct rev", test{
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
	})
	tests.Add("new_edits=true, with _revisions should succeed for correct history", test{
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
	})
	tests.Add("new_edits=true, with _revisions should fail for wrong history", test{
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
			"_attachments": map[string]interface{}{
				"foo.txt": map[string]interface{}{
					"content_type": "text/plain",
					"data":         "VGhpcyBpcyBhIGJhc2U2NCBlbmNvZGluZw==",
				},
			},
			"foo": "bar",
		},
		wantRev: "1-4b98474b255b67856668474854b0d5f8",
		wantRevs: []leaf{
			{
				ID:    "foo",
				Rev:   1,
				RevID: "4b98474b255b67856668474854b0d5f8",
			},
		},
		wantAttachments: []attachmentRow{
			{
				DocID:       "foo",
				Rev:         1,
				RevID:       "4b98474b255b67856668474854b0d5f8",
				Filename:    "foo.txt",
				ContentType: "text/plain",
				Length:      25,
				Digest:      "md5-TmfHxaRgUrE9l3tkAn4s0Q==",
				Data:        "This is a base64 encoding",
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
		wantRev: "1-1a46dc947908f36db2ac78b7edaecda3",
		wantRevs: []leaf{
			{
				ID:    "foo",
				Rev:   1,
				RevID: "1a46dc947908f36db2ac78b7edaecda3",
			},
		},
		wantAttachments: []attachmentRow{
			{
				DocID:       "foo",
				Rev:         1,
				RevID:       "1a46dc947908f36db2ac78b7edaecda3",
				Filename:    "foo.txt",
				ContentType: "application/octet-stream",
				Length:      25,
				Digest:      "md5-TmfHxaRgUrE9l3tkAn4s0Q==",
				Data:        "This is a base64 encoding",
			},
		},
	})
	tests.Add("update doc with attachments without deleting them", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "bar",
				"_attachments": map[string]interface{}{
					"foo.txt": map[string]interface{}{
						"content_type": "text/plain",
						"data":         "VGhpcyBpcyBhIGJhc2U2NCBlbmNvZGluZw==",
					},
				},
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		docID: "foo",
		doc: map[string]interface{}{
			"_rev": "1-4b98474b255b67856668474854b0d5f8",
			"foo":  "baz",
			"_attachments": map[string]interface{}{
				"foo.txt": map[string]interface{}{
					"stub": true,
				},
			},
		},
		wantRev: "2-a7cadffe4f950734f8eeae832e15f6c2",
		wantRevs: []leaf{
			{
				ID:    "foo",
				Rev:   1,
				RevID: "4b98474b255b67856668474854b0d5f8",
			},
			{
				ID:          "foo",
				Rev:         2,
				RevID:       "a7cadffe4f950734f8eeae832e15f6c2",
				ParentRev:   &[]int{1}[0],
				ParentRevID: &[]string{"4b98474b255b67856668474854b0d5f8"}[0],
			},
		},
		wantAttachments: []attachmentRow{
			{
				DocID:       "foo",
				Rev:         1,
				RevID:       "4b98474b255b67856668474854b0d5f8",
				Filename:    "foo.txt",
				ContentType: "text/plain",
				Length:      25,
				Digest:      "md5-TmfHxaRgUrE9l3tkAn4s0Q==",
				Data:        "This is a base64 encoding",
			},
		},
	})
	tests.Add("update doc with attachments, delete one", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "bar",
				"_attachments": map[string]interface{}{
					"foo.txt": map[string]interface{}{
						"content_type": "text/plain",
						"data":         "VGhpcyBpcyBhIGJhc2U2NCBlbmNvZGluZw==",
					},
					"bar.txt": map[string]interface{}{
						"content_type": "text/plain",
						"data":         "VGhpcyBpcyBhIGJhc2U2NCBlbmNvZGluZw==",
					},
				},
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		docID: "foo",
		doc: map[string]interface{}{
			"_rev": "1-7884bb688778892bd22837c5d8cba96b",
			"foo":  "baz",
			"_attachments": map[string]interface{}{
				"foo.txt": map[string]interface{}{
					"stub": true,
				},
			},
		},
		wantRev: "2-a7cadffe4f950734f8eeae832e15f6c2",
		wantRevs: []leaf{
			{
				ID:    "foo",
				Rev:   1,
				RevID: "7884bb688778892bd22837c5d8cba96b",
			},
			{
				ID:          "foo",
				Rev:         2,
				RevID:       "a7cadffe4f950734f8eeae832e15f6c2",
				ParentRev:   &[]int{1}[0],
				ParentRevID: &[]string{"7884bb688778892bd22837c5d8cba96b"}[0],
			},
		},
		wantAttachments: []attachmentRow{
			{
				DocID:        "foo",
				Rev:          1,
				RevID:        "7884bb688778892bd22837c5d8cba96b",
				Filename:     "bar.txt",
				ContentType:  "text/plain",
				Length:       25,
				Digest:       "md5-TmfHxaRgUrE9l3tkAn4s0Q==",
				Data:         "This is a base64 encoding",
				DeletedRev:   &[]int{2}[0],
				DeletedRevID: &[]string{"a7cadffe4f950734f8eeae832e15f6c2"}[0],
			},
			{
				DocID:       "foo",
				Rev:         1,
				RevID:       "7884bb688778892bd22837c5d8cba96b",
				Filename:    "foo.txt",
				ContentType: "text/plain",
				Length:      25,
				Digest:      "md5-TmfHxaRgUrE9l3tkAn4s0Q==",
				Data:        "This is a base64 encoding",
			},
		},
	})
	/*
		TODO:
		- update conflicting leaf
		- delete attachments only in one branch of a document
		- Omit attachments to delete
		- Include stub to update doc without deleting attachments
		- Include stub with invalid filename
		- Encoding/compression?
		- new_edits=false + attachment
		- new_edits=false + invalid attachment stub
		- filename validation?
	*/

	tests.Run(t, func(t *testing.T, tt test) {
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
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status: %d", status)
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
		checkAttachments(t, dbc, tt.wantAttachments)
	})
}

func checkAttachments(t *testing.T, d driver.DB, want []attachmentRow) {
	t.Helper()
	rows, err := d.(*db).db.Query(`
		SELECT id, rev, rev_id, filename, content_type, length, digest, data, deleted_rev, deleted_rev_id
		FROM test_attachments
		ORDER BY id, rev, rev_id, filename
	`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got []attachmentRow
	for rows.Next() {
		var att attachmentRow
		if err := rows.Scan(&att.DocID, &att.Rev, &att.RevID, &att.Filename, &att.ContentType, &att.Length, &att.Digest, &att.Data, &att.DeletedRev, &att.DeletedRevID); err != nil {
			t.Fatal(err)
		}
		got = append(got, att)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("Unexpected attachments: %s", d)
	}
}
