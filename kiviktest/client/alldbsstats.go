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
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.Register("AllDBsStats", allDBsStats)
}

func allDBsStats(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		testAllDBsStats(ctx, ctx.Admin, ctx.Admin)
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		testAllDBsStats(ctx, ctx.Admin, ctx.NoAuth)
	})
}

func testAllDBsStats(ctx *kt.Context, admin, client *kivik.Client) {
	// create a db
	dbName := ctx.TestDBName()
	if err := admin.CreateDB(context.Background(), dbName); err != nil {
		ctx.Fatalf("Failed to create test database %s: %s", dbName, err)
	}
	ctx.T.Cleanup(func() { ctx.DestroyDB(dbName) })
	stats, err := client.AllDBsStats(context.Background())
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	for _, stat := range stats {
		if stat.Name == dbName {
			return
		}
	}
	ctx.Errorf("Expected database %s to be in stats, but it was not found", dbName)
}
