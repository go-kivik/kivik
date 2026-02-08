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
	kt.Register("Delete", _delete)
}

func _delete(t *testing.T, c *kt.Context) {
	t.Helper()
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		c.RunAdmin(t, func(t *testing.T) {
			t.Helper()
			testDelete(t, c, c.Admin)
		})
		c.RunNoAuth(t, func(t *testing.T) {
			t.Helper()
			testDelete(t, c, c.NoAuth)
		})
	})
}

type deleteDoc struct {
	ID      string `json:"_id"`
	Rev     string `json:"_rev,omitempty"`
	Deleted bool   `json:"_deleted"`
}

func testDelete(t *testing.T, c *kt.Context, client *kivik.Client) { //nolint:thelper
	t.Parallel()
	dbName := c.TestDB(t)
	admdb := c.Admin.DB(dbName, c.Options(t, "db"))
	if err := admdb.Err(); err != nil {
		t.Errorf("Failed to connect to db as admin: %s", err)
	}
	db := client.DB(dbName, c.Options(t, "db"))
	if err := db.Err(); err != nil {
		t.Errorf("Failed to connect to db: %s", err)
		return
	}

	doc := &deleteDoc{
		ID: kt.TestDBName(t),
	}
	rev, err := admdb.Put(context.Background(), doc.ID, doc)
	if err != nil {
		t.Errorf("Failed to create test doc: %s", err)
		return
	}
	doc.Rev = rev

	doc2 := &deleteDoc{
		ID: kt.TestDBName(t),
	}
	rev, err = admdb.Put(context.Background(), doc2.ID, doc2)
	if err != nil {
		t.Errorf("Failed to create test doc: %s", err)
		return
	}
	doc2.Rev = rev

	ddoc := &testDoc{
		ID: "_design/foo",
	}
	rev, err = admdb.Put(context.Background(), ddoc.ID, ddoc)
	if err != nil {
		t.Fatalf("Failed to create design doc in test db: %s", err)
	}
	ddoc.Rev = rev

	local := &testDoc{
		ID: "_local/foo",
	}
	rev, err = admdb.Put(context.Background(), local.ID, local)
	if err != nil {
		t.Fatalf("Failed to create local doc in test db: %s", err)
	}
	local.Rev = rev

	c.Run(t, "WrongRev", func(t *testing.T) {
		t.Parallel()
		_, err := db.Delete(context.Background(), doc2.ID, "1-9c65296036141e575d32ba9c034dd3ee")
		c.CheckError(t, err)
	})
	c.Run(t, "InvalidRevFormat", func(t *testing.T) {
		t.Parallel()
		_, err := db.Delete(context.Background(), doc2.ID, "invalid rev format")
		c.CheckError(t, err)
	})
	c.Run(t, "MissingDoc", func(t *testing.T) {
		t.Parallel()
		_, err := db.Delete(context.Background(), "missing doc", "1-9c65296036141e575d32ba9c034dd3ee")
		c.CheckError(t, err)
	})
	c.Run(t, "ValidRev", func(t *testing.T) {
		t.Parallel()
		_, err := db.Delete(context.Background(), doc.ID, doc.Rev)
		c.CheckError(t, err)
	})
	c.Run(t, "DesignDoc", func(t *testing.T) {
		t.Parallel()
		_, err := db.Delete(context.Background(), ddoc.ID, ddoc.Rev)
		c.CheckError(t, err)
	})
	c.Run(t, "Local", func(t *testing.T) {
		t.Parallel()
		_, err := db.Delete(context.Background(), local.ID, local.Rev)
		c.CheckError(t, err)
	})
}
