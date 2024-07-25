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
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestDBPutAttachment(t *testing.T) {
	t.Parallel()
	type test struct {
		db              *testDB
		docID           string
		attachment      *driver.Attachment
		options         driver.Options
		wantRev         string
		wantRevs        []leaf
		wantStatus      int
		wantErr         string
		wantAttachments []attachmentRow
	}

	tests := testy.NewTable()
	tests.Add("create doc by adding attachment", test{
		docID: "foo",
		attachment: &driver.Attachment{
			Filename:    "foo.txt",
			ContentType: "text/plain",
			Content:     io.NopCloser(strings.NewReader("Hello, world!")),
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
				Digest:   "md5-bNNVbesNpUvKBgtMOUeYOQ==",
			},
		},
	})
	tests.Add("add attachment to existing doc", func(t *testing.T) interface{} {
		db := newDB(t)
		rev := db.tPut("foo", map[string]string{"foo": "bar"})

		return test{
			db:    db,
			docID: "foo",
			attachment: &driver.Attachment{
				Filename:    "foo.txt",
				ContentType: "text/plain",
				Content:     io.NopCloser(strings.NewReader("Hello, world!")),
			},
			options: kivik.Rev(rev),
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
					RevPos:   2,
					Rev:      2,
					Filename: "foo.txt",
					Digest:   "md5-bNNVbesNpUvKBgtMOUeYOQ==",
				},
			},
		}
	})
	tests.Add("non-existing doc with rev should conflict", test{
		docID: "foo",
		attachment: &driver.Attachment{
			Filename:    "foo.txt",
			ContentType: "text/plain",
			Content:     io.NopCloser(strings.NewReader("Hello, world!")),
		},
		options:    kivik.Rev("1-9bb58f26192e4ba00f01e2e7b136bbd8"),
		wantStatus: http.StatusConflict,
		wantErr:    "document update conflict",
	})
	tests.Add("existing doc, wrong rev", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]string{"foo": "bar"})

		return test{
			db:    db,
			docID: "foo",
			attachment: &driver.Attachment{
				Filename:    "foo.txt",
				ContentType: "text/plain",
				Content:     io.NopCloser(strings.NewReader("Hello, world!")),
			},
			options:    kivik.Rev("1-wrong"),
			wantStatus: http.StatusConflict,
			wantErr:    "document update conflict",
		}
	})
	tests.Add("don't delete existing attachment", func(t *testing.T) interface{} {
		db := newDB(t)
		rev := db.tPut("foo", map[string]interface{}{
			"foo":          "bar",
			"_attachments": newAttachments().add("foo.txt", "Hello, world!"),
		})

		return test{
			db:    db,
			docID: "foo",
			attachment: &driver.Attachment{
				Filename:    "bar.txt",
				ContentType: "text/plain",
				Content:     io.NopCloser(strings.NewReader("Hello, world!")),
			},
			options: kivik.Rev(rev),
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
					Digest:   "md5-bNNVbesNpUvKBgtMOUeYOQ==",
				},
				{
					DocID:    "foo",
					RevPos:   2,
					Rev:      2,
					Filename: "bar.txt",
					Digest:   "md5-bNNVbesNpUvKBgtMOUeYOQ==",
				},
			},
		}
	})
	tests.Add("update existing attachment", func(t *testing.T) interface{} {
		db := newDB(t)
		rev := db.tPut("foo", map[string]interface{}{
			"foo":          "bar",
			"_attachments": newAttachments().add("foo.txt", "Hello, world!"),
		})

		return test{
			db:    db,
			docID: "foo",
			attachment: &driver.Attachment{
				Filename:    "foo.txt",
				ContentType: "text/plain",
				Content:     io.NopCloser(strings.NewReader("Hello, everybody!")),
			},
			options: kivik.Rev(rev),
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
					Digest:   "md5-bNNVbesNpUvKBgtMOUeYOQ==",
				},
				{
					DocID:    "foo",
					RevPos:   2,
					Rev:      2,
					Filename: "foo.txt",
					Digest:   "md5-kDqL1OTtoET1YR0WdPZ5tQ==",
				},
			},
		}
	})
	tests.Add("updating attachments on non-winning leaf only alters that revision branch", func(t *testing.T) interface{} {
		d := newDB(t)
		rev1 := d.tPut("foo", map[string]interface{}{
			"cat": "meow",
		})
		rev2 := d.tPut("foo", map[string]interface{}{
			"dog":          "woof",
			"_attachments": newAttachments().add("foo.txt", "This is a base64 encoding"),
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
			db:    d,
			docID: "foo",
			attachment: &driver.Attachment{
				Filename:    "new.txt",
				ContentType: "text/plain",
				Content:     io.NopCloser(strings.NewReader("new data")),
			},
			options: kivik.Rev(rev2),
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
					RevID:       "def",
					ParentRev:   &[]int{1}[0],
					ParentRevID: &r1.id,
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
					Rev:         3,
					RevID:       "abc",
					ParentRev:   &[]int{2}[0],
					ParentRevID: &[]string{"def"}[0],
				},
				{
					ID:          "foo",
					Rev:         3,
					ParentRev:   &[]int{2}[0],
					ParentRevID: &r2.id,
				},
			},
			wantAttachments: []attachmentRow{
				{
					DocID:    "foo",
					RevPos:   2,
					Rev:      2,
					Filename: "foo.txt",
					Digest:   "md5-TmfHxaRgUrE9l3tkAn4s0Q==",
				},
				{
					DocID:    "foo",
					RevPos:   3,
					Rev:      3,
					Filename: "new.txt",
					Digest:   "md5-6DyjmnleVyg+wdEu2g/s0w==",
				},
			},
		}
	})

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
		rev, err := dbc.PutAttachment(context.Background(), tt.docID, tt.attachment, opts)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status: %d", status)
		}
		if err != nil {
			return
		}
		if !regexp.MustCompile(tt.wantRev).MatchString(rev) {
			t.Errorf("Unexpected rev: %s, want %s", rev, tt.wantRev)
		}
		if len(tt.wantRevs) == 0 {
			t.Errorf("No leaves to check")
		}
		leaves := readRevisions(t, dbc.underlying())
		for i, r := range tt.wantRevs {
			// allow tests to omit RevID
			if r.RevID == "" {
				leaves[i].RevID = ""
			}
			if r.ParentRevID == nil {
				leaves[i].ParentRevID = nil
			}
		}
		if d := cmp.Diff(tt.wantRevs, leaves); d != "" {
			t.Errorf("Unexpected leaves: %s", d)
		}
		checkAttachments(t, dbc.underlying(), tt.wantAttachments)
	})
}
