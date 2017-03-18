package client

import (
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

type updateEvent struct {
	Update *kivik.DBUpdate
	Error  error
}

func testUpdates(ctx *kt.Context, client *kivik.Client) {
	ctx.Parallel()
	updates, err := client.DBUpdates()
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	events := make(chan *updateEvent)
	go func() {
		var event *kivik.DBUpdate
		for {
			event, err = updates.Next()
			events <- &updateEvent{
				Update: event,
				Error:  err,
			}
			if err != nil {
				close(events)
				break
			}
		}
	}()
	dbname := ctx.TestDBName()
	defer ctx.Admin.DestroyDB(dbname)
	if err = ctx.Admin.CreateDB(dbname); err != nil {
		ctx.Fatalf("Failed to create db: %s", err)
	}
	timer := time.NewTimer(maxWait)
Loop:
	for {
		select {
		case event := <-events:
			if event.Error != nil {
				ctx.Fatalf("Error reading event: %s", err)
			}
			if event.Update.DBName == dbname {
				if event.Update.Type != "created" {
					ctx.Errorf("Unexpected event type '%s'", event.Update.Type)
				}
				break Loop
			}
		case <-timer.C:
			ctx.Fatalf("Failed to read expected event in %s", maxWait)
		}
	}
	if err := updates.Close(); err != nil {
		ctx.Errorf("Updates close failed: %s", err)
	}
}
