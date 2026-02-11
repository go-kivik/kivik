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
	"encoding/json"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestDeleteIndex(t *testing.T) {
	t.Parallel()
	type test struct {
		db      *testDB
		ddoc    string
		name    string
		wantErr string
		check   func()
	}

	tests := testy.NewTable()
	tests.Add("drops the real SQLite index", func(t *testing.T) any {
		db := newDB(t)
		err := db.CreateIndex(context.Background(), "_design/my-index", "my-index", json.RawMessage(`{"fields":["name"]}`), mock.NilOption)
		if err != nil {
			t.Fatalf("creating index: %s", err)
		}

		return test{
			db:   db,
			ddoc: "_design/my-index",
			name: "my-index",
			check: func() {
				rows, err := db.underlying().Query(
					`SELECT name FROM sqlite_master WHERE type='index' AND name LIKE 'idx_kivik$test$mango_%'`,
				)
				if err != nil {
					t.Fatalf("querying sqlite_master: %s", err)
				}
				defer rows.Close()

				var names []string
				for rows.Next() {
					var name string
					if err := rows.Scan(&name); err != nil {
						t.Fatalf("scanning index name: %s", err)
					}
					names = append(names, name)
				}
				if err := rows.Err(); err != nil {
					t.Fatalf("iterating rows: %s", err)
				}
				if len(names) != 0 {
					t.Errorf("expected 0 SQLite indexes matching idx_kivik$test$mango_*, got %d: %v", len(names), names)
				}
			},
		}
	})

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		db := tt.db
		if db == nil {
			db = newDB(t)
		}
		err := db.DeleteIndex(context.Background(), tt.ddoc, tt.name, mock.NilOption)
		if !testy.ErrorMatchesRE(tt.wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if tt.check != nil {
			tt.check()
		}
	})
}
