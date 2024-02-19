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
				got = append(got, rowResult{
					ID:    row.ID,
					Rev:   row.Rev,
					Error: errMsg,
				})
			}
			if d := cmp.Diff(tt.want, got); d != "" {
				t.Errorf("Unexpected rows:\n%s", d)
			}
		})
	}
}
