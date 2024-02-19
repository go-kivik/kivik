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

func TestDBDelete(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		setup      func(*testing.T, driver.DB)
		id         string
		options    driver.Options
		wantRev    string
		check      func(*testing.T, driver.DB)
		wantStatus int
		wantErr    string
	}{
		{
			name:       "not found",
			id:         "foo",
			wantStatus: http.StatusNotFound,
			wantErr:    "not found",
		},
		{
			name: "success",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
				if err != nil {
					t.Fatal(err)
				}
			},
			id:      "foo",
			options: kivik.Rev("1-9bb58f26192e4ba00f01e2e7b136bbd8"),
			wantRev: "2-df2a4fe30cde39c357c8d1105748d1b9",
			check: func(t *testing.T, d driver.DB) {
				var deleted bool
				err := d.(*db).db.QueryRow(`
					SELECT deleted
					FROM test
					WHERE id='foo'
					ORDER BY rev DESC, rev_id DESC
					LIMIT 1
				`).Scan(&deleted)
				if err != nil {
					t.Fatal(err)
				}
				if !deleted {
					t.Errorf("Document not marked deleted")
				}
			},
		},
		{
			name: "replay delete should conflict",
			setup: func(t *testing.T, d driver.DB) {
				rev, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Delete(context.Background(), "foo", kivik.Rev(rev))
				if err != nil {
					t.Fatal(err)
				}
			},
			id:         "foo",
			options:    kivik.Rev("1-9bb58f26192e4ba00f01e2e7b136bbd8"),
			wantStatus: http.StatusConflict,
			wantErr:    "conflict",
		},
		{
			name: "delete deleted doc should succeed",
			setup: func(t *testing.T, d driver.DB) {
				rev, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
				if err != nil {
					t.Fatal(err)
				}
				_, err = d.Delete(context.Background(), "foo", kivik.Rev(rev))
				if err != nil {
					t.Fatal(err)
				}
			},
			id:      "foo",
			options: kivik.Rev("2-df2a4fe30cde39c357c8d1105748d1b9"),
			wantRev: "3-df2a4fe30cde39c357c8d1105748d1b9",
		},
		{
			name: "delete without rev",
			setup: func(t *testing.T, d driver.DB) {
				_, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
				if err != nil {
					t.Fatal(err)
				}
			},
			id:         "foo",
			wantStatus: http.StatusConflict,
			wantErr:    "conflict",
		},
		{
			name: "delete losing rev for conflict",
			setup: func(t *testing.T, db driver.DB) {
				_, err := db.Put(context.Background(), "foo", map[string]string{
					"cat":  "meow",
					"_rev": "1-xxx",
				}, kivik.Param("new_edits", false))
				if err != nil {
					t.Fatal(err)
				}
				_, err = db.Put(context.Background(), "foo", map[string]string{
					"cat":  "purr",
					"_rev": "1-aaa",
				}, kivik.Param("new_edits", false))
				if err != nil {
					t.Fatal(err)
				}
			},
			id:      "foo",
			options: kivik.Rev("1-aaa"),
			wantRev: "2-xxxxx",
		},
		/*
			- _revisions
			- deleting losing rev should not conflict
		*/
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := newDB(t)
			if tt.setup != nil {
				tt.setup(t, db)
			}
			opts := tt.options
			if opts == nil {
				opts = mock.NilOption
			}
			rev, err := db.Delete(context.Background(), tt.id, opts)
			if !testy.ErrorMatches(tt.wantErr, err) {
				t.Errorf("Unexpected error: %s", err)
			}
			if status := kivik.HTTPStatus(err); status != tt.wantStatus {
				t.Errorf("Unexpected status: %d", status)
			}
			if err != nil {
				return
			}
			if rev != tt.wantRev {
				t.Errorf("Unexpected rev: %s", rev)
			}
			if tt.check != nil {
				tt.check(t, db)
			}
		})
	}
}
