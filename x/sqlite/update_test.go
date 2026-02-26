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
	"regexp"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestDBUpdate(t *testing.T) {
	t.Parallel()
	type test struct {
		db         *testDB
		ddoc       string
		funcName   string
		docID      string
		doc        any
		wantRev    string
		wantStatus int
		wantErr    string
	}

	tests := testy.NewTable()
	tests.Add("update function not found", test{
		ddoc:       "_design/myddoc",
		funcName:   "myfunc",
		docID:      "foo",
		doc:        map[string]any{},
		wantStatus: http.StatusNotFound,
		wantErr:    "missing",
	})

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		dbc := tt.db
		if dbc == nil {
			dbc = newDB(t)
		}
		updater := dbc.DB.(driver.Updater)
		rev, err := updater.Update(context.Background(), tt.ddoc, tt.funcName, tt.docID, tt.doc, mock.NilOption)
		if !testy.ErrorMatchesRE(tt.wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status: %d", status)
		}
		if err != nil {
			return
		}
		if !regexp.MustCompile(tt.wantRev).MatchString(rev) {
			t.Errorf("Unexpected rev: %s, want /%s/", rev, tt.wantRev)
		}
	})
}
