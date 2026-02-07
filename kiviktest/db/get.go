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

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.RegisterV2("Get", get)
}

type testDoc struct {
	ID   string `json:"_id"`
	Rev  string `json:"_rev,omitempty"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func get(t *testing.T, c *kt.ContextCore) {
	t.Helper()
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		const age = 32
		dbName := c.TestDB(t)
		db := c.Admin.DB(dbName, c.Options(t, "db"))
		if err := db.Err(); err != nil {
			t.Fatalf("Failed to connect to test db: %s", err)
		}

		doc := &testDoc{
			ID:   "bob",
			Name: "Robert",
			Age:  age,
		}
		rev, err := db.Put(context.Background(), doc.ID, doc)
		if err != nil {
			t.Fatalf("Failed to create doc in test db: %s", err)
		}
		doc.Rev = rev

		ddoc := &testDoc{
			ID:   "_design/foo",
			Name: "Designer",
		}
		rev, err = db.Put(context.Background(), ddoc.ID, ddoc)
		if err != nil {
			t.Fatalf("Failed to create design doc in test db: %s", err)
		}
		ddoc.Rev = rev

		local := &testDoc{
			ID:   "_local/foo",
			Name: "Designer",
		}
		rev, err = db.Put(context.Background(), local.ID, local)
		if err != nil {
			t.Fatalf("Failed to create local doc in test db: %s", err)
		}
		local.Rev = rev

		c.RunAdmin(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			db := c.Admin.DB(dbName, c.Options(t, "db"))
			if err := db.Err(); !c.IsExpectedSuccess(t, err) {
				return
			}
			testGet(t, c, db, doc)
			testGet(t, c, db, ddoc)
			testGet(t, c, db, local)
			testGet(t, c, db, &testDoc{ID: "bogus"})
		})
		c.RunNoAuth(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			db := c.NoAuth.DB(dbName, c.Options(t, "db"))
			if err := db.Err(); !c.IsExpectedSuccess(t, err) {
				return
			}
			testGet(t, c, db, doc)
			testGet(t, c, db, ddoc)
			testGet(t, c, db, local)
			testGet(t, c, db, &testDoc{ID: "bogus"})
		})
	})
}

func testGet(t *testing.T, c *kt.ContextCore, db *kivik.DB, expectedDoc *testDoc) { //nolint:thelper
	c.Run(t, expectedDoc.ID, func(t *testing.T) {
		t.Parallel()
		doc := &testDoc{}
		if !c.IsExpectedSuccess(t, db.Get(context.Background(), expectedDoc.ID).ScanDoc(&doc)) {
			return
		}
		if d := testy.DiffAsJSON(expectedDoc, doc); d != nil {
			t.Errorf("Fetched document not as expected:\n%s\n", d)
		}
	})
}
