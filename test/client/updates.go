package client

import (
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
	// Two instances to test concurrency
	updates2, err := client.DBUpdates()
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	dbname := ctx.TestDBName()
	readUpdates := func(updates *kivik.DBUpdateFeed, eventErrors chan<- error) {
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
	}
	eventErrors := make(chan error)
	eventErrors2 := make(chan error)
	go readUpdates(updates, eventErrors)
	go readUpdates(updates2, eventErrors2)
	defer ctx.Admin.DestroyDB(dbname)
	if err = ctx.Admin.CreateDB(dbname); err != nil {
		ctx.Fatalf("Failed to create db: %s", err)
	}
	timer := time.NewTimer(maxWait)
	select {
	case err := <-eventErrors:
		if err != nil {
			ctx.Fatalf("Error reading event: %s", err)
		}
	case err := <-eventErrors2:
		if err != nil {
			ctx.Fatalf("Error reading concurrent event: %s", err)
		}
	case <-timer.C:
		ctx.Fatalf("Failed to read expected event in %s", maxWait)
	}
	if err := updates.Close(); err != nil {
		ctx.Errorf("Updates close failed: %s", err)
	}
}
