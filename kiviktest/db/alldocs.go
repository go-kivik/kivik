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

// Package db provides integration tests for the kivik db.
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
	kt.RegisterV2("AllDocs", allDocs)
}

func allDocs(t *testing.T, c *kt.ContextCore) {
	t.Helper()
	c.RunAdmin(t, func(t *testing.T) {
		t.Helper()
		testAllDocs(t, c, c.Admin)
	})
	c.RunNoAuth(t, func(t *testing.T) {
		t.Helper()
		testAllDocs(t, c, c.NoAuth)
	})
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		testAllDocsRW(t, c)
	})
}

func testAllDocsRW(t *testing.T, c *kt.ContextCore) {
	t.Helper()
	if c.Admin == nil {
		return
	}
	dbName, expected, err := setUpAllDocsTest(t, c)
	if err != nil {
		t.Errorf("Failed to set up temp db: %s", err)
	}
	c.RunAdmin(t, func(t *testing.T) {
		t.Helper()
		doTest(t, c, c.Admin, dbName, 0, expected, true)
	})
	c.RunNoAuth(t, func(t *testing.T) {
		t.Helper()
		doTest(t, c, c.NoAuth, dbName, 0, expected, true)
	})
}

func setUpAllDocsTest(t *testing.T, c *kt.ContextCore) (dbName string, docIDs []string, err error) {
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

func testAllDocs(t *testing.T, c *kt.ContextCore, client *kivik.Client) { //nolint:thelper
	if !c.IsSet(t, "databases") {
		t.Errorf("databases not set; Did you configure this test?")
		return
	}
	for _, dbName := range c.StringSlice(t, "databases") {
		func(dbName string) {
			c.Run(t, dbName, func(t *testing.T) {
				t.Helper()
				doTest(t, c, client, dbName, int64(c.Int(t, "offset")), c.StringSlice(t, "expected"), false)
			})
		}(dbName)
	}
}

func doTest(t *testing.T, c *kt.ContextCore, client *kivik.Client, dbName string, expOffset int64, expected []string, exact bool) { //nolint:thelper
	c.Run(t, "WithDocs", func(t *testing.T) {
		t.Helper()
		doTestWithDocs(t, c, client, dbName, expOffset, expected, exact)
	})
	c.Run(t, "WithoutDocs", func(t *testing.T) {
		t.Helper()
		doTestWithoutDocs(t, c, client, dbName, expOffset, expected, exact)
	})
}

func doTestWithoutDocs(t *testing.T, c *kt.ContextCore, client *kivik.Client, dbName string, expOffset int64, expected []string, exact bool) { //nolint:thelper
	t.Parallel()
	db := client.DB(dbName, c.Options(t, "db"))
	// Errors may be deferred here, so only return if we actually get
	// an error.
	if err := db.Err(); err != nil && !c.IsExpectedSuccess(t, err) {
		return
	}

	rows := db.AllDocs(context.Background())
	if !c.IsExpectedSuccess(t, rows.Err()) {
		return
	}
	docIDs := make([]string, 0, len(expected))
	for rows.Next() {
		id, _ := rows.ID()
		docIDs = append(docIDs, id)
	}
	meta, err := rows.Metadata()
	if err != nil {
		t.Fatalf("Failed to fetch row: %s", rows.Err())
	}
	testExpectedDocs(t, expected, docIDs, exact)
	if expOffset != meta.Offset {
		t.Errorf("offset: Expected %d, got %d", expOffset, meta.Offset)
	}
	if exact {
		if int64(len(expected)) != meta.TotalRows {
			t.Errorf("total rows: Expected %d, got %d", len(expected), meta.TotalRows)
		}
	}
}

func doTestWithDocs(t *testing.T, c *kt.ContextCore, client *kivik.Client, dbName string, expOffset int64, expected []string, exact bool) { //nolint:thelper
	t.Parallel()
	db := client.DB(dbName, c.Options(t, "db"))
	// Errors may be deferred here, so only return if we actually get
	// an error.
	if err := db.Err(); err != nil && !c.IsExpectedSuccess(t, err) {
		return
	}
	opts := kivik.Params(map[string]any{
		"include_docs": true,
		"update_seq":   true,
	})

	rows := db.AllDocs(context.Background(), opts)
	if !c.IsExpectedSuccess(t, rows.Err()) {
		return
	}
	docIDs := make([]string, 0, len(expected))
	for rows.Next() {
		var doc struct {
			ID  string `json:"_id"`
			Rev string `json:"_rev"`
		}
		if err := rows.ScanDoc(&doc); err != nil {
			t.Errorf("Failed to scan doc: %s", err)
		}
		var value struct {
			Rev string `json:"rev"`
		}
		if err := rows.ScanValue(&value); err != nil {
			t.Errorf("Failed to scan value: %s", err)
		}
		if value.Rev != doc.Rev {
			t.Errorf("doc._rev = %s, but value.rev = %s", doc.Rev, value.Rev)
		}
		id, _ := rows.ID()
		if doc.ID != id {
			t.Errorf("doc._id = %s, but rows.ID = %s", doc.ID, id)
		}
		docIDs = append(docIDs, id)
	}
	meta, err := rows.Metadata()
	if err != nil {
		t.Fatalf("Failed to fetch row: %s", rows.Err())
	}
	testExpectedDocs(t, expected, docIDs, exact)
	if expOffset != meta.Offset {
		t.Errorf("offset: Expected %d, got %d", expOffset, meta.Offset)
	}
	c.Run(t, "UpdateSeq", func(t *testing.T) {
		if meta.UpdateSeq == "" {
			t.Errorf("Expected updated sequence")
		}
	})
	if exact {
		if int64(len(expected)) != meta.TotalRows {
			t.Errorf("total rows: Expected %d, got %d", len(expected), meta.TotalRows)
		}
	}
}

func testExpectedDocs(t *testing.T, expected, actual []string, exact bool) { //nolint:thelper
	if exact {
		if d := testy.DiffTextSlices(expected, actual); d != nil {
			t.Errorf("Unexpected document IDs returned:\n%s\n", d)
		}
		return
	}
	actualIDs := make(map[string]struct{})
	for _, id := range actual {
		actualIDs[id] = struct{}{}
	}
	for _, id := range expected {
		if _, ok := actualIDs[id]; !ok {
			t.Errorf("Expected document '%s' not found", id)
		}
	}
}
