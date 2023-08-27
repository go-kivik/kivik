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
	kt.Register("Find", find)
}

func find(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		testFind(ctx, ctx.Admin)
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		testFind(ctx, ctx.NoAuth)
	})
	ctx.RunRW(func(ctx *kt.Context) {
		testFindRW(ctx)
	})
}

func testFindRW(ctx *kt.Context) {
	if ctx.Admin == nil {
		// Can't do anything here without admin access
		return
	}
	dbName, expected, err := setUpFindTest(ctx)
	if err != nil {
		ctx.Errorf("Failed to set up temp db: %s", err)
	}
	defer ctx.DestroyDB(dbName)
	ctx.Run("group", func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			doFindTest(ctx, ctx.Admin, dbName, 0, expected)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			doFindTest(ctx, ctx.NoAuth, dbName, 0, expected)
		})
	})
}

func setUpFindTest(ctx *kt.Context) (dbName string, docIDs []string, err error) {
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

func testFind(ctx *kt.Context, client *kivik.Client) {
	if !ctx.IsSet("databases") {
		ctx.Errorf("databases not set; Did you configure this test?")
		return
	}
	for _, dbName := range ctx.StringSlice("databases") {
		func(dbName string) {
			ctx.Run(dbName, func(ctx *kt.Context) {
				doFindTest(ctx, client, dbName, int64(ctx.Int("offset")), ctx.StringSlice("expected"))
			})
		}(dbName)
	}
}

func doFindTest(ctx *kt.Context, client *kivik.Client, dbName string, expOffset int64, expected []string) {
	ctx.Parallel()
	db := client.DB(dbName, ctx.Options("db"))
	// Errors may be deferred here, so only return if we actually get
	// an error.
	if err := db.Err(); err != nil && !ctx.IsExpectedSuccess(err) {
		return
	}

	var rows *kivik.ResultSet
	err := kt.Retry(func() error {
		rows = db.Find(context.Background(), `{"selector":{"_id":{"$gt":null}}}`)
		return rows.Err()
	})

	if !ctx.IsExpectedSuccess(err) {
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
			ctx.Errorf("Failed to scan doc: %s", err)
		}
		docIDs = append(docIDs, doc.DocID)
	}
	meta, err := rows.Metadata()
	if err != nil {
		ctx.Fatalf("Failed to fetch row: %s", rows.Err())
	}
	sort.Strings(docIDs) // normalize order
	if d := testy.DiffTextSlices(expected, docIDs); d != nil {
		ctx.Errorf("Unexpected document IDs returned:\n%s\n", d)
	}
	if meta.Offset != expOffset {
		ctx.Errorf("Unexpected offset: %v", meta.Offset)
	}
	ctx.Run("Warning", func(ctx *kt.Context) {
		rows := db.Find(context.Background(), `{"selector":{"foo":{"$gt":null}}}`)
		if !ctx.IsExpectedSuccess(rows.Err()) {
			return
		}
		for rows.Next() {
		}
		if w := ctx.String("warning"); w != meta.Warning {
			ctx.Errorf("Warning:\nExpected: %s\n  Actual: %s", w, meta.Warning)
		}
	})
}
