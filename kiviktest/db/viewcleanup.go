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
	kt.Register("ViewCleanup", viewCleanup)
}

func viewCleanup(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			ctx.Parallel()
			testViewCleanup(ctx, ctx.Admin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			ctx.Parallel()
			testViewCleanup(ctx, ctx.NoAuth)
		})
	})
}

func testViewCleanup(ctx *kt.Context, client *kivik.Client) {
	dbname := ctx.TestDB()
	defer ctx.DestroyDB(dbname)
	db := client.DB(dbname, ctx.Options("db"))
	if err := db.Err(); err != nil {
		ctx.Fatalf("Failed to connect to db: %s", err)
	}
	ctx.CheckError(db.ViewCleanup(context.Background()))
}
