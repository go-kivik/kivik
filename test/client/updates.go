package client

import (
	"context"
	"fmt"
	"time"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
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
	updates, err := client.DBUpdates()
	if !ctx.IsExpectedSuccess(err) {
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
				} else {
					eventErrors <- fmt.Errorf("Unexpected event type '%s'", updates.Type())
				}
			}
		}
		eventErrors <- updates.Err()
		close(eventErrors)
	}()
	defer ctx.Admin.DestroyDB(context.Background(), dbname, ctx.Options("db"))
	if err = ctx.Admin.CreateDB(context.Background(), dbname, ctx.Options("db")); err != nil {
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
