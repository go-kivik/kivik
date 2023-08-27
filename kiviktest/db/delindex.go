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

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kiviktest/v4/kt"
)

func init() {
	kt.Register("DeleteIndex", delindex)
}

func delindex(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			ctx.Parallel()
			testDelIndex(ctx, ctx.Admin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			ctx.Parallel()
			testDelIndex(ctx, ctx.NoAuth)
		})
	})
}

func testDelIndex(ctx *kt.Context, client *kivik.Client) {
	dbname := ctx.TestDB()
	defer ctx.Admin.DestroyDB(context.Background(), dbname, ctx.Options("db")) // nolint: errcheck
	dba := ctx.Admin.DB(dbname, ctx.Options("db"))
	if err := dba.Err(); err != nil {
		ctx.Fatalf("Failed to open db as admin: %s", err)
	}
	if err := dba.CreateIndex(context.Background(), "foo", "bar", `{"fields":["foo"]}`); err != nil {
		ctx.Fatalf("Failed to create index: %s", err)
	}
	db := client.DB(dbname, ctx.Options("db"))
	if err := db.Err(); err != nil {
		ctx.Fatalf("Failed to open db: %s", err)
	}
	ctx.Run("group", func(ctx *kt.Context) {
		ctx.Run("ValidIndex", func(ctx *kt.Context) {
			ctx.Parallel()
			ctx.CheckError(db.DeleteIndex(context.Background(), "foo", "bar"))
		})
		ctx.Run("NotFoundDdoc", func(ctx *kt.Context) {
			ctx.Parallel()
			ctx.CheckError(db.DeleteIndex(context.Background(), "notFound", "bar"))
		})
		ctx.Run("NotFoundName", func(ctx *kt.Context) {
			ctx.Parallel()
			ctx.CheckError(db.DeleteIndex(context.Background(), "foo", "notFound"))
		})
	})
}
