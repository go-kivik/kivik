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

type rowResult struct {
	ID    string
	Rev   string
	Value string
	Error string
}

func TestDBAllDocs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		setup      func(*testing.T, driver.DB)
		options    driver.Options
		want       []rowResult
		wantStatus int
		wantErr    string
	}{
		{
			name: "no docs in db",
			want: nil,
		},
		{
			name: "single doc",
			setup: func(t *testing.T, db driver.DB) {
				_, err := db.Put(context.Background(), "foo", map[string]string{"cat": "meow"}, mock.NilOption)
				if err != nil {
					t.Fatal(err)
				}
			},
			want: []rowResult{
				{
					ID:    "foo",
					Rev:   "1-274558516009acbe973682d27a58b598",
					Value: `{"value":{"rev":"1-274558516009acbe973682d27a58b598"}}` + "\n",
				},
			},
		},
		/*
			TODO:
			- AllDocs() called for DB that doesn't exit
			- UpdateSeq() called on rows
			- Offset() called on rows
			- TotalRows() called on rows
		*/
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := newDB(t)
			opts := tt.options
			if tt.setup != nil {
				tt.setup(t, db)
			}
			if opts == nil {
				opts = mock.NilOption
			}
			rows, err := db.AllDocs(context.Background(), opts)
			if !testy.ErrorMatches(tt.wantErr, err) {
				t.Errorf("Unexpected error: %s", err)
			}
			if status := kivik.HTTPStatus(err); status != tt.wantStatus {
				t.Errorf("Unexpected status: %d", status)
			}
			if err != nil {
				return
			}
			// iterate over rows
			var got []rowResult

		loop:
			for {
				row := driver.Row{}
				err := rows.Next(&row)
				switch err {
				case io.EOF:
					break loop
				case driver.EOQ:
					continue
				case nil:
					// continue
				default:
					t.Fatalf("Next() returned error: %s", err)
				}
				var errMsg string
				if row.Error != nil {
					errMsg = row.Error.Error()
				}
				value, err := io.ReadAll(row.Value)
				if err != nil {
					t.Fatal(err)
				}
				got = append(got, rowResult{
					ID:    row.ID,
					Rev:   row.Rev,
					Value: string(value),
					Error: errMsg,
				})
			}
			if d := cmp.Diff(tt.want, got); d != "" {
				t.Errorf("Unexpected rows:\n%s", d)
			}
		})
	}
}
