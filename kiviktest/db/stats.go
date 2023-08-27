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

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kiviktest/v4/kt"
)

func init() {
	kt.Register("Stats", stats)
}

func stats(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		ctx.Parallel()
		roTests(ctx, ctx.Admin)
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		ctx.Parallel()
		roTests(ctx, ctx.NoAuth)
	})
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.Parallel()
		ctx.RunAdmin(func(ctx *kt.Context) {
			ctx.Parallel()
			rwTests(ctx, ctx.Admin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			ctx.Parallel()
			rwTests(ctx, ctx.NoAuth)
		})
	})
}

func rwTests(ctx *kt.Context, client *kivik.Client) {
	dbname := ctx.TestDB()
	defer ctx.DestroyDB(dbname)
	db := ctx.Admin.DB(dbname, ctx.Options("db"))
	if err := db.Err(); err != nil {
		ctx.Fatalf("Failed to connect to db: %s", err)
	}
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("%d", i)
		rev, err := db.Put(context.Background(), id, struct{}{})
		if err != nil {
			ctx.Fatalf("Failed to create document ID %s: %s", id, err)
		}
		if i > 5 {
			if _, err = db.Delete(context.Background(), id, rev); err != nil {
				ctx.Fatalf("Failed to delete document ID %s: %s", id, err)
			}
		}
	}
	testDBInfo(ctx, client, dbname, 6)
}

func roTests(ctx *kt.Context, client *kivik.Client) {
	for _, dbname := range ctx.MustStringSlice("databases") {
		func(dbname string) {
			ctx.Run(dbname, func(ctx *kt.Context) {
				ctx.Parallel()
				testDBInfo(ctx, client, dbname, 0)
			})
		}(dbname)
	}
}

func testDBInfo(ctx *kt.Context, client *kivik.Client, dbname string, docCount int64) {
	stats, err := client.DB(dbname, ctx.Options("db")).Stats(context.Background())
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	if stats.Name != dbname {
		ctx.Errorf("Name: Expected '%s', actual '%s'", dbname, stats.Name)
	}
	if docCount > 0 {
		if docCount != stats.DocCount {
			ctx.Errorf("DocCount: Expected %d, actual %d", docCount, stats.DocCount)
		}
	}
}
