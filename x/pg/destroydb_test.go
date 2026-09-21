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

	"gitlab.com/flimzy/testy/v2"
)

func TestDestroyDB(t *testing.T) {
	t.Parallel()

	type test struct {
		client     *client
		dbName     string
		wantErr    string
		wantStatus int
	}

	tests := testy.NewTable[test]()
	tests.Add("invalid db name", test{
		client:     &client{},
		dbName:     "Capitalized",
		wantErr:    "invalid database name",
		wantStatus: http.StatusBadRequest,
	})
	tests.AddFunc("db connection error", func(t *testing.T) test {
		client := testClient(t)

		client.pool.Close()

		return test{
			client:     client,
			dbName:     "testdb",
			wantErr:    "closed pool",
			wantStatus: http.StatusInternalServerError,
		}
	})
	tests.AddFunc("destroy db succeeds", func(t *testing.T) test {
		client := testClient(t)

		const dbName = "deleteme"

		if err := client.CreateDB(t.Context(), dbName, nil); err != nil {
			t.Fatal(err)
		}

		return test{
			client: client,
			dbName: dbName,
		}
	})
	tests.AddFunc("delete db that does not exist", func(t *testing.T) test {
		client := testClient(t)

		return test{
			client:     client,
			dbName:     "notfound",
			wantErr:    "not found",
			wantStatus: http.StatusNotFound,
		}
	})

	tests.Parallel()
	tests.Run(t, func(t *testing.T, tt test) {
		err := tt.client.DestroyDB(t.Context(), tt.dbName, nil)
		if !testy.ErrorMatchesRE(tt.wantErr, err) {
			t.Errorf("DestroyDB(%s, %s) error = %s, want %s", tt.dbName, "", err, tt.wantErr)
		}

		if err != nil {
			return
		}

		var found bool
		if err := tt.client.pool.QueryRow(t.Context(), `
			SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_name   = $1
			)
		`, tablePrefix+tt.dbName).Scan(&found); err != nil {
			t.Fatal(err)
		}
		if found {
			t.Fatal("expected database to have been deleted")
		}
	})
}
