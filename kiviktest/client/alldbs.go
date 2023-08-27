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
	"sort"

	"gitlab.com/flimzy/testy"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kiviktest/v4/kt"
)

func init() {
	kt.Register("AllDBs", allDBs)
}

func allDBs(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		testAllDBs(ctx, ctx.Admin, ctx.StringSlice("expected"))
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		testAllDBs(ctx, ctx.NoAuth, ctx.StringSlice("expected"))
	})
	if ctx.RW && ctx.Admin != nil {
		ctx.Run("RW", func(ctx *kt.Context) {
			testAllDBsRW(ctx)
		})
	}
}

func testAllDBsRW(ctx *kt.Context) {
	dbName := ctx.TestDB()
	defer ctx.DestroyDB(dbName)
	expected := append(ctx.StringSlice("expected"), dbName)
	ctx.Run("group", func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			testAllDBs(ctx, ctx.Admin, expected)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			testAllDBs(ctx, ctx.NoAuth, expected)
		})
	})
}

func testAllDBs(ctx *kt.Context, client *kivik.Client, expected []string) {
	ctx.Parallel()
	allDBs, err := client.AllDBs(context.Background())
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	sort.Strings(expected)
	sort.Strings(allDBs)
	if d := testy.DiffTextSlices(expected, allDBs); d != nil {
		ctx.Errorf("AllDBs() returned unexpected list:\n%s\n", d)
	}
}
