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
	"fmt"
	"sort"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.RegisterV2("Find", find)
}

func find(t *testing.T, c *kt.ContextCore) {
	t.Helper()
	c.RunAdmin(t, func(t *testing.T) {
		t.Helper()
		testFind(t, c, c.Admin)
	})
	c.RunNoAuth(t, func(t *testing.T) {
		t.Helper()
		testFind(t, c, c.NoAuth)
	})
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		testFindRW(t, c)
	})
}

func testFindRW(t *testing.T, c *kt.ContextCore) {
	t.Helper()
	if c.Admin == nil {
		return
	}
	dbName, expected, err := setUpFindTest(t, c)
	if err != nil {
		t.Errorf("Failed to set up temp db: %s", err)
	}
	c.RunAdmin(t, func(t *testing.T) {
		t.Helper()
		doFindTest(t, c, c.Admin, dbName, 0, expected)
	})
	c.RunNoAuth(t, func(t *testing.T) {
		t.Helper()
		doFindTest(t, c, c.NoAuth, dbName, 0, expected)
	})
}

func setUpFindTest(t *testing.T, c *kt.ContextCore) (dbName string, docIDs []string, err error) {
	t.Helper()
	dbName = c.TestDB(t)
	db := c.Admin.DB(dbName, c.Options(t, "db"))
	if err := db.Err(); err != nil {
		return dbName, nil, fmt.Errorf("failed to connect to db: %w", err)
	}
	const maxDocs = 10
	docIDs = make([]string, maxDocs)
	for i := range docIDs {
		id := kt.TestDBName(t)
		doc := struct {
			ID string `json:"id"`
		}{
			ID: id,
		}
		if _, err := db.Put(context.Background(), doc.ID, doc); err != nil {
			return dbName, nil, fmt.Errorf("failed to create doc: %w", err)
		}
		docIDs[i] = id
	}
	sort.Strings(docIDs)
	return dbName, docIDs, nil
}

func testFind(t *testing.T, c *kt.ContextCore, client *kivik.Client) { //nolint:thelper
	if !c.IsSet(t, "databases") {
		t.Errorf("databases not set; Did you configure this test?")
		return
	}
	for _, dbName := range c.StringSlice(t, "databases") {
		func(dbName string) {
			c.Run(t, dbName, func(t *testing.T) {
				doFindTest(t, c, client, dbName, int64(c.Int(t, "offset")), c.StringSlice(t, "expected"))
			})
		}(dbName)
	}
}

func doFindTest(t *testing.T, c *kt.ContextCore, client *kivik.Client, dbName string, expOffset int64, expected []string) { //nolint:thelper
	t.Parallel()
	db := client.DB(dbName, c.Options(t, "db"))
	// Errors may be deferred here, so only return if we actually get
	// an error.
	if err := db.Err(); err != nil && !c.IsExpectedSuccess(t, err) {
		return
	}

	var rows *kivik.ResultSet
	err := kt.Retry(func() error {
		rows = db.Find(context.Background(), `{"selector":{"_id":{"$gt":null}}}`)
		return rows.Err()
	})

	if !c.IsExpectedSuccess(t, err) {
		return
	}
	docIDs := make([]string, 0, len(expected))
	for rows.Next() {
		var doc struct {
			DocID string `json:"_id"`
			Rev   string `json:"_rev"`
			ID    string `json:"id"`
		}
		if err := rows.ScanDoc(&doc); err != nil {
			t.Errorf("Failed to scan doc: %s", err)
		}
		docIDs = append(docIDs, doc.DocID)
	}
	meta, err := rows.Metadata()
	if err != nil {
		t.Fatalf("Failed to fetch row: %s", rows.Err())
	}
	sort.Strings(docIDs) // normalize order
	if d := testy.DiffTextSlices(expected, docIDs); d != nil {
		t.Errorf("Unexpected document IDs returned:\n%s\n", d)
	}
	if meta.Offset != expOffset {
		t.Errorf("Unexpected offset: %v", meta.Offset)
	}
	c.Run(t, "Warning", func(t *testing.T) {
		rows := db.Find(context.Background(), `{"selector":{"foo":{"$gt":null}}}`)
		if !c.IsExpectedSuccess(t, rows.Err()) {
			return
		}
		for rows.Next() {
		}
		if w := c.String(t, "warning"); w != meta.Warning {
			t.Errorf("Warning:\nExpected: %s\n  Actual: %s", w, meta.Warning)
		}
	})
}
