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
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

func TestDBChanges(t *testing.T) {
	t.Parallel()
	type test struct {
		db            *testDB
		ctx           context.Context
		options       driver.Options
		wantErr       string
		wantStatus    int
		wantChanges   []driver.Change
		wantChangesFn func() []driver.Change
		wantNextErr   string
		wantLastSeq   *string
		wantETag      *string
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
	tests.Add("since=1", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("doc1", map[string]string{"foo": "bar"})
		rev2 := d.tDelete("doc1", kivik.Rev(rev))

		return test{
			db:      d,
			options: kivik.Param("since", "1"),
			wantChanges: []driver.Change{
				{
					ID:      "doc1",
					Seq:     "2",
					Deleted: true,
					Changes: driver.ChangedRevs{rev2},
				},
			},
			wantLastSeq: &[]string{"2"}[0],
			wantETag:    &[]string{"bf701dae9aff5bb22b8f000dc9bf6199"}[0],
		}
	})
	tests.Add("malformed sequence id", test{
		options:    kivik.Param("since", "invalid"),
		wantErr:    "malformed sequence supplied in 'since' parameter",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("future since value returns only latest change", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("doc1", map[string]string{"foo": "bar"})
		rev2 := d.tDelete("doc1", kivik.Rev(rev))

		return test{
			db:      d,
			options: kivik.Param("since", "9000"),
			wantChanges: []driver.Change{
				{
					ID:      "doc1",
					Seq:     "2",
					Deleted: true,
					Changes: driver.ChangedRevs{rev2},
				},
			},
			wantLastSeq: &[]string{"2"}[0],
			wantETag:    &[]string{"bf701dae9aff5bb22b8f000dc9bf6199"}[0],
		}
	})
	tests.Add("future since value returns only latest change, longpoll mode", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("doc1", map[string]string{"foo": "bar"})
		rev2 := d.tDelete("doc1", kivik.Rev(rev))

		return test{
			db: d,
			options: kivik.Params(map[string]interface{}{
				"since": "9000",
				"feed":  "longpoll",
			}),
			wantChanges: []driver.Change{
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
	tests.Add("invalid limit value", test{
		options:    kivik.Param("limit", "invalid"),
		wantErr:    "malformed 'limit' parameter",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("longpoll + since in past should return all historical changes since that seqid", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("doc1", map[string]string{"foo": "bar"})
		rev2 := d.tDelete("doc1", kivik.Rev(rev))
		rev3 := d.tPut("doc2", map[string]string{"foo": "bar"})

		return test{
			db: d,
			options: kivik.Params(map[string]interface{}{
				"since": "1",
				"feed":  "longpoll",
			}),
			wantChanges: []driver.Change{
				{
					ID:      "doc1",
					Seq:     "2",
					Deleted: true,
					Changes: driver.ChangedRevs{rev2},
				},
				{
					ID:      "doc2",
					Seq:     "3",
					Changes: driver.ChangedRevs{rev3},
				},
			},
			wantLastSeq: &[]string{"3"}[0],
			wantETag:    &[]string{""}[0],
		}
	})
	tests.Add("feed=normal, context cancellation", func(t *testing.T) interface{} {
		d := newDB(t)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		return test{
			db:  d,
			ctx: ctx,
			options: kivik.Params(map[string]interface{}{
				"feed": "normal",
			}),
			wantErr:    "context canceled",
			wantStatus: http.StatusInternalServerError,
		}
	})
	tests.Add("feed=normal, since=now", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("doc1", map[string]string{"foo": "bar"})
		_ = d.tDelete("doc1", kivik.Rev(rev))

		return test{
			db: d,
			options: kivik.Params(map[string]interface{}{
				"since": "now",
			}),
			wantChanges: nil,
			wantLastSeq: &[]string{"2"}[0],
			wantETag:    &[]string{"c7ba27130f956748671e845893fd6b80"}[0],
		}
	})
	tests.Add("limit=0 acts the same as limit=1", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("doc1", map[string]string{"foo": "bar"})
		_ = d.tDelete("doc1", kivik.Rev(rev))

		return test{
			db:      d,
			options: kivik.Param("limit", "0"),
			wantChanges: []driver.Change{
				{
					ID:      "doc1",
					Seq:     "1",
					Changes: driver.ChangedRevs{rev},
				},
			},
			wantLastSeq: &[]string{"1"}[0],
			wantETag:    &[]string{"9562870d7e8245d03c2ac6055dff735f"}[0],
		}
	})
	tests.Add("limit=1", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("doc1", map[string]string{"foo": "bar"})
		_ = d.tDelete("doc1", kivik.Rev(rev))

		return test{
			db:      d,
			options: kivik.Param("limit", "1"),
			wantChanges: []driver.Change{
				{
					ID:      "doc1",
					Seq:     "1",
					Changes: driver.ChangedRevs{rev},
				},
			},
			wantLastSeq: &[]string{"1"}[0],
			wantETag:    &[]string{"9562870d7e8245d03c2ac6055dff735f"}[0],
		}
	})
	tests.Add("limit=1 as int", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("doc1", map[string]string{"foo": "bar"})
		_ = d.tDelete("doc1", kivik.Rev(rev))

		return test{
			db:      d,
			options: kivik.Param("limit", 1),
			wantChanges: []driver.Change{
				{
					ID:      "doc1",
					Seq:     "1",
					Changes: driver.ChangedRevs{rev},
				},
			},
			wantLastSeq: &[]string{"1"}[0],
			wantETag:    &[]string{"9562870d7e8245d03c2ac6055dff735f"}[0],
		}
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
		ctx := tt.ctx
		if ctx == nil {
			ctx = context.Background()
		}
		opts := tt.options
		if opts == nil {
			opts = mock.NilOption
		}
		feed, err := dbc.Changes(ctx, opts)
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
				if !testy.ErrorMatches(tt.wantNextErr, err) {
					t.Errorf("Unexpected error from Next(): %s", err)
				}
				break loop
			}
			got = append(got, change)
		}

		wantChanges := tt.wantChanges
		if tt.wantChangesFn != nil {
			wantChanges = tt.wantChangesFn()
		}

		if d := cmp.Diff(wantChanges, got); d != "" {
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

func TestDBChanges_longpoll_context_cancellation_during_iteration(t *testing.T) {
	t.Parallel()
	db := newDB(t)

	// First create a single document to seed the changes feed
	_ = db.tPut("doc1", map[string]string{"foo": "bar"})

	ctx, cancel := context.WithCancel(context.Background())

	// Start the changes feed, with feed=longpoll&since=now to block until
	// another change is made.
	feed, err := db.Changes(ctx, kivik.Params(map[string]interface{}{
		"feed":  "longpoll",
		"since": "now",
	}))
	if err != nil {
		t.Fatalf("Failed to start changes feed: %s", err)
	}
	t.Cleanup(func() {
		_ = feed.Close()
	})

	// Now cancel the context
	cancel()

	var iterationErr error
	// Meanwhile, the changes feed should block until the context is cancelled
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
			iterationErr = err
			break loop
		}
	}

	if !testy.ErrorMatches("context canceled", iterationErr) {
		t.Errorf("Unexpected error from Next(): %s", iterationErr)
	}
}

func TestDBChanges_longpoll(t *testing.T) {
	t.Parallel()
	db := newDB(t)

	// First create a single document to seed the changes feed
	_ = db.tPut("doc1", map[string]string{"foo": "bar"})

	// Start the changes feed, with feed=longpoll&since=now to block until
	// another change is made.
	feed, err := db.Changes(context.Background(), kivik.Params(map[string]interface{}{
		"feed":  "longpoll",
		"since": "now",
	}))
	if err != nil {
		t.Fatalf("Failed to start changes feed: %s", err)
	}
	t.Cleanup(func() {
		_ = feed.Close()
	})

	var mu sync.Mutex
	var rev2 string
	// Make a change to the database after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		mu.Lock()
		rev2 = db.tPut("doc2", map[string]string{"foo": "bar"})
		mu.Unlock()
	}()

	start := time.Now()
	// Meanwhile, the changes feed should block until the change is made
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
			t.Fatalf("iteration failed: %s", err)
		}
		got = append(got, change)
	}

	if time.Since(start) < 100*time.Millisecond {
		t.Errorf("Changes feed returned too quickly")
	}

	mu.Lock()
	wantChanges := []driver.Change{
		{
			ID:      "doc2",
			Seq:     "2",
			Changes: driver.ChangedRevs{rev2},
		},
	}
	mu.Unlock()

	if d := cmp.Diff(wantChanges, got); d != "" {
		t.Errorf("Unexpected changes:\n%s", d)
	}
}
