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
	"strings"
	"testing"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.Register("GetRev", getRev)
}

func getRev(t *testing.T, c *kt.Context) {
	t.Helper()
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		dbName := c.TestDB(t)
		db := c.Admin.DB(dbName, c.Options(t, "db"))
		if err := db.Err(); err != nil {
			t.Fatalf("Failed to connect to test db: %s", err)
		}
		doc := &testDoc{
			ID: "bob",
		}
		rev, err := db.Put(context.Background(), doc.ID, doc)
		if err != nil {
			t.Fatalf("Failed to create doc in test db: %s", err)
		}
		doc.Rev = rev

		ddoc := &testDoc{
			ID: "_design/foo",
		}
		rev, err = db.Put(context.Background(), ddoc.ID, ddoc)
		if err != nil {
			t.Fatalf("Failed to create design doc in test db: %s", err)
		}
		ddoc.Rev = rev

		local := &testDoc{
			ID: "_local/foo",
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
			testGetRev(t, c, db, doc)
			testGetRev(t, c, db, ddoc)
			testGetRev(t, c, db, local)
			testGetRev(t, c, db, &testDoc{ID: "bogus"})
		})
		c.RunNoAuth(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			db := c.NoAuth.DB(dbName, c.Options(t, "db"))
			if err := db.Err(); !c.IsExpectedSuccess(t, err) {
				return
			}
			testGetRev(t, c, db, doc)
			testGetRev(t, c, db, ddoc)
			testGetRev(t, c, db, local)
			testGetRev(t, c, db, &testDoc{ID: "bogus"})
		})
	})
}

func testGetRev(t *testing.T, c *kt.Context, db *kivik.DB, expectedDoc *testDoc) { //nolint:thelper
	c.Run(t, expectedDoc.ID, func(t *testing.T) {
		t.Parallel()
		rev, err := db.GetRev(context.Background(), expectedDoc.ID)
		if !c.IsExpectedSuccess(t, err) {
			return
		}
		doc := &testDoc{}
		if err = db.Get(context.Background(), expectedDoc.ID).ScanDoc(&doc); err != nil {
			t.Fatalf("Failed to scan doc: %s\n", err)
		}
		if strings.HasPrefix(expectedDoc.ID, "_local/") {
			// Revisions are meaningless for _local docs
			return
		}
		if rev != doc.Rev {
			t.Errorf("Unexpected rev. Expected: %s, Actual: %s", doc.Rev, rev)
		}
	})
}
