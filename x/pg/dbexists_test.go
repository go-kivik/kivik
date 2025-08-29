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

package pg

import (
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
)

func TestDBExists(t *testing.T) {
	t.Parallel()

	type test struct {
		client     *client
		dbName     string
		want       bool
		wantErr    string
		wantStatus int
	}

	tests := testy.NewTable()
	tests.Add("invalid db name", test{
		client:     &client{},
		dbName:     "Capitalized",
		wantErr:    `database "Capitalized" not found`,
		wantStatus: http.StatusNotFound,
	})

	/*
		TODO:
		- test for connection error
		- test for existing db
		- test for non-existing db
	*/

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()

		got, err := tt.client.DBExists(t.Context(), tt.dbName, nil)
		if !testy.ErrorMatchesRE(tt.wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status: %d", status)
		}
		if got != tt.want {
			t.Errorf("Unexpected result: %v", got)
		}
	})
}
