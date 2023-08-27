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
	kt.Register("CreateDoc", createDoc)
}

func createDoc(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		dbname := ctx.TestDB()
		defer ctx.DestroyDB(dbname)
		ctx.Run("group", func(ctx *kt.Context) {
			ctx.RunAdmin(func(ctx *kt.Context) {
				ctx.Parallel()
				testCreate(ctx, ctx.Admin, dbname)
			})
			ctx.RunNoAuth(func(ctx *kt.Context) {
				ctx.Parallel()
				testCreate(ctx, ctx.NoAuth, dbname)
			})
		})
	})
}

func testCreate(ctx *kt.Context, client *kivik.Client, dbname string) {
	db := client.DB(dbname, ctx.Options("db"))
	if err := db.Err(); err != nil {
		ctx.Fatalf("Failed to connect to database: %s", err)
	}
	ctx.Run("WithoutID", func(ctx *kt.Context) {
		ctx.Parallel()
		err := kt.Retry(func() error {
			_, _, err := db.CreateDoc(context.Background(), map[string]string{"foo": "bar"})
			return err
		})
		ctx.CheckError(err)
	})
	ctx.Run("WithID", func(ctx *kt.Context) {
		ctx.Parallel()
		id := ctx.TestDBName()
		var docID string
		err := kt.Retry(func() error {
			var err error
			docID, _, err = db.CreateDoc(context.Background(), map[string]string{"foo": "bar", "_id": id})
			return err
		})
		if !ctx.IsExpectedSuccess(err) {
			return
		}
		if id != docID {
			ctx.Errorf("CreateDoc didn't honor provided ID. Expected '%s', Got '%s'", id, docID)
		}
	})
}
