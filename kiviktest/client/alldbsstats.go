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

	"github.com/google/go-cmp/cmp"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.Register("AllDBsStats", allDBsStats)
}

func allDBsStats(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		testAllDBsStats(ctx, ctx.Admin)
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		testAllDBsStats(ctx, ctx.NoAuth)
	})
}

func testAllDBsStats(ctx *kt.Context, client *kivik.Client) {
	stats, err := client.AllDBsStats(context.Background())
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	if wantStats, ok := ctx.Interface("expected").([]*kivik.DBStats); ok {
		if d := cmp.Diff(wantStats, stats); d != "" {
			ctx.Errorf("Unexpected database stats: %s", d)
		}
	} else {
		const wantResults = 3
		if len(stats) != wantResults {
			ctx.Errorf("Expected %d database stats, got %d", wantResults, len(stats))
		}
	}
}
