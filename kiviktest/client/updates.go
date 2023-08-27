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
	"fmt"
	"time"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kiviktest/v4/kt"
)

func init() {
	kt.Register("DBUpdates", updates)
}

func updates(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			testUpdates(ctx, ctx.Admin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			testUpdates(ctx, ctx.NoAuth)
		})
	})
}

const maxWait = 5 * time.Second

func testUpdates(ctx *kt.Context, client *kivik.Client) {
	ctx.Parallel()
	updates := client.DBUpdates(context.TODO())
	if !ctx.IsExpectedSuccess(updates.Err()) {
		return
	}
	// It seems that DBUpdates doesn't always start responding immediately,
	// so introduce a small delay to ensure we're reporting updates before we
	// actually do the updates.
	time.Sleep(10 * time.Millisecond)
	dbname := ctx.TestDBName()
	eventErrors := make(chan error)
	go func() {
		for updates.Next() {
			if updates.DBName() == dbname {
				if updates.Type() == "created" {
					break
				}
				eventErrors <- fmt.Errorf("Unexpected event type '%s'", updates.Type())
			}
		}
		eventErrors <- updates.Err()
		close(eventErrors)
	}()
	defer ctx.DestroyDB(dbname)
	if err := ctx.Admin.CreateDB(context.Background(), dbname, ctx.Options("db")); err != nil {
		ctx.Fatalf("Failed to create db: %s", err)
	}
	timer := time.NewTimer(maxWait)
	select {
	case err := <-eventErrors:
		if err != nil {
			ctx.Fatalf("Error reading event: %s", err)
		}
	case <-timer.C:
		ctx.Fatalf("Failed to read expected event in %s", maxWait)
	}
	if err := updates.Close(); err != nil {
		ctx.Errorf("Updates close failed: %s", err)
	}
}
