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

package sqlite

import (
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestClientCreateDB(t *testing.T) {
	t.Run("invalid name", func(t *testing.T) {
		d := drv{}
		dClient, err := d.NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}

		err = dClient.CreateDB(t.Context(), "Foo", mock.NilOption)
		if err == nil {
			t.Fatal("err should not be nil")
		}
		const wantErr = "invalid database name: Foo"
		if !testy.ErrorMatches(wantErr, err) {
			t.Fatalf("Unexpected error: %s", err)
		}
		const wantStatus = http.StatusBadRequest
		if status := kivik.HTTPStatus(err); status != wantStatus {
			t.Fatalf("status should be %d", wantStatus)
		}
	})
	t.Run("success", func(t *testing.T) {
		d := drv{}
		dClient, err := d.NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}

		if err := dClient.CreateDB(t.Context(), "foo", mock.NilOption); err != nil {
			t.Fatal(err)
		}

		exists, err := dClient.DBExists(t.Context(), "foo", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}
		if !exists {
			t.Fatal("foo should exist")
		}
	})
	t.Run("db already exists", func(t *testing.T) {
		d := drv{}
		dClient, err := d.NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}

		if err := dClient.CreateDB(t.Context(), "foo", mock.NilOption); err != nil {
			t.Fatal(err)
		}

		err = dClient.CreateDB(t.Context(), "foo", mock.NilOption)
		if err == nil {
			t.Fatal("err should not be nil")
		}
		const wantErr = "database already exists"
		if err.Error() != wantErr {
			t.Fatalf("err should be %s, got %s", wantErr, err)
		}
		const wantStatus = http.StatusPreconditionFailed
		if status := kivik.HTTPStatus(err); status != wantStatus {
			t.Fatalf("status should be %d", wantStatus)
		}
	})
}
