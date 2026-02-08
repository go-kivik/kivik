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
	kt.Register("DeleteIndex", delindex)
}

func delindex(t *testing.T, c *kt.Context) {
	t.Helper()
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		c.RunAdmin(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			testDelIndex(t, c, c.Admin)
		})
		c.RunNoAuth(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			testDelIndex(t, c, c.NoAuth)
		})
	})
}

func testDelIndex(t *testing.T, c *kt.Context, client *kivik.Client) { //nolint:thelper
	dbname := c.TestDB(t)
	// t.Cleanup(func() { c.Admin.DestroyDB(context.Background(), dbname, c.Options(t, "db")) }) // nolint: errcheck
	dba := c.AdminDB(t, dbname)
	if err := dba.CreateIndex(context.Background(), "foo", "bar", `{"fields":["foo"]}`); err != nil {
		t.Fatalf("Failed to create index: %s", err)
	}
	db := c.DB(t, client, dbname)
	c.Run(t, "ValidIndex", func(t *testing.T) {
		t.Parallel()
		c.CheckError(t, db.DeleteIndex(context.Background(), "foo", "bar"))
	})
	c.Run(t, "NotFoundDdoc", func(t *testing.T) {
		t.Parallel()
		c.CheckError(t, db.DeleteIndex(context.Background(), "notFound", "bar"))
	})
	c.Run(t, "NotFoundName", func(t *testing.T) {
		t.Parallel()
		c.CheckError(t, db.DeleteIndex(context.Background(), "foo", "notFound"))
	})
}
