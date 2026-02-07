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
	kt.RegisterV2("Query", query)
}

var ddoc = map[string]any{
	"_id":      "_design/testddoc",
	"language": "javascript",
	"views": map[string]any{
		"testview": map[string]any{
			"map": `function(doc) {
                if (doc.include) {
                    emit(doc._id, doc.index);
                }
            }`,
		},
	},
}

func query(t *testing.T, c *kt.ContextCore) {
	t.Helper()
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		testQueryRW(t, c)
	})
}

func testQueryRW(t *testing.T, c *kt.ContextCore) {
	t.Helper()
	if c.Admin == nil {
		return
	}
	dbName, expected, err := setUpQueryTest(t, c)
	if err != nil {
		t.Errorf("Failed to set up temp db: %s", err)
	}
	c.RunAdmin(t, func(t *testing.T) {
		t.Helper()
		doQueryTest(t, c, c.Admin, dbName, 0, expected)
	})
	c.RunNoAuth(t, func(t *testing.T) {
		t.Helper()
		doQueryTest(t, c, c.NoAuth, dbName, 0, expected)
	})
}

func setUpQueryTest(t *testing.T, c *kt.ContextCore) (dbName string, docIDs []string, err error) {
	t.Helper()
	dbName = c.TestDB(t)
	db := c.Admin.DB(dbName, c.Options(t, "db"))
	if err := db.Err(); err != nil {
		return dbName, nil, fmt.Errorf("fialed to connect to db: %w", err)
	}
	if _, err := db.Put(context.Background(), ddoc["_id"].(string), ddoc); err != nil {
		return dbName, nil, fmt.Errorf("fialed to create design doc: %w", err)
	}
	const maxDocs = 10
	docIDs = make([]string, maxDocs)
	for i := range docIDs {
		id := kt.TestDBName(t)
		doc := struct {
			ID      string `json:"id"`
			Include bool   `json:"include"`
			Index   int    `json:"index"`
		}{
			ID:      id,
			Include: true,
			Index:   i,
		}
		if _, err := db.Put(context.Background(), doc.ID, doc); err != nil {
			return dbName, nil, fmt.Errorf("failed to create doc: %w", err)
		}
		docIDs[i] = id
	}
	sort.Strings(docIDs)
	return dbName, docIDs, nil
}

func doQueryTest(t *testing.T, c *kt.ContextCore, client *kivik.Client, dbName string, expOffset int64, expected []string) { //nolint:thelper
	c.Run(t, "WithDocs", func(t *testing.T) {
		doQueryTestWithDocs(t, c, client, dbName, expOffset, expected)
	})
	c.Run(t, "WithoutDocs", func(t *testing.T) {
		doQueryTestWithoutDocs(t, c, client, dbName, expOffset, expected)
	})
}

func doQueryTestWithoutDocs(t *testing.T, c *kt.ContextCore, client *kivik.Client, dbName string, expOffset int64, expected []string) { //nolint:thelper
	t.Parallel()
	db := client.DB(dbName, c.Options(t, "db"))
	// Errors may be deferred here, so only return if we actually get
	// an error.
	if err := db.Err(); err != nil && !c.IsExpectedSuccess(t, err) {
		return
	}

	rows := db.Query(context.Background(), "testddoc", "testview")
	if !c.IsExpectedSuccess(t, rows.Err()) {
		return
	}
	docIDs := make([]string, 0, len(expected))
	var scanTested bool
	for rows.Next() {
		id, _ := rows.ID()
		docIDs = append(docIDs, id)
		if !scanTested {
			scanTested = true
			c.Run(t, "ScanDoc", func(t *testing.T) {
				var i any
				c.CheckError(t, rows.ScanDoc(&i))
			})
			c.Run(t, "ScanValue", func(t *testing.T) {
				var i any
				c.CheckError(t, rows.ScanValue(&i))
			})
		}
	}
	meta, err := rows.Metadata()
	if err != nil {
		t.Fatalf("Failed to fetch row: %s", rows.Err())
	}
	if d := testy.DiffTextSlices(expected, docIDs); d != nil {
		t.Errorf("Unexpected document IDs returned:\n%s\n", d)
	}
	if expOffset != meta.Offset {
		t.Errorf("offset: Expected %d, got %d", expOffset, meta.Offset)
	}
	if int64(len(expected)) != meta.TotalRows {
		t.Errorf("total rows: Expected %d, got %d", len(expected), meta.TotalRows)
	}
}

func doQueryTestWithDocs(t *testing.T, c *kt.ContextCore, client *kivik.Client, dbName string, expOffset int64, expected []string) { //nolint:thelper
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

	rows := db.Query(context.Background(), "testddoc", "testview", opts)
	if !c.IsExpectedSuccess(t, rows.Err()) {
		return
	}
	docIDs := make([]string, 0, len(expected))
	for rows.Next() {
		var doc struct {
			ID    string `json:"_id"`
			Rev   string `json:"_rev"`
			Index int    `json:"index"`
		}
		if err := rows.ScanDoc(&doc); err != nil {
			t.Errorf("Failed to scan doc: %s", err)
		}
		var value int
		if err := rows.ScanValue(&value); err != nil {
			t.Errorf("Failed to scan value: %s", err)
		}
		if value != doc.Index {
			t.Errorf("doc._rev = %d, but value = %d", doc.Index, value)
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
	if d := testy.DiffTextSlices(expected, docIDs); d != nil {
		t.Errorf("Unexpected document IDs returned:\n%s\n", d)
	}
	if expOffset != meta.Offset {
		t.Errorf("offset: Expected %d, got %d", expOffset, meta.Offset)
	}
	c.Run(t, "UpdateSeq", func(t *testing.T) {
		if meta.UpdateSeq == "" {
			t.Errorf("Expected updated sequence")
		}
	})
	if int64(len(expected)) != meta.TotalRows {
		t.Errorf("total rows: Expected %d, got %d", len(expected), meta.TotalRows)
	}
}
