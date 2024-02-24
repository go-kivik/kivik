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
	/*
		TODO:
		- Add an attachment to an existing document
		- Don't delete existing attachments when adding a new one with this method
		- Update an existing attachment
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
