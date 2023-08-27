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
	"sort"

	"github.com/pkg/errors"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kiviktest/v4/kt"
)

func init() {
	kt.Register("Query", query)
}

func query(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		testQueryRW(ctx)
	})
}

func testQueryRW(ctx *kt.Context) {
	if ctx.Admin == nil {
		// Can't do anything here without admin access
		return
	}
	dbName, expected, err := setUpQueryTest(ctx)
	if err != nil {
		ctx.Errorf("Failed to set up temp db: %s", err)
	}
	defer ctx.DestroyDB(dbName)
	ctx.Run("group", func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			doQueryTest(ctx, ctx.Admin, dbName, 0, expected)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			doQueryTest(ctx, ctx.NoAuth, dbName, 0, expected)
		})
	})
}

var ddoc = map[string]interface{}{
	"_id":      "_design/testddoc",
	"language": "javascript",
	"views": map[string]interface{}{
		"testview": map[string]interface{}{
			"map": `function(doc) {
                if (doc.include) {
                    emit(doc._id, doc.index);
                }
            }`,
		},
	},
}

func setUpQueryTest(ctx *kt.Context) (dbName string, docIDs []string, err error) {
	dbName = ctx.TestDB()
	db := ctx.Admin.DB(dbName, ctx.Options("db"))
	if err := db.Err(); err != nil {
		return dbName, nil, errors.Wrap(err, "failed to connect to db")
	}
	if _, err := db.Put(context.Background(), ddoc["_id"].(string), ddoc); err != nil {
		return dbName, nil, errors.Wrap(err, "failed to create design doc")
	}
	docIDs = make([]string, 10)
	for i := range docIDs {
		id := ctx.TestDBName()
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
			return dbName, nil, errors.Wrap(err, "failed to create doc")
		}
		docIDs[i] = id
	}
	sort.Strings(docIDs)
	return dbName, docIDs, nil
}

func doQueryTest(ctx *kt.Context, client *kivik.Client, dbName string, expOffset int64, expected []string) {
	ctx.Run("WithDocs", func(ctx *kt.Context) {
		doQueryTestWithDocs(ctx, client, dbName, expOffset, expected)
	})
	ctx.Run("WithoutDocs", func(ctx *kt.Context) {
		doQueryTestWithoutDocs(ctx, client, dbName, expOffset, expected)
	})
}

func doQueryTestWithoutDocs(ctx *kt.Context, client *kivik.Client, dbName string, expOffset int64, expected []string) {
	ctx.Parallel()
	db := client.DB(dbName, ctx.Options("db"))
	// Errors may be deferred here, so only return if we actually get
	// an error.
	if err := db.Err(); err != nil && !ctx.IsExpectedSuccess(err) {
		return
	}

	rows := db.Query(context.Background(), "testddoc", "testview")
	if !ctx.IsExpectedSuccess(rows.Err()) {
		return
	}
	docIDs := make([]string, 0, len(expected))
	var scanTested bool
	for rows.Next() {
		id, _ := rows.ID()
		docIDs = append(docIDs, id)
		if !scanTested {
			scanTested = true
			ctx.Run("ScanDoc", func(ctx *kt.Context) {
				var i interface{}
				ctx.CheckError(rows.ScanDoc(&i))
			})
			ctx.Run("ScanValue", func(ctx *kt.Context) {
				var i interface{}
				ctx.CheckError(rows.ScanValue(&i))
			})
		}
	}
	meta, err := rows.Metadata()
	if err != nil {
		ctx.Fatalf("Failed to fetch row: %s", rows.Err())
	}
	if d := testy.DiffTextSlices(expected, docIDs); d != nil {
		ctx.Errorf("Unexpected document IDs returned:\n%s\n", d)
	}
	if expOffset != meta.Offset {
		ctx.Errorf("offset: Expected %d, got %d", expOffset, meta.Offset)
	}
	if int64(len(expected)) != meta.TotalRows {
		ctx.Errorf("total rows: Expected %d, got %d", len(expected), meta.TotalRows)
	}
}

func doQueryTestWithDocs(ctx *kt.Context, client *kivik.Client, dbName string, expOffset int64, expected []string) {
	ctx.Parallel()
	db := client.DB(dbName, ctx.Options("db"))
	// Errors may be deferred here, so only return if we actually get
	// an error.
	if err := db.Err(); err != nil && !ctx.IsExpectedSuccess(err) {
		return
	}
	opts := map[string]interface{}{
		"include_docs": true,
		"update_seq":   true,
	}

	rows := db.Query(context.Background(), "testddoc", "testview", opts)
	if !ctx.IsExpectedSuccess(rows.Err()) {
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
			ctx.Errorf("Failed to scan doc: %s", err)
		}
		var value int
		if err := rows.ScanValue(&value); err != nil {
			ctx.Errorf("Failed to scan value: %s", err)
		}
		if value != doc.Index {
			ctx.Errorf("doc._rev = %d, but value = %d", doc.Index, value)
		}
		id, _ := rows.ID()
		if doc.ID != id {
			ctx.Errorf("doc._id = %s, but rows.ID = %s", doc.ID, id)
		}
		docIDs = append(docIDs, id)
	}
	meta, err := rows.Metadata()
	if err != nil {
		ctx.Fatalf("Failed to fetch row: %s", rows.Err())
	}
	if d := testy.DiffTextSlices(expected, docIDs); d != nil {
		ctx.Errorf("Unexpected document IDs returned:\n%s\n", d)
	}
	if expOffset != meta.Offset {
		ctx.Errorf("offset: Expected %d, got %d", expOffset, meta.Offset)
	}
	ctx.Run("UpdateSeq", func(ctx *kt.Context) {
		if meta.UpdateSeq == "" {
			ctx.Errorf("Expected updated sequence")
		}
	})
	if int64(len(expected)) != meta.TotalRows {
		ctx.Errorf("total rows: Expected %d, got %d", len(expected), meta.TotalRows)
	}
}
