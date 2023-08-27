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

package client

import (
	"context"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kiviktest/v4/kt"
)

func init() {
	kt.Register("DBExists", dbExists)
}

func dbExists(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		for _, dbName := range ctx.MustStringSlice("databases") {
			checkDBExists(ctx, ctx.Admin, dbName)
		}
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		for _, dbName := range ctx.MustStringSlice("databases") {
			checkDBExists(ctx, ctx.NoAuth, dbName)
		}
	})
	ctx.RunRW(func(ctx *kt.Context) {
		dbName := ctx.TestDB()
		defer ctx.DestroyDB(dbName)
		ctx.Run("group", func(ctx *kt.Context) {
			ctx.RunAdmin(func(ctx *kt.Context) {
				checkDBExists(ctx, ctx.Admin, dbName)
			})
			ctx.RunNoAuth(func(ctx *kt.Context) {
				checkDBExists(ctx, ctx.NoAuth, dbName)
			})
		})
	})
}

func checkDBExists(ctx *kt.Context, client *kivik.Client, dbName string) {
	ctx.Run(dbName, func(ctx *kt.Context) {
		ctx.Parallel()
		exists, err := client.DBExists(context.Background(), dbName)
		if !ctx.IsExpectedSuccess(err) {
			return
		}
		if ctx.MustBool("exists") != exists {
			ctx.Errorf("Expected: %t, actual: %t", ctx.Bool("exists"), exists)
		}
	})
}
