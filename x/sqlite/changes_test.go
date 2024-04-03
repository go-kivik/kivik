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
	}
	tests := testy.NewTable()
	tests.Add("no changes in db", test{})
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
		}
	})

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
	})
}
