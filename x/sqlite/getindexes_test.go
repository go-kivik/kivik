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

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestGetIndexes(t *testing.T) {
	t.Parallel()
	type test struct {
		db      *testDB
		want    []driver.Index
		wantErr string
	}

	tests := testy.NewTable()
	tests.Add("returns all_docs and user-created index", func(t *testing.T) interface{} {
		db := newDB(t)
		err := db.CreateIndex(context.Background(), "_design/my-index", "my-index", json.RawMessage(`{"fields":["name"]}`), mock.NilOption)
		if err != nil {
			t.Fatalf("CreateIndex failed: %s", err)
		}
		return test{
			db: db,
			want: []driver.Index{
				{
					DesignDoc:  "",
					Name:       "_all_docs",
					Type:       "special",
					Definition: map[string]interface{}{"fields": []map[string]string{{"_id": "asc"}}},
				},
				{
					DesignDoc:  "_design/my-index",
					Name:       "my-index",
					Type:       "json",
					Definition: map[string]interface{}{"fields": []map[string]string{{"name": "asc"}}},
				},
			},
		}
	})

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		db := tt.db
		if db == nil {
			db = newDB(t)
		}
		got, err := db.GetIndexes(context.Background(), mock.NilOption)
		if !testy.ErrorMatchesRE(tt.wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if d := cmp.Diff(tt.want, got); d != "" {
			t.Errorf("Unexpected indexes:\n%s", d)
		}
	})
}
