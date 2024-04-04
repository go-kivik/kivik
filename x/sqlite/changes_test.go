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

func TestDBChanges(t *testing.T) {
	t.Parallel()
	type test struct {
		db          *testDB
		options     driver.Options
		wantErr     string
		wantStatus  int
		wantChanges []driver.Change
		wantLastSeq *string
		wantETag    *string
	}
	tests := testy.NewTable()
	tests.Add("no changes in db", test{
		wantLastSeq: &[]string{""}[0],
		wantETag:    &[]string{"c7ba27130f956748671e845893fd6b80"}[0],
	})
	tests.Add("one change", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("doc1", map[string]string{"foo": "bar"})
		return test{
			db: d,
			wantChanges: []driver.Change{
				{
					ID:      "doc1",
					Seq:     "1",
					Changes: driver.ChangedRevs{rev},
				},
			},
			wantLastSeq: &[]string{"1"}[0],
			wantETag:    &[]string{"872ccd9c6dce18ce6ea4d5106540f089"}[0],
		}
	})
	tests.Add("deleted event", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("doc1", map[string]string{"foo": "bar"})
		rev2 := d.tDelete("doc1", kivik.Rev(rev))

		return test{
			db: d,
			wantChanges: []driver.Change{
				{
					ID:      "doc1",
					Seq:     "1",
					Changes: driver.ChangedRevs{rev},
				},
				{
					ID:      "doc1",
					Seq:     "2",
					Deleted: true,
					Changes: driver.ChangedRevs{rev2},
				},
			},
			wantLastSeq: &[]string{"2"}[0],
			wantETag:    &[]string{"9562870d7e8245d03c2ac6055dff735f"}[0],
		}
	})
	tests.Add("longpoll", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("doc1", map[string]string{"foo": "bar"})
		rev2 := d.tDelete("doc1", kivik.Rev(rev))

		return test{
			db:      d,
			options: kivik.Param("feed", "longpoll"),
			wantChanges: []driver.Change{
				{
					ID:      "doc1",
					Seq:     "1",
					Changes: driver.ChangedRevs{rev},
				},
				{
					ID:      "doc1",
					Seq:     "2",
					Deleted: true,
					Changes: driver.ChangedRevs{rev2},
				},
			},
			wantLastSeq: &[]string{"2"}[0],
			wantETag:    &[]string{""}[0],
		}
	})
	tests.Add("invalid feed type", test{
		options:    kivik.Param("feed", "invalid"),
		wantErr:    "supported `feed` types: normal, longpoll",
		wantStatus: http.StatusBadRequest,
	})

	/*
		TODO:
		- Set Pending
		- Options
			- doc_ids
			- conflicts
			- descending
			- feed
				- normal
				- longpoll
				- continuous
			- filter
			- heartbeat
			- include_docs
			- attachments
			- att_encoding_info
			- last-event-id
			- limit
			- since
			- style
			- timeout
			- view
			- seq_interval
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
		feed, err := dbc.Changes(context.Background(), opts)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status: %d", status)
		}
		if err != nil {
			return
		}

		// iterate over feed
		var got []driver.Change

	loop:
		for {
			change := driver.Change{}
			err := feed.Next(&change)
			switch err {
			case io.EOF:
				break loop
			case nil:
				// continue
			default:
				t.Fatalf("Next() returned error: %s", err)
			}
			got = append(got, change)
		}

		if d := cmp.Diff(tt.wantChanges, got); d != "" {
			t.Errorf("Unexpected changes:\n%s", d)
		}

		if tt.wantLastSeq != nil {
			got := feed.LastSeq()
			if got != *tt.wantLastSeq {
				t.Errorf("Unexpected LastSeq: %s", got)
			}
		}
		if tt.wantETag != nil {
			got := feed.ETag()
			if got != *tt.wantETag {
				t.Errorf("Unexpected ETag: %s", got)
			}
		}
	})
}
