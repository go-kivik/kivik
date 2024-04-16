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

func TestDBQuery(t *testing.T) {
	t.Parallel()
	type test struct {
		db         *testDB
		ddoc, view string
		options    driver.Options
		want       []rowResult
		wantStatus int
		wantErr    string
	}
	tests := testy.NewTable()
	tests.Add("ddoc does not exist", test{
		ddoc:       "_design/foo",
		wantErr:    "missing",
		wantStatus: http.StatusNotFound,
	})
	tests.Add("ddoc does exist but view does not", func(t *testing.T) interface{} {
		d := newDB(t)
		_ = d.tPut("_design/foo", map[string]string{"cat": "meow"})

		return test{
			db:         d,
			ddoc:       "_design/foo",
			view:       "_view/bar",
			wantErr:    "missing named view",
			wantStatus: http.StatusNotFound,
		}
	})
	tests.Add("simple view with a single document", func(t *testing.T) interface{} {
		d := newDB(t)
		_ = d.tPut("_design/foo", map[string]interface{}{
			"views": map[string]interface{}{
				"bar": map[string]string{
					"map": `function(doc) { emit(doc._id, null); }`,
				},
			},
		})
		_ = d.tPut("foo", map[string]string{"_id": "foo"})

		return test{
			db:   d,
			ddoc: "_design/foo",
			view: "_view/bar",
			want: []rowResult{
				{ID: "_design/foo", Key: `"_design/foo"`, Value: "null"},
				{ID: "foo", Key: `"foo"`, Value: "null"},
			},
		}
	})

	/*
		TODO:
		- update view index before returning
		- wait for pending index update before returning
		- map function takes too long
		- expose attachment stubs to map function
		- Are conflicts or other metadata exposed to map function?
		- custom/standard CouchDB collation https://pkg.go.dev/modernc.org/sqlite#RegisterCollationUtf8
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
		rows, err := db.Query(context.Background(), tt.ddoc, tt.view, opts)
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
