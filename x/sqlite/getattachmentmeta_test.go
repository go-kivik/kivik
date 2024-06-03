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
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestDBGetAttachmentMeta(t *testing.T) {
	t.Parallel()
	type attachment struct {
		Filename    string
		ContentType string
		Length      int64
		Digest      string
	}
	type test struct {
		db         *testDB
		docID      string
		filename   string
		options    driver.Options
		want       *attachment
		wantErr    string
		wantStatus int
	}
	tests := testy.NewTable()
	tests.Add("document does not exist", test{
		docID:      "foo",
		filename:   "foo.txt",
		wantStatus: http.StatusNotFound,
		wantErr:    "missing",
	})
	tests.Add("return an attachment when it exists", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]interface{}{
			"_id":          "foo",
			"_attachments": newAttachments().add("foo.txt", "This is a base64 encoding"),
		})

		return test{
			db:       db,
			docID:    "foo",
			filename: "foo.txt",
			want: &attachment{
				Filename:    "foo.txt",
				ContentType: "text/plain",
				Length:      25,
				Digest:      "md5-TmfHxaRgUrE9l3tkAn4s0Q==",
			},
		}
	})

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		db := tt.db
		if db == nil {
			db = newDB(t)
		}
		opts := tt.options
		if opts == nil {
			opts = mock.NilOption
		}
		att, err := db.GetAttachmentMeta(context.Background(), tt.docID, tt.filename, opts)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if err != nil {
			return
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status: %d", status)
		}

		if tt.want == nil {
			return
		}

		got := &attachment{
			Filename:    att.Filename,
			ContentType: att.ContentType,
			Length:      att.Size,
			Digest:      att.Digest,
		}
		if d := cmp.Diff(tt.want, got); d != "" {
			t.Errorf("Unexpected attachment metadata:\n%s", d)
		}
	})
}
