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

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kiviktest/v4/kt"
)

func init() {
	kt.Register("GetRev", getRev)
}

func getRev(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		dbName := ctx.TestDB()
		defer ctx.DestroyDB(dbName)
		db := ctx.Admin.DB(dbName, ctx.Options("db"))
		if err := db.Err(); err != nil {
			ctx.Fatalf("Failed to connect to test db: %s", err)
		}
		doc := &testDoc{
			ID: "bob",
		}
		rev, err := db.Put(context.Background(), doc.ID, doc)
		if err != nil {
			ctx.Fatalf("Failed to create doc in test db: %s", err)
		}
		doc.Rev = rev

		ddoc := &testDoc{
			ID: "_design/foo",
		}
		rev, err = db.Put(context.Background(), ddoc.ID, ddoc)
		if err != nil {
			ctx.Fatalf("Failed to create design doc in test db: %s", err)
		}
		ddoc.Rev = rev

		local := &testDoc{
			ID: "_local/foo",
		}
		rev, err = db.Put(context.Background(), local.ID, local)
		if err != nil {
			ctx.Fatalf("Failed to create local doc in test db: %s", err)
		}
		local.Rev = rev

		ctx.Run("group", func(ctx *kt.Context) {
			ctx.RunAdmin(func(ctx *kt.Context) {
				ctx.Parallel()
				db := ctx.Admin.DB(dbName, ctx.Options("db"))
				if err := db.Err(); !ctx.IsExpectedSuccess(err) {
					return
				}
				testGetRev(ctx, db, doc)
				testGetRev(ctx, db, ddoc)
				testGetRev(ctx, db, local)
				testGetRev(ctx, db, &testDoc{ID: "bogus"})
			})
			ctx.RunNoAuth(func(ctx *kt.Context) {
				ctx.Parallel()
				db := ctx.NoAuth.DB(dbName, ctx.Options("db"))
				if err := db.Err(); !ctx.IsExpectedSuccess(err) {
					return
				}
				testGetRev(ctx, db, doc)
				testGetRev(ctx, db, ddoc)
				testGetRev(ctx, db, local)
				testGetRev(ctx, db, &testDoc{ID: "bogus"})
			})
		})
	})
}

func testGetRev(ctx *kt.Context, db *kivik.DB, expectedDoc *testDoc) {
	ctx.Run(expectedDoc.ID, func(ctx *kt.Context) {
		ctx.Parallel()
		rev, err := db.GetRev(context.Background(), expectedDoc.ID)
		if !ctx.IsExpectedSuccess(err) {
			return
		}
		doc := &testDoc{}
		if err = db.Get(context.Background(), expectedDoc.ID).ScanDoc(&doc); err != nil {
			ctx.Fatalf("Failed to scan doc: %s\n", err)
		}
		if strings.HasPrefix(expectedDoc.ID, "_local/") {
			// Revisions are meaningless for _local docs
			return
		}
		if rev != doc.Rev {
			ctx.Errorf("Unexpected rev. Expected: %s, Actual: %s", doc.Rev, rev)
		}
	})
}
