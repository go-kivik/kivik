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

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

func TestDBDeleteAttachment(t *testing.T) {
	t.Parallel()
	type test struct {
		db              *testDB
		docID           string
		filename        string
		options         driver.Options
		wantRev         string
		wantRevs        []leaf
		wantStatus      int
		wantErr         string
		wantAttachments []attachmentRow
	}

	tests := testy.NewTable()
	tests.Add("doc not found", test{
		docID:      "foo",
		filename:   "foo.txt",
		options:    kivik.Rev("1-9bb58f26192e4ba00f01e2e7b136bbd8"),
		wantErr:    "document not found",
		wantStatus: http.StatusNotFound,
	})
	tests.Add("doc exists, but no rev provided", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]string{"foo": "bar"})

		return test{
			db:         db,
			docID:      "foo",
			filename:   "foo.txt",
			wantErr:    "conflict",
			wantStatus: http.StatusConflict,
		}
	})
	tests.Add("doc exists, but wrong rev provided", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]string{"foo": "bar"})

		return test{
			db:         db,
			docID:      "foo",
			filename:   "foo.txt",
			options:    kivik.Rev("1-wrong"),
			wantErr:    "conflict",
			wantStatus: http.StatusConflict,
		}
	})
	tests.Add("success", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("foo", map[string]interface{}{
			"cat": "meow",
			"_attachments": map[string]interface{}{
				"foo.txt": map[string]interface{}{
					"content_type": "text/plain",
					"data":         "VGhpcyBpcyBhIGJhc2U2NCBlbmNvZGluZw==",
				},
			},
		})

		return test{
			db:       d,
			docID:    "foo",
			filename: "foo.txt",
			options:  kivik.Rev(rev),
			wantRev:  "2-.*",
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
					DocID:       "foo",
					RevPos:      1,
					Rev:         1,
					Filename:    "foo.txt",
					ContentType: "text/plain",
					Length:      25,
					Digest:      "md5-TmfHxaRgUrE9l3tkAn4s0Q==",
					Data:        "This is a base64 encoding",
				},
			},
		}
	})

	/*
		TODO:
		- db missing => db not found
		- file does not exist => file not found
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
		rev, err := dbc.DeleteAttachment(context.Background(), tt.docID, tt.filename, opts)
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
		leaves := readRevisions(t, dbc.underlying(), tt.docID)
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
