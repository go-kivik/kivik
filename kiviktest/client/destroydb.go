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
	kt.Register("DestroyDB", destroyDB)
}

func destroyDB(ctx *kt.Context) {
	// All DestroyDB tests are RW by nature.
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			ctx.Parallel()
			testDestroy(ctx, ctx.Admin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			ctx.Parallel()
			testDestroy(ctx, ctx.NoAuth)
		})
	})
}

func testDestroy(ctx *kt.Context, client *kivik.Client) {
	ctx.Run("ExistingDB", func(ctx *kt.Context) {
		ctx.Parallel()
		dbName := ctx.TestDB()
		defer ctx.DestroyDB(dbName)
		ctx.CheckError(client.DestroyDB(context.Background(), dbName, ctx.Options("db")))
	})
	ctx.Run("NonExistantDB", func(ctx *kt.Context) {
		ctx.Parallel()
		ctx.CheckError(client.DestroyDB(context.Background(), ctx.TestDBName(), ctx.Options("db")))
	})
}
