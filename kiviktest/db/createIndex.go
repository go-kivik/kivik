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
	kt.Register("CreateIndex", createIndex)
}

func createIndex(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			ctx.Parallel()
			testCreateIndex(ctx, ctx.Admin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			ctx.Parallel()
			testCreateIndex(ctx, ctx.NoAuth)
		})
	})
}

func testCreateIndex(ctx *kt.Context, client *kivik.Client) {
	dbname := ctx.TestDB()
	defer ctx.DestroyDB(dbname)
	db := client.DB(dbname, ctx.Options("db"))
	if err := db.Err(); err != nil {
		ctx.Fatalf("Failed to open db: %s", err)
	}
	ctx.Run("group", func(ctx *kt.Context) {
		ctx.Run("Valid", func(ctx *kt.Context) {
			doCreateIndex(ctx, db, `{"fields":["foo"]}`)
		})
		ctx.Run("NilIndex", func(ctx *kt.Context) {
			doCreateIndex(ctx, db, nil)
		})
		ctx.Run("BlankIndex", func(ctx *kt.Context) {
			doCreateIndex(ctx, db, "")
		})
		ctx.Run("EmptyIndex", func(ctx *kt.Context) {
			doCreateIndex(ctx, db, "{}")
		})
		ctx.Run("InvalidIndex", func(ctx *kt.Context) {
			doCreateIndex(ctx, db, `{"oink":true}`)
		})
		ctx.Run("InvalidJSON", func(ctx *kt.Context) {
			doCreateIndex(ctx, db, `chicken`)
		})
	})
}

func doCreateIndex(ctx *kt.Context, db *kivik.DB, index interface{}) {
	ctx.Parallel()
	err := kt.Retry(func() error {
		return db.CreateIndex(context.Background(), "", "", index)
	})
	ctx.CheckError(err)
}
