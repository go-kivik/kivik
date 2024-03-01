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

func TestDBDeleteAttachment(t *testing.T) {
	t.Parallel()
	type test struct {
		setup           func(*testing.T, driver.DB)
		docID           string
		filename        string
		options         driver.Options
		check           func(*testing.T, driver.DB)
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
	tests.Add("doc exists, but no rev provided", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		docID:      "foo",
		filename:   "foo.txt",
		wantErr:    "conflict",
		wantStatus: http.StatusConflict,
	})
	tests.Add("doc exists, but wrong rev provided", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		docID:      "foo",
		filename:   "foo.txt",
		options:    kivik.Rev("1-wrong"),
		wantErr:    "document not found",
		wantStatus: http.StatusNotFound,
	})

	/*
		TODO:
		- db missing => db not found
		- file does not exist => file not found
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
		rev, err := dbc.DeleteAttachment(context.Background(), tt.docID, tt.filename, opts)
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
