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
)

func TestDBRevsDiff(t *testing.T) {
	t.Parallel()
	type test struct {
		db         *testDB
		revMap     interface{}
		want       []rowResult
		wantErr    string
		wantStatus int
	}
	tests := testy.NewTable()
	tests.Add("invalid revMap", test{
		revMap:     "foo",
		wantErr:    "invalid body",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("empty revMap", test{
		revMap: map[string][]string{},
		want:   nil,
	})
	tests.Add("all missing", test{
		revMap: map[string][]string{
			"foo": {"1-abc", "2-def"},
			"bar": {"3-ghi"},
		},
		want: []rowResult{
			{ID: "bar", Value: `{"missing":["3-ghi"]}`},
			{ID: "foo", Value: `{"missing":["1-abc","2-def"]}`},
		},
	})

	/*
		TODO:
		- populate `possible_ancestors`
	*/

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		db := tt.db
		if db == nil {
			db = newDB(t)
		}
		rows, err := db.RevsDiff(context.Background(), tt.revMap)
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
