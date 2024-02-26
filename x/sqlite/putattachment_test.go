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
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

func TestDBPutAttachment(t *testing.T) {
	t.Parallel()
	type test struct {
		setup           func(*testing.T, driver.DB)
		docID           string
		attachment      *driver.Attachment
		options         driver.Options
		check           func(*testing.T, driver.DB)
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
		wantRev: "1-99914b932bd37a50b983c5e7c90ae93b",
		wantRevs: []leaf{
			{
				ID:    "foo",
				Rev:   1,
				RevID: "99914b932bd37a50b983c5e7c90ae93b",
			},
		},
		wantAttachments: []attachmentRow{
			{
				DocID:       "foo",
				Rev:         1,
				RevID:       "99914b932bd37a50b983c5e7c90ae93b",
				Filename:    "foo.txt",
				ContentType: "text/plain",
				Digest:      "md5-bNNVbesNpUvKBgtMOUeYOQ==",
				Length:      13,
				Data:        "Hello, world!",
			},
		},
	})
	tests.Add("add attachment to existing doc", test{
		setup: func(t *testing.T, db driver.DB) {
			_, err := db.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		docID: "foo",
		attachment: &driver.Attachment{
			Filename:    "foo.txt",
			ContentType: "text/plain",
			Content:     io.NopCloser(strings.NewReader("Hello, world!")),
		},
		options: kivik.Rev("1-9bb58f26192e4ba00f01e2e7b136bbd8"),
		wantRev: "2-9bb58f26192e4ba00f01e2e7b136bbd8",
		wantRevs: []leaf{
			{
				ID:    "foo",
				Rev:   1,
				RevID: "9bb58f26192e4ba00f01e2e7b136bbd8",
			},
			{
				ID:          "foo",
				Rev:         2,
				RevID:       "9bb58f26192e4ba00f01e2e7b136bbd8",
				ParentRev:   &[]int{1}[0],
				ParentRevID: &[]string{"9bb58f26192e4ba00f01e2e7b136bbd8"}[0],
			},
		},
		wantAttachments: []attachmentRow{
			{
				DocID:       "foo",
				Rev:         2,
				RevID:       "9bb58f26192e4ba00f01e2e7b136bbd8",
				Filename:    "foo.txt",
				ContentType: "text/plain",
				Digest:      "md5-bNNVbesNpUvKBgtMOUeYOQ==",
				Length:      13,
				Data:        "Hello, world!",
			},
		},
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
		wantErr:    "conflict",
	})
	tests.Add("existing doc, wrong rev", test{
		setup: func(t *testing.T, db driver.DB) {
			_, err := db.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		docID: "foo",
		attachment: &driver.Attachment{
			Filename:    "foo.txt",
			ContentType: "text/plain",
			Content:     io.NopCloser(strings.NewReader("Hello, world!")),
		},
		options:    kivik.Rev("1-wrong"),
		wantStatus: http.StatusConflict,
		wantErr:    "conflict",
	})
	tests.Add("don't delete existing attachment", test{
		setup: func(t *testing.T, db driver.DB) {
			_, err := db.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "bar",
				"_attachments": map[string]interface{}{
					"foo.txt": map[string]interface{}{
						"content_type": "text/plain",
						"data":         "SGVsbG8sIHdvcmxkIQ==",
					},
				},
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		docID: "foo",
		attachment: &driver.Attachment{
			Filename:    "bar.txt",
			ContentType: "text/plain",
			Content:     io.NopCloser(strings.NewReader("Hello, world!")),
		},
		options: kivik.Rev("1-53929381825df5c0a2b57f34d168999d"),
		wantRev: "2-53929381825df5c0a2b57f34d168999d",
		wantRevs: []leaf{
			{
				ID:    "foo",
				Rev:   1,
				RevID: "53929381825df5c0a2b57f34d168999d",
			},
			{
				ID:          "foo",
				Rev:         2,
				RevID:       "53929381825df5c0a2b57f34d168999d",
				ParentRev:   &[]int{1}[0],
				ParentRevID: &[]string{"53929381825df5c0a2b57f34d168999d"}[0],
			},
		},
		wantAttachments: []attachmentRow{
			{
				DocID:       "foo",
				Rev:         1,
				RevID:       "53929381825df5c0a2b57f34d168999d",
				Filename:    "foo.txt",
				ContentType: "text/plain",
				Digest:      "md5-bNNVbesNpUvKBgtMOUeYOQ==",
				Length:      13,
				Data:        "Hello, world!",
			},
			{
				DocID:       "foo",
				Rev:         2,
				RevID:       "53929381825df5c0a2b57f34d168999d",
				Filename:    "bar.txt",
				ContentType: "text/plain",
				Digest:      "md5-bNNVbesNpUvKBgtMOUeYOQ==",
				Length:      13,
				Data:        "Hello, world!",
			},
		},
	})
	tests.Add("update existing attachment", test{
		setup: func(t *testing.T, db driver.DB) {
			_, err := db.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "bar",
				"_attachments": map[string]interface{}{
					"foo.txt": map[string]interface{}{
						"content_type": "text/plain",
						"data":         "SGVsbG8sIHdvcmxkIQ==",
					},
				},
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		docID: "foo",
		attachment: &driver.Attachment{
			Filename:    "foo.txt",
			ContentType: "text/plain",
			Content:     io.NopCloser(strings.NewReader("Hello, everybody!")),
		},
		options: kivik.Rev("1-53929381825df5c0a2b57f34d168999d"),
		wantRev: "2-53929381825df5c0a2b57f34d168999d",
		wantRevs: []leaf{
			{
				ID:    "foo",
				Rev:   1,
				RevID: "53929381825df5c0a2b57f34d168999d",
			},
			{
				ID:          "foo",
				Rev:         2,
				RevID:       "53929381825df5c0a2b57f34d168999d",
				ParentRev:   &[]int{1}[0],
				ParentRevID: &[]string{"53929381825df5c0a2b57f34d168999d"}[0],
			},
		},
		wantAttachments: []attachmentRow{
			{
				DocID:       "foo",
				Rev:         1,
				RevID:       "53929381825df5c0a2b57f34d168999d",
				Filename:    "foo.txt",
				Digest:      "md5-bNNVbesNpUvKBgtMOUeYOQ==",
				Length:      13,
				ContentType: "text/plain",
				Data:        "Hello, world!",
			},
			{
				DocID:       "foo",
				Rev:         2,
				RevID:       "53929381825df5c0a2b57f34d168999d",
				Filename:    "foo.txt",
				ContentType: "text/plain",
				Digest:      "md5-kDqL1OTtoET1YR0WdPZ5tQ==",
				Length:      17,
				Data:        "Hello, everybody!",
			},
		},
	})

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
		rev, err := dbc.PutAttachment(context.Background(), tt.docID, tt.attachment, opts)
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
