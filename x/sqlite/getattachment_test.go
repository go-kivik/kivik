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
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

func TestDBGetAttachment(t *testing.T) {
	t.Parallel()
	type attachment struct {
		Filename    string
		ContentType string
		Length      int64
		RevPos      int64
		Data        string
	}
	type test struct {
		db       driver.DB
		docID    string
		filename string

		wantAttachment *attachment
		wantStatus     int
		wantErr        string
	}

	tests := testy.NewTable()
	tests.Add("document does not exist", test{
		docID:      "foo",
		filename:   "foo.txt",
		wantStatus: http.StatusNotFound,
		wantErr:    "Not Found: missing",
	})
	tests.Add("when the attachment exists, return it", func(t *testing.T) interface{} {
		db := newDB(t)
		_, err := db.Put(context.Background(), "foo", map[string]interface{}{
			"_id": "foo",
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

		return test{
			db:       db,
			docID:    "foo",
			filename: "foo.txt",
		}
	})
	tests.Add("return an attachment when it exists", func(t *testing.T) interface{} {
		db := newDB(t)
		_, err := db.Put(context.Background(), "foo", map[string]interface{}{
			"_id": "foo",
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

		return test{
			db:       db,
			docID:    "foo",
			filename: "foo.txt",
			wantAttachment: &attachment{
				Filename:    "foo.txt",
				ContentType: "text/plain",
				Length:      25,
				RevPos:      1,
				Data:        "This is a base64 encoding",
			},
		}
	})
	tests.Add("document has been deleted, should return not-found", func(t *testing.T) interface{} {
		db := newDB(t)
		rev, err := db.Put(context.Background(), "foo", map[string]interface{}{
			"_id": "foo",
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
		_, err = db.Delete(context.Background(), "foo", kivik.Rev(rev))
		if err != nil {
			t.Fatal(err)
		}

		return test{
			db:         db,
			docID:      "foo",
			filename:   "foo.txt",
			wantStatus: http.StatusNotFound,
			wantErr:    "Not Found: missing",
		}
	})
	tests.Add("document has been been updated since attachment was added, should succeed", func(t *testing.T) interface{} {
		db := newDB(t)
		rev, err := db.Put(context.Background(), "foo", map[string]interface{}{
			"_id": "foo",
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
		_, err = db.Put(context.Background(), "foo", map[string]interface{}{
			"_id":     "foo",
			"updated": true,
		}, kivik.Rev(rev))
		if err != nil {
			t.Fatal(err)
		}

		return test{
			db:       db,
			docID:    "foo",
			filename: "foo.txt",
			wantAttachment: &attachment{
				Filename:    "foo.txt",
				ContentType: "text/plain",
				Length:      25,
				RevPos:      1,
				Data:        "This is a base64 encoding",
			},
		}
	})
	// GetAttachment returns the latest revision by default
	//

	/*
		TODO:
		- return correct attachment in case of a conflict
		- return existing file from existing doc
		- request attachment from historical revision
		- failure: request attachment from historical revision that does not exist

		- GetAttachment returns 404 when the document does exist, but the attachment has never existed
		- GetAttachment returns 404 when the document has never existed
		- GetAttachment returns 404 when the document was deleted
		- GetAttachment returns 404 when the latest revision was deleted
		- GetAttachment returns 404 when the document does exist, but the attachment has been deleted
		- GetAttachment returns the latest revision
	*/

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		db := tt.db
		if db == nil {
			db = newDB(t)
		}
		// opts := tt.options
		// if opts == nil {
		opts := mock.NilOption
		// }
		att, err := db.GetAttachment(context.Background(), tt.docID, tt.filename, opts)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status: %d", status)
		}

		if tt.wantAttachment == nil {
			return
		}
		data, err := io.ReadAll(att.Content)
		if err != nil {
			t.Fatal(err)
		}
		got := &attachment{
			Filename:    att.Filename,
			ContentType: att.ContentType,
			Length:      att.Size,
			RevPos:      att.RevPos,
			Data:        string(data),
		}
		if d := cmp.Diff(tt.wantAttachment, got); d != "" {
			t.Errorf("Unexpected attachment metadata:\n%s", d)
		}
	})
}
