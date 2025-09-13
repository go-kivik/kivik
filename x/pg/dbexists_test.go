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

	tests := testy.NewTable[test]()
	tests.Add("invalid db name", test{
		client: testClient(t),
		dbName: "Capitalized",
		want:   false,
	})
	tests.AddFunc("db exists", func(t *testing.T) test {
		client := testClient(t)
		const dbName = "testdb"

		if err := client.CreateDB(t.Context(), dbName, nil); err != nil {
			t.Fatalf("Failed to create test db: %s", err)
		}

		return test{
			client: client,
			dbName: dbName,
			want:   true,
		}
	})
	tests.AddFunc("db connection error", func(t *testing.T) test {
		client := testClient(t)
		client.pool.Close()

		return test{
			client:     client,
			dbName:     "asdfsadf",
			wantErr:    "closed pool",
			wantStatus: http.StatusInternalServerError,
		}
	})

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
