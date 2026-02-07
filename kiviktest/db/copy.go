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
	kt.RegisterV2("Copy", _copy)
}

func _copy(t *testing.T, c *kt.ContextCore) {
	t.Helper()
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		dbname := c.TestDB(t)
		db := c.Admin.DB(dbname, c.Options(t, "db"))
		if err := db.Err(); err != nil {
			t.Fatalf("Failed to open db: %s", err)
		}

		doc := map[string]string{
			"_id":  "foo",
			"name": "Robert",
		}
		rev, err := db.Put(context.Background(), doc["_id"], doc)
		if err != nil {
			t.Fatalf("Failed to create source doc: %s", err)
		}
		doc["_rev"] = rev

		ddoc := map[string]string{
			"_id":  "_design/foo",
			"name": "Robert",
		}
		rev, err = db.Put(context.Background(), ddoc["_id"], ddoc)
		if err != nil {
			t.Fatalf("Failed to create source design doc: %s", err)
		}
		ddoc["_rev"] = rev

		local := map[string]string{
			"_id":  "_local/foo",
			"name": "Robert",
		}
		rev, err = db.Put(context.Background(), local["_id"], local)
		if err != nil {
			t.Fatalf("Failed to create source design doc: %s", err)
		}
		local["_rev"] = rev

		c.RunAdmin(t, func(t *testing.T) {
			t.Helper()
			copyTest(t, c, c.Admin, dbname, doc)
			copyTest(t, c, c.Admin, dbname, ddoc)
			copyTest(t, c, c.Admin, dbname, local)
		})
		c.RunNoAuth(t, func(t *testing.T) {
			t.Helper()
			copyTest(t, c, c.NoAuth, dbname, doc)
			copyTest(t, c, c.NoAuth, dbname, ddoc)
			copyTest(t, c, c.NoAuth, dbname, local)
		})
	})
}

func copyTest(t *testing.T, c *kt.ContextCore, client *kivik.Client, dbname string, source map[string]string) { //nolint:thelper
	c.Run(t, source["_id"], func(t *testing.T) {
		t.Parallel()
		db := client.DB(dbname, c.Options(t, "db"))
		if err := db.Err(); err != nil {
			t.Fatalf("Failed to open db: %s", err)
		}
		targetID := kt.TestDBName(t)
		rev, err := db.Copy(context.Background(), targetID, source["_id"])
		if !c.IsExpectedSuccess(t, err) {
			return
		}
		c.Run(t, "RevCopy", func(t *testing.T) {
			cp := map[string]string{
				"_id":  targetID,
				"name": "Bob",
				"_rev": rev,
			}
			if _, e := db.Put(context.Background(), targetID, cp); e != nil {
				t.Fatalf("Failed to update copy: %s", e)
			}
			targetID2 := kt.TestDBName(t)
			if _, e := db.Copy(context.Background(), targetID2, targetID, kivik.Rev(rev)); e != nil {
				t.Fatalf("Failed to copy doc with rev option: %s", e)
			}
			var readCopy map[string]string
			if err = db.Get(context.Background(), targetID2).ScanDoc(&readCopy); err != nil {
				t.Fatalf("Failed to scan copy: %s", err)
			}
			if readCopy["name"] != "Robert" {
				t.Errorf("Copy-with-rev failed. Name = %s, expected %s", readCopy["name"], "Robert")
			}
		})
	})
}
