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
	/*
		TODO:
		- leaf rev is deleted -- returned as usual
		- No revs provided == returns winning leaf
		- document not found, open_revs=["something"] = 200 + missing
		- document found, rev not found
		- all revs
		- latest=true
		- non-leaf rev specified -- missing, I think
		- Include attachment info when relevant (https://docs.couchdb.org/en/stable/replication/protocol.html#:~:text=In%20case%20the%20Document%20contains%20attachments%2C%20Source%20MUST%20return%20information%20only%20for%20those%20ones%20that%20had%20been%20changed%20(added%20or%20updated)%20since%20the%20specified%20Revision%20values.%20If%20an%20attachment%20was%20deleted%2C%20the%20Document%20MUST%20NOT%20have%20stub%20information%20for%20it)

		Do other GET options have any effect?
		- attachments - No
		- att_encoding_info - No
		- atts_since - No
		- conflicts
		- deleted_conflicts
		- local_seq
		- meta
		- rev - no
		- revs
		- revs_info
		- If-None-Match header
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
