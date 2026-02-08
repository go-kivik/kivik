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

package kt

import (
	"testing"

	"gitlab.com/flimzy/testy"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/mockdb"
)

func TestContextDB(t *testing.T) {
	t.Parallel()

	type tt struct {
		ctx    *Context
		client *kivik.Client
		dbname string
	}

	tests := testy.NewTable()

	tests.Add("success", func(t *testing.T) interface{} {
		client, mock, err := mockdb.New()
		if err != nil {
			t.Fatal(err)
		}
		mock.ExpectDB().WithName("testdb")

		return tt{
			ctx:    &Context{},
			client: client,
			dbname: "testdb",
		}
	})
	tests.Add("passes options from config", func(t *testing.T) interface{} {
		client, mock, err := mockdb.New()
		if err != nil {
			t.Fatal(err)
		}
		opt := kivik.Param("foo", "bar")
		mock.ExpectDB().WithName("testdb").WithOptions(opt)

		return tt{
			ctx:    &Context{Config: SuiteConfig{"db": opt}},
			client: client,
			dbname: "testdb",
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		t.Parallel()
		db := tt.ctx.DB(t, tt.client, tt.dbname)
		if err := db.Err(); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	})
}

func TestContextAdminDB(t *testing.T) {
	t.Parallel()

	type tt struct {
		ctx    *Context
		dbname string
	}

	tests := testy.NewTable()

	tests.Add("success", func(t *testing.T) interface{} {
		client, mock, err := mockdb.New()
		if err != nil {
			t.Fatal(err)
		}
		mock.ExpectDB().WithName("admindb")

		return tt{
			ctx:    &Context{Admin: client},
			dbname: "admindb",
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		t.Parallel()
		db := tt.ctx.AdminDB(t, tt.dbname)
		if err := db.Err(); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	})
}
