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

func TestDBOpenRevs(t *testing.T) {
	t.Parallel()
	type test struct {
		db         *testDB
		docID      string
		revs       []string
		options    driver.Options
		want       []rowResult
		wantErr    string
		wantStatus int
	}
	tests := testy.NewTable()
	tests.Add("all revs, document not found", test{
		docID:      "not there",
		revs:       []string{"all"},
		wantErr:    "missing",
		wantStatus: http.StatusNotFound,
	})
	tests.Add("invalid rev format", func(t *testing.T) interface{} {
		d := newDB(t)
		docID := "foo"
		_ = d.tPut(docID, map[string]string{"foo": "bar"})

		return test{
			db:         d,
			docID:      docID,
			revs:       []string{"oink", "all"},
			wantErr:    "invalid rev format",
			wantStatus: http.StatusBadRequest,
		}
	})
	/*
		TODO:
		- No revs provided == returns winning leaf
		- document not found, open_revs=["something"] = 200 + missing
		- document found, rev not found
		- all revs
		- latest=true

		Do other GET options have any effect?
		- attachments
		- att_encoding_info
		- atts_since
		- conflicts
		- deleted_conflicts
		- local_seq
		- meta
		- rev
		- revs
		- revs_info
		- If-None-Match header
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

		rows, err := db.OpenRevs(context.Background(), tt.docID, tt.revs, opts)
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
