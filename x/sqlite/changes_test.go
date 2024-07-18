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
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/mock"
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
		wantPending   *int64
	}
	tests := testy.NewTable()
	tests.Add("no changes in db", test{
		wantLastSeq: &[]string{""}[0],
		wantETag:    &[]string{"cfcd208495d565ef66e7dff9f98764da"}[0],
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
			wantETag:    &[]string{"c4ca4238a0b923820dcc509a6f75849b"}[0],
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
			wantETag:    &[]string{"c81e728d9d4c2f636f067f89cc14862c"}[0],
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
			wantETag:    &[]string{"c81e728d9d4c2f636f067f89cc14862c"}[0],
		}
	})
	tests.Add("malformed sequence id", test{
		options:    kivik.Param("since", "invalid"),
		wantErr:    "malformed sequence supplied in 'since' parameter: invalid",
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
			wantETag:    &[]string{"c81e728d9d4c2f636f067f89cc14862c"}[0],
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
		wantErr:    "invalid value for 'limit': invalid",
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
			wantETag:    &[]string{"cfcd208495d565ef66e7dff9f98764da"}[0],
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
			wantETag:    &[]string{"c81e728d9d4c2f636f067f89cc14862c"}[0],
			wantPending: &[]int64{1}[0],
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
			wantETag:    &[]string{"c81e728d9d4c2f636f067f89cc14862c"}[0],
			wantPending: &[]int64{1}[0],
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
			wantETag:    &[]string{"c81e728d9d4c2f636f067f89cc14862c"}[0],
			wantPending: &[]int64{1}[0],
		}
	})
	tests.Add("feed=longpoll, limit=1, pending is set", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("doc1", map[string]string{"foo": "bar"})
		_ = d.tDelete("doc1", kivik.Rev(rev))

		return test{
			db: d,
			options: kivik.Params(map[string]interface{}{
				"feed":  "longpoll",
				"limit": 1,
			}),
			wantChanges: []driver.Change{
				{
					ID:      "doc1",
					Seq:     "1",
					Changes: driver.ChangedRevs{rev},
				},
			},
			wantLastSeq: &[]string{"1"}[0],
			wantETag:    &[]string{""}[0],
			wantPending: &[]int64{1}[0],
		}
	})
	tests.Add("Descending order", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("doc1", map[string]string{"foo": "bar"})
		rev2 := d.tDelete("doc1", kivik.Rev(rev))

		return test{
			db:      d,
			options: kivik.Param("descending", true),
			wantChanges: []driver.Change{
				{
					ID:      "doc1",
					Seq:     "2",
					Deleted: true,
					Changes: driver.ChangedRevs{rev2},
				},
				{
					ID:      "doc1",
					Seq:     "1",
					Changes: driver.ChangedRevs{rev},
				},
			},
			wantLastSeq: &[]string{"1"}[0],
			wantETag:    &[]string{"c81e728d9d4c2f636f067f89cc14862c"}[0],
		}
	})
	tests.Add("include docs, normal feed", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("doc1", map[string]string{"foo": "bar"})
		rev2 := d.tDelete("doc1", kivik.Rev(rev))

		return test{
			db:      d,
			options: kivik.Param("include_docs", true),
			wantChanges: []driver.Change{
				{
					ID:      "doc1",
					Seq:     "1",
					Changes: driver.ChangedRevs{rev},
					Doc:     []byte(`{"_id":"doc1","_rev":"` + rev + `","foo":"bar"}`),
				},
				{
					ID:      "doc1",
					Seq:     "2",
					Deleted: true,
					Changes: driver.ChangedRevs{rev2},
					Doc:     []byte(`{"_id":"doc1","_rev":"` + rev2 + `","_deleted":true}`),
				},
			},
			wantLastSeq: &[]string{"2"}[0],
			wantETag:    &[]string{"c81e728d9d4c2f636f067f89cc14862c"}[0],
		}
	})
	tests.Add("include docs, attachment stubs, normal feed", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("doc1", map[string]interface{}{
			"foo": "bar",
			"_attachments": newAttachments().
				add("text.txt", "boring text").
				add("text2.txt", "more boring text"),
		})

		return test{
			db:      d,
			options: kivik.Param("include_docs", true),
			wantChanges: []driver.Change{
				{
					ID:      "doc1",
					Seq:     "1",
					Changes: driver.ChangedRevs{rev},
					Doc:     []byte(`{"_id":"doc1","_rev":"` + rev + `","foo":"bar","_attachments":{"text.txt":{"content_type":"text/plain","digest":"md5-OIJSy6hr5f32Yfxm8ex95w==","length":11,"revpos":1,"stub":true},"text2.txt":{"content_type":"text/plain","digest":"md5-JlqzqsA7DA4Lw2arCp9iXQ==","length":16,"revpos":1,"stub":true}}}`),
				},
			},
			wantLastSeq: &[]string{"1"}[0],
			wantETag:    &[]string{"c4ca4238a0b923820dcc509a6f75849b"}[0],
		}
	})
	tests.Add("include docs and attachments, normal feed", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("doc1", map[string]interface{}{
			"foo": "bar",
			"_attachments": newAttachments().
				add("text.txt", "boring text").
				add("text2.txt", "more boring text"),
		})

		return test{
			db: d,
			options: kivik.Params(map[string]interface{}{
				"include_docs": true,
				"attachments":  true,
			}),
			wantChanges: []driver.Change{
				{
					ID:      "doc1",
					Seq:     "1",
					Changes: driver.ChangedRevs{rev},
					Doc:     []byte(`{"_id":"doc1","_rev":"` + rev + `","foo":"bar","_attachments":{"text.txt":{"content_type":"text/plain","digest":"md5-OIJSy6hr5f32Yfxm8ex95w==","length":11,"revpos":1,"data":"Ym9yaW5nIHRleHQ="},"text2.txt":{"content_type":"text/plain","digest":"md5-JlqzqsA7DA4Lw2arCp9iXQ==","length":16,"revpos":1,"data":"bW9yZSBib3JpbmcgdGV4dA=="}}}`),
				},
			},
			wantLastSeq: &[]string{"1"}[0],
			wantETag:    &[]string{"c4ca4238a0b923820dcc509a6f75849b"}[0],
		}
	})

	/*
		TODO:
		- Options
			- doc_ids
			- conflicts
			- feed
				- normal
				- longpoll
				- continuous
			- filter
			- att_encoding_info
			- style
			- timeout
			- view
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
		if tt.wantPending != nil {
			got := feed.Pending()
			if got != *tt.wantPending {
				t.Errorf("Unexpected Pending: %d", got)
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
		rev, err := db.Put(context.Background(), "doc2", interface{}(map[string]string{"foo": "bar"}), mock.NilOption)
		if err != nil {
			panic(fmt.Sprintf("Failed to put doc: %s", err))
		}
		mu.Lock()
		rev2 = rev
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

func TestDBChanges_longpoll_include_docs(t *testing.T) {
	t.Parallel()
	db := newDB(t)

	// First create a single document to seed the changes feed
	rev := db.tPut("doc1", map[string]string{"foo": "bar"})

	// Start the changes feed, with feed=longpoll&since=now to block until
	// another change is made.
	feed, err := db.Changes(context.Background(), kivik.Params(map[string]interface{}{
		"feed":         "longpoll",
		"since":        "now",
		"include_docs": true,
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
		rev2 = db.tDelete("doc1", kivik.Rev(rev))
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
			ID:      "doc1",
			Seq:     "2",
			Deleted: true,
			Changes: driver.ChangedRevs{rev2},
			Doc:     []byte(`{"_id":"doc1","_rev":"` + rev2 + `","_deleted":true}`),
		},
	}
	mu.Unlock()

	if d := cmp.Diff(wantChanges, got); d != "" {
		t.Errorf("Unexpected changes:\n%s", d)
	}
}

func TestDBChanges_longpoll_include_docs_and_attachments(t *testing.T) {
	t.Parallel()
	db := newDB(t)

	// First create a single document to seed the changes feed
	rev := db.tPut("doc1", map[string]string{"foo": "bar"})

	// Start the changes feed, with feed=longpoll&since=now to block until
	// another change is made.
	feed, err := db.Changes(context.Background(), kivik.Params(map[string]interface{}{
		"feed":         "longpoll",
		"attachments":  true,
		"since":        "now",
		"include_docs": true,
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
		rev2 = db.tPut("doc1", map[string]interface{}{
			"_attachments": newAttachments().
				add("text.txt", "boring text").
				add("text2.txt", "more boring text"),
		}, kivik.Rev(rev))
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
			ID:      "doc1",
			Seq:     "2",
			Changes: driver.ChangedRevs{rev2},
			Doc:     []byte(`{"_id":"doc1","_rev":"` + rev2 + `","_attachments":{"text.txt":{"content_type":"text/plain","digest":"md5-OIJSy6hr5f32Yfxm8ex95w==","length":11,"revpos":2,"data":"Ym9yaW5nIHRleHQ="},"text2.txt":{"content_type":"text/plain","digest":"md5-JlqzqsA7DA4Lw2arCp9iXQ==","length":16,"revpos":2,"data":"bW9yZSBib3JpbmcgdGV4dA=="}}}`),
		},
	}
	mu.Unlock()

	if d := cmp.Diff(wantChanges, got); d != "" {
		t.Errorf("Unexpected changes:\n%s", d)
	}
}

func TestDBChanges_longpoll_include_docs_with_attachment_stubs(t *testing.T) {
	t.Parallel()
	db := newDB(t)

	// First create a single document to seed the changes feed
	rev := db.tPut("doc1", map[string]string{"foo": "bar"})

	// Start the changes feed, with feed=longpoll&since=now to block until
	// another change is made.
	feed, err := db.Changes(context.Background(), kivik.Params(map[string]interface{}{
		"feed":         "longpoll",
		"since":        "now",
		"include_docs": true,
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
		rev2 = db.tPut("doc1", map[string]interface{}{
			"_attachments": newAttachments().
				add("text.txt", "boring text").
				add("text2.txt", "more boring text"),
		}, kivik.Rev(rev))
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
			ID:      "doc1",
			Seq:     "2",
			Changes: driver.ChangedRevs{rev2},
			Doc:     []byte(`{"_id":"doc1","_rev":"` + rev2 + `","_attachments":{"text.txt":{"content_type":"text/plain","digest":"md5-OIJSy6hr5f32Yfxm8ex95w==","length":11,"revpos":2,"stub":true},"text2.txt":{"content_type":"text/plain","digest":"md5-JlqzqsA7DA4Lw2arCp9iXQ==","length":16,"revpos":2,"stub":true}}}`),
		},
	}
	mu.Unlock()

	if d := cmp.Diff(wantChanges, got); d != "" {
		t.Errorf("Unexpected changes:\n%s", d)
	}
}

// This test validates that the query for the normal changes feed does not
// duplicate unnecessary fields when returning multiple attachments.
func Test_normal_changes_query(t *testing.T) {
	t.Parallel()

	filename1, filename2 := "text.txt", "text2.txt"

	d := newDB(t)
	rev := d.tPut("doc1", map[string]interface{}{
		"_attachments": newAttachments().
			add(filename1, "boring text").
			add(filename2, "more boring text"),
	})

	changes, err := d.DB.(*db).newNormalChanges(context.Background(), optsMap{"include_docs": true}, nil, nil, false, "normal")
	if err != nil {
		t.Fatal(err)
	}

	defer changes.rows.Close()

	type row struct {
		ID              *string
		Seq             *string
		Deleted         *bool
		Rev             *string
		Doc             *string
		AttachmentCount int
		Filename        *string
	}
	var got []row
	for changes.rows.Next() {
		var result row
		if err := changes.rows.Scan(
			&result.ID, &result.Seq, &result.Deleted, &result.Rev, &result.Doc,
			&result.AttachmentCount, &result.Filename, discard{}, discard{}, discard{}, discard{}, discard{},
		); err != nil {
			t.Fatal(err)
		}
		got = append(got, result)
	}
	if err := changes.rows.Err(); err != nil {
		t.Fatal(err)
	}

	want := []row{
		{ID: &[]string{"doc1"}[0], Seq: &[]string{"1"}[0], Deleted: &[]bool{false}[0], Rev: &rev, Doc: &[]string{"{}"}[0], AttachmentCount: 2, Filename: &filename1},
		{AttachmentCount: 2, Filename: &filename2},
	}

	if d := cmp.Diff(got, want); d != "" {
		t.Errorf("Unexpected rows:\n%s", d)
	}
}

// This test validates that the query for the normal changes feed does not
// include attachments fields when include_docs=false.
func Test_normal_changes_query_without_docs(t *testing.T) {
	t.Parallel()

	filename1, filename2 := "text.txt", "text2.txt"

	d := newDB(t)
	rev := d.tPut("doc1", map[string]interface{}{
		"_attachments": newAttachments().
			add(filename1, "boring text").
			add(filename2, "more boring text"),
	})

	changes, err := d.DB.(*db).newNormalChanges(context.Background(), nil, nil, nil, false, "normal")
	if err != nil {
		t.Fatal(err)
	}

	defer changes.rows.Close()

	type row struct {
		ID              *string
		Seq             *string
		Deleted         *bool
		Rev             *string
		Doc             *string
		AttachmentCount int
		Filename        *string
	}
	var got []row
	for changes.rows.Next() {
		var result row
		if err := changes.rows.Scan(
			&result.ID, &result.Seq, &result.Deleted, &result.Rev, &result.Doc,
			&result.AttachmentCount, &result.Filename, discard{}, discard{}, discard{}, discard{}, discard{},
		); err != nil {
			t.Fatal(err)
		}
		got = append(got, result)
	}
	if err := changes.rows.Err(); err != nil {
		t.Fatal(err)
	}

	want := []row{
		{ID: &[]string{"doc1"}[0], Seq: &[]string{"1"}[0], Deleted: &[]bool{false}[0], Rev: &rev},
	}

	if d := cmp.Diff(got, want); d != "" {
		t.Errorf("Unexpected rows:\n%s", d)
	}
}

// This test validates that the query for the longpolll changes feed does not
// duplicate unnecessary fields when returning multiple attachments.
func Test_longpoll_changes_query(t *testing.T) {
	t.Parallel()

	filename1, filename2 := "text.txt", "text2.txt"

	d := newDB(t)

	changes, err := d.DB.(*db).newLongpollChanges(context.Background(), true, false)
	if err != nil {
		t.Fatal(err)
	}

	// Create a change
	rev := d.tPut("doc1", map[string]interface{}{
		"_attachments": newAttachments().
			add(filename1, "boring text").
			add(filename2, "more boring text"),
	})

	// Then execute the prepared statement
	rows, err := changes.stmt.Query(0, true)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	type row struct {
		ID       *string
		Seq      *string
		Deleted  *bool
		Rev      *string
		Doc      *string
		Filename *string
	}
	var got []row
	for rows.Next() {
		var result row
		if err := rows.Scan(
			&result.ID, &result.Seq, &result.Deleted, &result.Rev, &result.Doc,
			&result.Filename, discard{}, discard{}, discard{}, discard{}, discard{},
		); err != nil {
			t.Fatal(err)
		}
		got = append(got, result)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	want := []row{
		{ID: &[]string{"doc1"}[0], Seq: &[]string{"1"}[0], Deleted: &[]bool{false}[0], Rev: &rev, Doc: &[]string{"{}"}[0], Filename: &filename1},
		{Filename: &filename2},
	}

	if d := cmp.Diff(got, want); d != "" {
		t.Errorf("Unexpected rows:\n%s", d)
	}
}

// This test validates that the query for the longpolll changes feed does not
// include any attachment data when include_docs=false
func Test_longpoll_changes_query_without_docs(t *testing.T) {
	t.Parallel()

	filename1, filename2 := "text.txt", "text2.txt"

	d := newDB(t)

	changes, err := d.DB.(*db).newLongpollChanges(context.Background(), false, false)
	if err != nil {
		t.Fatal(err)
	}

	// Create a change
	rev := d.tPut("doc1", map[string]interface{}{
		"_attachments": newAttachments().
			add(filename1, "boring text").
			add(filename2, "more boring text"),
	})

	// Then execute the prepared statement
	rows, err := changes.stmt.Query(0, true)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	type row struct {
		ID       *string
		Seq      *string
		Deleted  *bool
		Rev      *string
		Doc      *string
		Filename *string
	}
	var got []row
	for rows.Next() {
		var result row
		if err := rows.Scan(
			&result.ID, &result.Seq, &result.Deleted, &result.Rev, &result.Doc,
			&result.Filename, discard{}, discard{}, discard{}, discard{}, discard{},
		); err != nil {
			t.Fatal(err)
		}
		got = append(got, result)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	want := []row{
		{ID: &[]string{"doc1"}[0], Seq: &[]string{"1"}[0], Deleted: &[]bool{false}[0], Rev: &rev},
	}

	if d := cmp.Diff(got, want); d != "" {
		t.Errorf("Unexpected rows:\n%s", d)
	}
}
