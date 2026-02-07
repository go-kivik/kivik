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

package db

import (
	"context"
	"testing"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.RegisterV2("ViewCleanup", viewCleanup)
}

func viewCleanup(t *testing.T, c *kt.ContextCore) {
	t.Helper()
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		c.RunAdmin(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			testViewCleanup(t, c, c.Admin)
		})
		c.RunNoAuth(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			testViewCleanup(t, c, c.NoAuth)
		})
	})
}

func testViewCleanup(t *testing.T, c *kt.ContextCore, client *kivik.Client) { //nolint:thelper
	dbname := c.TestDB(t)
	db := client.DB(dbname, c.Options(t, "db"))
	if err := db.Err(); err != nil {
		t.Fatalf("Failed to connect to db: %s", err)
	}
	c.CheckError(t, db.ViewCleanup(context.Background()))
}
