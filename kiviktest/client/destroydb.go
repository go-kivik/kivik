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

package client

import (
	"context"
	"testing"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.Register("DestroyDB", destroyDB)
}

func destroyDB(t *testing.T, c *kt.Context) {
	t.Helper()
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		c.RunAdmin(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			testDestroy(t, c, c.Admin)
		})
		c.RunNoAuth(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			testDestroy(t, c, c.NoAuth)
		})
	})
}

func testDestroy(t *testing.T, c *kt.Context, client *kivik.Client) { //nolint:thelper
	c.Run(t, "ExistingDB", func(t *testing.T) {
		t.Parallel()
		dbName := c.TestDB(t)
		c.CheckError(t, client.DestroyDB(context.Background(), dbName, c.Options(t, "db")))
	})
	c.Run(t, "NonExistantDB", func(t *testing.T) {
		t.Parallel()
		c.CheckError(t, client.DestroyDB(context.Background(), kt.TestDBName(t), c.Options(t, "db")))
	})
}
