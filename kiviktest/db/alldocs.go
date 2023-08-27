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
	kt.Register("AllDocs", allDocs)
}

func allDocs(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		testAllDocs(ctx, ctx.Admin)
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		testAllDocs(ctx, ctx.NoAuth)
	})
	ctx.RunRW(func(ctx *kt.Context) {
		testAllDocsRW(ctx)
	})
}

func testAllDocsRW(ctx *kt.Context) {
	if ctx.Admin == nil {
		// Can't do anything here without admin access
		return
	}
	dbName, expected, err := setUpAllDocsTest(ctx)
	if err != nil {
		ctx.Errorf("Failed to set up temp db: %s", err)
	}
	defer ctx.DestroyDB(dbName)
	ctx.Run("group", func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			doTest(ctx, ctx.Admin, dbName, 0, expected, true)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			doTest(ctx, ctx.NoAuth, dbName, 0, expected, true)
		})
	})
}

func setUpAllDocsTest(ctx *kt.Context) (dbName string, docIDs []string, err error) {
	dbName = ctx.TestDB()
	db := ctx.Admin.DB(dbName, ctx.Options("db"))
	if err := db.Err(); err != nil {
		return dbName, nil, errors.Wrap(err, "failed to connect to db")
	}
	docIDs = make([]string, 10)
	for i := range docIDs {
		id := ctx.TestDBName()
		doc := struct {
			ID string `json:"id"`
		}{
			ID: id,
		}
		if _, err := db.Put(context.Background(), doc.ID, doc); err != nil {
			return dbName, nil, errors.Wrap(err, "failed to create doc")
		}
		docIDs[i] = id
	}
	sort.Strings(docIDs)
	return dbName, docIDs, nil
}

func testAllDocs(ctx *kt.Context, client *kivik.Client) {
	if !ctx.IsSet("databases") {
		ctx.Errorf("databases not set; Did you configure this test?")
		return
	}
	for _, dbName := range ctx.StringSlice("databases") {
		func(dbName string) {
			ctx.Run(dbName, func(ctx *kt.Context) {
				doTest(ctx, client, dbName, int64(ctx.Int("offset")), ctx.StringSlice("expected"), false)
			})
		}(dbName)
	}
}

func doTest(ctx *kt.Context, client *kivik.Client, dbName string, expOffset int64, expected []string, exact bool) {
	ctx.Run("WithDocs", func(ctx *kt.Context) {
		doTestWithDocs(ctx, client, dbName, expOffset, expected, exact)
	})
	ctx.Run("WithoutDocs", func(ctx *kt.Context) {
		doTestWithoutDocs(ctx, client, dbName, expOffset, expected, exact)
	})
}

func doTestWithoutDocs(ctx *kt.Context, client *kivik.Client, dbName string, expOffset int64, expected []string, exact bool) {
	ctx.Parallel()
	db := client.DB(dbName, ctx.Options("db"))
	// Errors may be deferred here, so only return if we actually get
	// an error.
	if err := db.Err(); err != nil && !ctx.IsExpectedSuccess(err) {
		return
	}

	rows := db.AllDocs(context.Background())
	if !ctx.IsExpectedSuccess(rows.Err()) {
		return
	}
	docIDs := make([]string, 0, len(expected))
	for rows.Next() {
		id, _ := rows.ID()
		docIDs = append(docIDs, id)
	}
	meta, err := rows.Metadata()
	if err != nil {
		ctx.Fatalf("Failed to fetch row: %s", rows.Err())
	}
	testExpectedDocs(ctx, expected, docIDs, exact)
	if expOffset != meta.Offset {
		ctx.Errorf("offset: Expected %d, got %d", expOffset, meta.Offset)
	}
	if exact {
		if int64(len(expected)) != meta.TotalRows {
			ctx.Errorf("total rows: Expected %d, got %d", len(expected), meta.TotalRows)
		}
	}
}

func doTestWithDocs(ctx *kt.Context, client *kivik.Client, dbName string, expOffset int64, expected []string, exact bool) {
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

	rows := db.AllDocs(context.Background(), opts)
	if !ctx.IsExpectedSuccess(rows.Err()) {
		return
	}
	docIDs := make([]string, 0, len(expected))
	for rows.Next() {
		var doc struct {
			ID  string `json:"_id"`
			Rev string `json:"_rev"`
		}
		if err := rows.ScanDoc(&doc); err != nil {
			ctx.Errorf("Failed to scan doc: %s", err)
		}
		var value struct {
			Rev string `json:"rev"`
		}
		if err := rows.ScanValue(&value); err != nil {
			ctx.Errorf("Failed to scan value: %s", err)
		}
		if value.Rev != doc.Rev {
			ctx.Errorf("doc._rev = %s, but value.rev = %s", doc.Rev, value.Rev)
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
	testExpectedDocs(ctx, expected, docIDs, exact)
	if expOffset != meta.Offset {
		ctx.Errorf("offset: Expected %d, got %d", expOffset, meta.Offset)
	}
	ctx.Run("UpdateSeq", func(ctx *kt.Context) {
		if meta.UpdateSeq == "" {
			ctx.Errorf("Expected updated sequence")
		}
	})
	if exact {
		if int64(len(expected)) != meta.TotalRows {
			ctx.Errorf("total rows: Expected %d, got %d", len(expected), meta.TotalRows)
		}
	}
}

func testExpectedDocs(ctx *kt.Context, expected, actual []string, exact bool) {
	if exact {
		if d := testy.DiffTextSlices(expected, actual); d != nil {
			ctx.Errorf("Unexpected document IDs returned:\n%s\n", d)
		}
		return
	}
	actualIDs := make(map[string]struct{})
	for _, id := range actual {
		actualIDs[id] = struct{}{}
	}
	for _, id := range expected {
		if _, ok := actualIDs[id]; !ok {
			ctx.Errorf("Expected document '%s' not found", id)
		}
	}
}
