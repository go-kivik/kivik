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

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gitlab.com/flimzy/testy/v2"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

func TestAllDBs(t *testing.T) {
	t.Parallel()

	type test struct {
		client     *client
		options    driver.Options
		want       []string
		wantErr    string
		wantStatus int
	}

	tests := testy.NewTable[test]()
	tests.Add("no databases found", test{
		client: testClient(t),
		want:   []string{},
	})
	tests.AddFunc("some dbs exist", func(t *testing.T) test {
		client := testClient(t)

		const dbName1, dbName2 = "testdb1", "testdb2"
		if err := client.CreateDB(t.Context(), dbName1, nil); err != nil {
			t.Fatalf("Failed to create test db: %s", err)
		}
		if err := client.CreateDB(t.Context(), dbName2, nil); err != nil {
			t.Fatalf("Failed to create test db: %s", err)
		}

		return test{
			client: client,
			want:   []string{dbName1, dbName2},
		}
	})
	tests.AddFunc("exclude non-kivik tables", func(t *testing.T) test {
		client := testClient(t)

		const dbName = "testdb"
		if err := client.CreateDB(t.Context(), dbName, nil); err != nil {
			t.Fatalf("Failed to create test db: %s", err)
		}
		if _, err := client.pool.Exec(t.Context(), "CREATE TABLE foo (id TEXT)"); err != nil {
			t.Fatalf("Failed to create non-kivik table: %s", err)
		}

		return test{
			client: client,
			want:   []string{dbName},
		}
	})
	tests.AddFunc("db connection error", func(t *testing.T) test {
		client := testClient(t)

		client.pool.Close()

		return test{
			client:     client,
			wantErr:    "closed pool",
			wantStatus: http.StatusInternalServerError,
		}
	})
	tests.AddFunc("option: descending=true", func(t *testing.T) test {
		client := testClient(t)

		const dbName1, dbName2 = "testdb1", "testdb2"
		if err := client.CreateDB(t.Context(), dbName1, nil); err != nil {
			t.Fatalf("Failed to create test db: %s", err)
		}
		if err := client.CreateDB(t.Context(), dbName2, nil); err != nil {
			t.Fatalf("Failed to create test db: %s", err)
		}

		return test{
			client:  client,
			options: kivik.Param("descending", "true"),
			want:    []string{dbName2, dbName1},
		}
	})
	tests.AddFunc("end key", func(t *testing.T) test {
		client := testClient(t)

		const dbName1, dbName2 = "testdb1", "testdb2"
		if err := client.CreateDB(t.Context(), dbName1, nil); err != nil {
			t.Fatalf("Failed to create test db: %s", err)
		}
		if err := client.CreateDB(t.Context(), dbName2, nil); err != nil {
			t.Fatalf("Failed to create test db: %s", err)
		}

		return test{
			client:  client,
			options: kivik.Param("endkey", dbName1),
			want:    []string{dbName1},
		}
	})

	/*
		TODO:
		- options:
			- limit (number) – Limit the number of the returned databases to the specified number.
			- skip (number) – Skip this number of databases before starting to return the results. Default is 0.
			- startkey (json) – Return databases starting with the specified key.
			- start_key (json) – Alias for startkey.
			- inclusive_end (boolean) – If true, the specified endkey will be included in the result. Default is true.
	*/

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()

		got, err := tt.client.AllDBs(t.Context(), tt.options)
		if !testy.ErrorMatchesRE(tt.wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status: %d", status)
		}
		if d := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); d != "" {
			t.Errorf("Unexpected result:\n%s", d)
		}
	})
}
