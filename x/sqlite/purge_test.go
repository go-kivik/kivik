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

package sqlite

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

func TestDBPurge(t *testing.T) {
	t.Parallel()
	type test struct {
		db         *testDB
		arg        map[string][]string
		want       *driver.PurgeResult
		wantErr    string
		wantStatus int
		wantRevs   []leaf
	}
	tests := testy.NewTable()
	tests.Add("nothing to purge", test{
		arg: map[string][]string{
			"foo": {"1-1234", "2-5678"},
		},
		want: &driver.PurgeResult{
			Seq: 1,
		},
	})
	tests.Add("success", func(t *testing.T) any {
		d := newDB(t)
		rev := d.tPut("foo", map[string]string{"foo": "bar"})

		return test{
			db: d,
			arg: map[string][]string{
				"foo": {rev},
			},
			want: &driver.PurgeResult{
				Seq: 1,
				Purged: map[string][]string{
					"foo": {rev},
				},
			},
			wantRevs: nil,
		}
	})
	tests.Add("malformed rev", test{
		arg: map[string][]string{
			"foo": {"abc"},
		},
		wantErr:    "invalid rev format",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("attempt to purge non-leaf rev does nothing", func(t *testing.T) any {
		d := newDB(t)
		rev := d.tPut("foo", map[string]string{"foo": "bar"})
		_ = d.tPut("foo", map[string]string{"foo": "baz"}, kivik.Rev(rev))

		return test{
			db: d,
			arg: map[string][]string{
				"foo": {rev},
			},
			want: &driver.PurgeResult{
				Seq: 1,
			},
			wantRevs: []leaf{
				{ID: "foo", Rev: 1},
				{ID: "foo", Rev: 2, ParentRev: &[]int{1}[0]},
			},
		}
	})
	tests.Add("deleting conflict leaves non-conflicting leaf", func(t *testing.T) any {
		d := newDB(t)
		_ = d.tPut("foo", map[string]any{
			"version": "one",
			"_revisions": map[string]any{
				"start": 3,
				"ids":   []string{"ccc", "bbb", "aaa"},
			},
		}, kivik.Param("new_edits", false))
		_ = d.tPut("foo", map[string]any{
			"version": "two",
			"_revisions": map[string]any{
				"start": 3,
				"ids":   []string{"rrr", "qqq", "aaa"},
			},
		}, kivik.Param("new_edits", false))

		return test{
			db: d,
			arg: map[string][]string{
				"foo": {"3-ccc"},
			},
			want: &driver.PurgeResult{
				Seq: 1,
				Purged: map[string][]string{
					"foo": {"3-ccc"},
				},
			},
			wantRevs: []leaf{
				{ID: "foo", Rev: 1, RevID: "aaa"},
				{ID: "foo", Rev: 2, RevID: "bbb", ParentRev: &[]int{1}[0], ParentRevID: &[]string{"aaa"}[0]},
				{ID: "foo", Rev: 2, RevID: "qqq", ParentRev: &[]int{1}[0], ParentRevID: &[]string{"aaa"}[0]},
				{ID: "foo", Rev: 3, RevID: "rrr", ParentRev: &[]int{2}[0], ParentRevID: &[]string{"qqq"}[0]},
			},
		}
	})

	/*
		TODO:
		- What happens when purging a leaf, and its parent at the same time?
		- refactor: bulk delete, bulk lookup
	*/

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		db := tt.db
		if db == nil {
			db = newDB(t)
		}
		got, err := db.Purge(context.Background(), tt.arg)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status: %d", status)
		}
		if err != nil {
			return
		}
		if d := cmp.Diff(got, tt.want); d != "" {
			t.Errorf("Unexpected result:\n%s", d)
		}
		leaves := readRevisions(t, db.underlying())
		for i, r := range tt.wantRevs {
			// allow tests to omit RevID
			if r.RevID == "" {
				leaves[i].RevID = ""
			}
			if r.ParentRevID == nil {
				leaves[i].ParentRevID = nil
			}
		}
		if d := cmp.Diff(tt.wantRevs, leaves); d != "" {
			t.Errorf("Unexpected leaves: %s", d)
		}
	})
}

func TestDBPurge_seq(t *testing.T) {
	t.Parallel()

	type purgeStep struct {
		arg     map[string][]string
		wantSeq int64
	}

	type test struct {
		db    *testDB
		steps []purgeStep
	}

	tests := testy.NewTable()

	tests.Add("sequential purges increment seq", func(t *testing.T) any {
		d := newDB(t)
		rev1 := d.tPut("doc1", map[string]string{"key": "val1"})
		rev2 := d.tPut("doc2", map[string]string{"key": "val2"})

		return test{
			db: d,
			steps: []purgeStep{
				{
					arg:     map[string][]string{"doc1": {rev1}},
					wantSeq: 1,
				},
				{
					arg:     map[string][]string{"doc2": {rev2}},
					wantSeq: 2,
				},
			},
		}
	})

	tests.Add("empty purge still increments seq", func(t *testing.T) any {
		d := newDB(t)
		rev := d.tPut("doc1", map[string]string{"key": "val"})

		return test{
			db: d,
			steps: []purgeStep{
				{
					arg:     map[string][]string{"nonexistent": {"1-abc"}},
					wantSeq: 1,
				},
				{
					arg:     map[string][]string{"doc1": {rev}},
					wantSeq: 2,
				},
			},
		}
	})

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		for i, step := range tt.steps {
			got, err := tt.db.Purge(context.Background(), step.arg)
			if err != nil {
				t.Fatalf("step %d: unexpected error: %s", i, err)
			}
			if got.Seq != step.wantSeq {
				t.Errorf("step %d: unexpected Seq: got %d, want %d", i, got.Seq, step.wantSeq)
			}
		}
	})
}
