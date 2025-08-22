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
	"github.com/go-kivik/kivik/v4/driver"
)

func TestCreateDB(t *testing.T) {
	t.Parallel()

	type test struct {
		client     *client
		dbName     string
		options    driver.Options
		wantErr    string
		wantStatus int
	}

	tests := testy.NewTable()
	tests.Add("invalid db name", test{
		// This test violates the documented database name requirements
		// found at https://docs.couchdb.org/en/stable/api/database/common.html#put--db
		client:     &client{},
		dbName:     "Capitalized",
		wantErr:    "invalid database name",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("db connection error", func(t *testing.T) any {
		client := testClient(t)

		client.pool.Close()

		return test{
			client:     client,
			dbName:     "testdb",
			wantErr:    "closed pool",
			wantStatus: http.StatusInternalServerError,
		}
	})
	tests.Add("create db", test{
		client: testClient(t),
		dbName: "testdb",
	})

	/*
		TODO:
		- db already exists
	*/

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()

		err := tt.client.CreateDB(t.Context(), tt.dbName, tt.options)
		if !testy.ErrorMatchesRE(tt.wantErr, err) {
			t.Errorf("CreateDB(%s, %s) error = %s, want %s", tt.dbName, tt.options, err, tt.wantErr)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("CreateDB(%s, %s) status = %d, want %d", tt.dbName, tt.options, status, tt.wantStatus)
		}
		if err != nil {
			return
		}
		// Verify the database was created
		var found bool
		err = tt.client.pool.QueryRow(t.Context(), "SELECT true FROM pg_tables WHERE tablename = $1", tt.dbName).Scan(&found)
		if err != nil {
			t.Errorf("Failed to verify database creation: %s", err)
			return
		}
		if !found {
			t.Errorf("Database %s was not created", tt.dbName)
			return
		}
	})
}
