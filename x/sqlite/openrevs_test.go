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

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

func TestDBOpenRevs(t *testing.T) {
	t.Parallel()
	type test struct {
		db         *testDB
		docID      string
		revs       []string
		options    driver.Options
		want       []rowResult
		wantErr    string
		wantStatus int
	}
	tests := testy.NewTable()
	tests.Add("all revs, document not found", test{
		docID:      "not there",
		revs:       []string{"all"},
		wantErr:    "missing",
		wantStatus: http.StatusNotFound,
	})
	tests.Add("invalid rev format", func(t *testing.T) interface{} {
		d := newDB(t)
		docID := "foo"
		_ = d.tPut(docID, map[string]string{"foo": "bar"})

		return test{
			db:         d,
			docID:      docID,
			revs:       []string{"oink", "all"},
			wantErr:    "invalid rev format",
			wantStatus: http.StatusBadRequest,
		}
	})
	tests.Add("all, with single rev", func(t *testing.T) interface{} {
		d := newDB(t)
		docID := "foo"
		rev := d.tPut(docID, map[string]string{"foo": "bar"})

		return test{
			db:    d,
			docID: docID,
			revs:  []string{"all"},
			want: []rowResult{
				{ID: docID, Rev: rev, Doc: `{"_id":"` + docID + `","_rev":"` + rev + `","foo":"bar"}`},
			},
		}
	})
	tests.Add("all, with deleted rev", func(t *testing.T) interface{} {
		d := newDB(t)
		docID := "foo"
		rev := d.tPut(docID, map[string]string{"foo": "bar"})
		rev2 := d.tDelete(docID, kivik.Rev(rev))

		return test{
			db:    d,
			docID: docID,
			revs:  []string{"all"},
			want: []rowResult{
				{ID: docID, Rev: rev2, Doc: `{"_id":"` + docID + `","_rev":"` + rev2 + `","_deleted":true}`},
			},
		}
	})
	tests.Add("all, with conflicting leaves", func(t *testing.T) interface{} {
		d := newDB(t)
		docID := "foo"
		rev := d.tPut(docID, map[string]string{"_rev": "1-xyz", "foo": "bar"}, kivik.Param("new_edits", false))
		rev2 := d.tPut(docID, map[string]string{"_rev": "1-abc", "foo": "baz"}, kivik.Param("new_edits", false))

		return test{
			db:    d,
			docID: docID,
			revs:  []string{"all"},
			want: []rowResult{
				{ID: docID, Rev: rev2, Doc: `{"_id":"` + docID + `","_rev":"` + rev2 + `","foo":"baz"}`},
				{ID: docID, Rev: rev, Doc: `{"_id":"` + docID + `","_rev":"` + rev + `","foo":"bar"}`},
			},
		}
	})
	tests.Add("no revs provided returns winning leaf", func(t *testing.T) interface{} {
		d := newDB(t)
		docID := "foo"
		rev := d.tPut(docID, map[string]string{"_rev": "1-xyz", "foo": "bar"}, kivik.Param("new_edits", false))
		_ = d.tPut(docID, map[string]string{"_rev": "1-abc", "foo": "baz"}, kivik.Param("new_edits", false))

		return test{
			db:    d,
			docID: docID,
			revs:  []string{},
			want: []rowResult{
				{ID: docID, Rev: rev, Doc: `{"_id":"` + docID + `","_rev":"` + rev + `","foo":"bar"}`},
			},
		}
	})
	tests.Add("specific rev", func(t *testing.T) interface{} {
		d := newDB(t)
		docID := "foo"
		rev := d.tPut(docID, map[string]string{"foo": "bar"})
		rev2 := d.tPut(docID, map[string]string{"foo": "baz"}, kivik.Rev(rev))

		return test{
			db:    d,
			docID: docID,
			revs:  []string{rev2},
			want: []rowResult{
				{ID: docID, Rev: rev2, Doc: `{"_id":"` + docID + `","_rev":"` + rev2 + `","foo":"baz"}`},
			},
		}
	})
	tests.Add("specific revs, including non-leaf revs", func(t *testing.T) interface{} {
		d := newDB(t)
		docID := "foo"
		rev := d.tPut(docID, map[string]string{"foo": "bar"})
		rev2 := d.tPut(docID, map[string]string{"foo": "baz"}, kivik.Rev(rev))

		return test{
			db:    d,
			docID: docID,
			revs:  []string{rev, rev2},
			want: []rowResult{
				{ID: docID, Rev: rev, Doc: `{"_id":"` + docID + `","_rev":"` + rev + `","foo":"bar"}`},
				{ID: docID, Rev: rev2, Doc: `{"_id":"` + docID + `","_rev":"` + rev2 + `","foo":"baz"}`},
			},
		}
	})
	tests.Add("specific revs, including one that doesn't exist", func(t *testing.T) interface{} {
		d := newDB(t)
		docID := "foo"
		rev := d.tPut(docID, map[string]string{"foo": "bar"})
		rev2 := d.tPut(docID, map[string]string{"foo": "baz"}, kivik.Rev(rev))

		return test{
			db:    d,
			docID: docID,
			revs:  []string{rev, rev2, "99-asdf"},
			want: []rowResult{
				{ID: docID, Rev: rev, Doc: `{"_id":"` + docID + `","_rev":"` + rev + `","foo":"bar"}`},
				{ID: docID, Rev: rev2, Doc: `{"_id":"` + docID + `","_rev":"` + rev2 + `","foo":"baz"}`},
				{ID: docID, Rev: "99-asdf", Error: "missing"},
			},
		}
	})
	/*
		TODO:
		- rev calculation is broken
		- latest=true, returns latest leaf of the same branch only
		- non-leaf rev specified, returns non-leaf if available
		- Include attachment info when relevant (https://docs.couchdb.org/en/stable/replication/protocol.html#:~:text=In%20case%20the%20Document%20contains%20attachments%2C%20Source%20MUST%20return%20information%20only%20for%20those%20ones%20that%20had%20been%20changed%20(added%20or%20updated)%20since%20the%20specified%20Revision%20values.%20If%20an%20attachment%20was%20deleted%2C%20the%20Document%20MUST%20NOT%20have%20stub%20information%20for%20it)

		- revs=true
	*/

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

		rows, err := db.OpenRevs(context.Background(), tt.docID, tt.revs, opts)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status: %d", status)
		}
		if err != nil {
			return
		}

		checkRows(t, rows, tt.want)
	})
}
