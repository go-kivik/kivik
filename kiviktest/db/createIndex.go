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
	kt.RegisterV2("CreateIndex", createIndex)
}

func createIndex(t *testing.T, c *kt.ContextCore) {
	t.Helper()
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		c.RunAdmin(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			testCreateIndex(t, c, c.Admin)
		})
		c.RunNoAuth(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			testCreateIndex(t, c, c.NoAuth)
		})
	})
}

func testCreateIndex(t *testing.T, c *kt.ContextCore, client *kivik.Client) { //nolint:thelper
	dbname := c.TestDB(t)
	db := client.DB(dbname, c.Options(t, "db"))
	if err := db.Err(); err != nil {
		t.Fatalf("Failed to open db: %s", err)
	}
	c.Run(t, "Valid", func(t *testing.T) {
		doCreateIndex(t, c, db, `{"fields":["foo"]}`)
	})
	c.Run(t, "NilIndex", func(t *testing.T) {
		doCreateIndex(t, c, db, nil)
	})
	c.Run(t, "BlankIndex", func(t *testing.T) {
		doCreateIndex(t, c, db, "")
	})
	c.Run(t, "EmptyIndex", func(t *testing.T) {
		doCreateIndex(t, c, db, "{}")
	})
	c.Run(t, "InvalidIndex", func(t *testing.T) {
		doCreateIndex(t, c, db, `{"oink":true}`)
	})
	c.Run(t, "InvalidJSON", func(t *testing.T) {
		doCreateIndex(t, c, db, `chicken`)
	})
}

func doCreateIndex(t *testing.T, c *kt.ContextCore, db *kivik.DB, index any) { //nolint:thelper
	t.Parallel()
	err := kt.Retry(func() error {
		return db.CreateIndex(context.Background(), "", "", index)
	})
	c.CheckError(t, err)
}
