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

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestGetRev(t *testing.T) {
	t.Parallel()
	type test struct {
		db         *testDB
		id         string
		options    driver.Options
		want       string
		wantStatus int
		wantErr    string
	}
	tests := testy.NewTable()
	tests.Add("not found", test{
		id:         "foo",
		wantStatus: http.StatusNotFound,
		wantErr:    "not found",
	})
	tests.Add("success", func(t *testing.T) interface{} {
		db := newDB(t)
		rev := db.tPut("foo", map[string]string{"foo": "bar"})

		return test{
			db:   db,
			id:   "foo",
			want: rev,
		}
	})
	tests.Add("get specific rev", func(t *testing.T) interface{} {
		db := newDB(t)
		rev := db.tPut("foo", map[string]string{"foo": "bar"})
		_ = db.tPut("foo", map[string]string{"foo": "baz"}, kivik.Rev(rev))

		return test{
			db:      db,
			id:      "foo",
			options: kivik.Rev(rev),
			want:    rev,
		}
	})
	tests.Add("specific rev not found", test{
		id:         "foo",
		options:    kivik.Rev("1-9bb58f26192e4ba00f01e2e7b136bbd8"),
		wantStatus: http.StatusNotFound,
		wantErr:    "not found",
	})
	tests.Add("deleted document", func(t *testing.T) interface{} {
		db := newDB(t)
		rev := db.tPut("foo", map[string]string{"foo": "bar"})
		_ = db.tDelete("foo", kivik.Rev(rev))

		return test{
			db:         db,
			id:         "foo",
			wantStatus: http.StatusNotFound,
			wantErr:    "not found",
		}
	})
	tests.Add("deleted document by rev", func(t *testing.T) interface{} {
		db := newDB(t)
		rev := db.tPut("foo", map[string]string{"foo": "bar"})
		rev = db.tDelete("foo", kivik.Rev(rev))

		return test{
			db:      db,
			id:      "foo",
			options: kivik.Rev(rev),
			want:    rev,
		}
	})
	tests.Add("get latest winning leaf", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]interface{}{"foo": "aaa", "_rev": "1-aaa"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))
		_ = db.tPut("foo", map[string]interface{}{
			"foo": "bbb",
			"_revisions": map[string]interface{}{
				"ids":   []string{"bbb", "aaa"},
				"start": 2,
			},
		}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))
		_ = db.tPut("foo", map[string]interface{}{
			"foo": "ddd",
			"_revisions": map[string]interface{}{
				"ids":   []string{"yyy", "aaa"},
				"start": 2,
			},
		}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))

		return test{
			db: db,
			id: "foo",
			options: kivik.Params(map[string]interface{}{
				"latest": true,
				"rev":    "1-aaa",
			}),
			want: "2-yyy",
		}
	})
	tests.Add("get latest non-winning leaf", func(t *testing.T) interface{} {
		db := newDB(t)
		// common root doc
		_ = db.tPut("foo", map[string]interface{}{"foo": "aaa", "_rev": "1-aaa"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))
		// losing branch
		_ = db.tPut("foo", map[string]interface{}{
			"foo": "bbb",
			"_revisions": map[string]interface{}{
				"ids":   []string{"ccc", "bbb", "aaa"},
				"start": 3,
			},
		}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))

		// winning branch
		_ = db.tPut("foo", map[string]interface{}{
			"foo": "ddd",
			"_revisions": map[string]interface{}{
				"ids":   []string{"xxx", "yyy", "aaa"},
				"start": 3,
			},
		}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))

		return test{
			db: db,
			id: "foo",
			options: kivik.Params(map[string]interface{}{
				"latest": true,
				"rev":    "2-bbb",
			}),
			want: "3-ccc",
		}
	})
	tests.Add("get latest rev with deleted leaf, reverts to the winning branch", func(t *testing.T) interface{} {
		db := newDB(t)
		// common root doc
		_ = db.tPut("foo", map[string]interface{}{"foo": "aaa", "_rev": "1-aaa"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))
		// losing branch
		_ = db.tPut("foo", map[string]interface{}{
			"foo": "bbb",
			"_revisions": map[string]interface{}{
				"ids":   []string{"ccc", "bbb", "aaa"},
				"start": 3,
			},
		}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))
		// now delete the losing leaf
		_ = db.tDelete("foo", kivik.Rev("3-ccc"))

		// winning branch
		_ = db.tPut("foo", map[string]interface{}{
			"foo": "ddd",
			"_revisions": map[string]interface{}{
				"ids":   []string{"xxx", "yyy", "aaa"},
				"start": 3,
			},
		}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))

		return test{
			db: db,
			id: "foo",
			options: kivik.Params(map[string]interface{}{
				"latest": true,
				"rev":    "2-bbb",
			}),
			want: "3-xxx",
		}
	})

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
		rev, err := db.GetRev(context.Background(), tt.id, opts)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status: %d", status)
		}
		if err != nil {
			return
		}
		if rev != tt.want {
			t.Errorf("Unexpected rev: %s", rev)
		}
	})
}
