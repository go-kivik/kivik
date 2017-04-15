package client

import (
	"context"
	"time"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("Replicate", replicate)
}

func replicate(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			testReplication(ctx, ctx.Admin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			testReplication(ctx, ctx.NoAuth)
		})
	})
}

const replicationTimeLimit = 10 * time.Second

func testReplication(ctx *kt.Context, client *kivik.Client) {
	dbname1 := ctx.TestDBName()
	dbname2 := ctx.TestDBName()
	defer ctx.Admin.DestroyDB(dbname1)
	defer ctx.Admin.DestroyDB(dbname2)
	if err := ctx.Admin.CreateDB(dbname1); err != nil {
		ctx.Fatalf("Failed to create db: %s", err)
	}
	if err := ctx.Admin.CreateDB(dbname2); err != nil {
		ctx.Fatalf("Failed to create db: %s", err)
	}
	ctx.Run("group", func(ctx *kt.Context) {
		ctx.Run("ValidReplication", func(ctx *kt.Context) {
			ctx.Parallel()
			replID := ctx.TestDBName()
			rep, err := client.Replicate(kt.CTX, dbname1, dbname2, kivik.Options{"_id": replID})
			if !ctx.IsExpectedSuccess(err) {
				return
			}
			defer rep.Delete(kt.CTX)
			done := make(chan struct{})
			cx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				for rep.Active() {
					err = rep.Update(cx)
					time.Sleep(100 * time.Millisecond)
				}
				done <- struct{}{}
			}()
			select {
			case <-time.After(replicationTimeLimit):
				ctx.Errorf("Replication failed to complete in %s", replicationTimeLimit)
			case <-done:
				if err != nil {
					ctx.Errorf("Replication update failed: %s", err)
				}
			}
			if !ctx.IsExpectedSuccess(rep.Err()) {
				return
			}
			if rep.ReplicationID == "" {
				ctx.Errorf("Expected a replication ID")
			}
			if rep.Source != dbname2 {
				ctx.Errorf("Unexpected source. Expected: %s, Actual: %s\n", dbname2, rep.Source)
			}
			if rep.Target != dbname1 {
				ctx.Errorf("Unexpected target. Expected: %s, Actual: %s\n", dbname1, rep.Target)
			}
		})
		ctx.Run("Cancel", func(ctx *kt.Context) {
			ctx.Parallel()
			dbnameA := ctx.TestDBName()
			dbnameB := ctx.TestDBName()
			defer ctx.Admin.DestroyDB(dbnameA)
			defer ctx.Admin.DestroyDB(dbnameB)
			if err := ctx.Admin.CreateDB(dbnameA); err != nil {
				ctx.Fatalf("Failed to create db: %s", err)
			}
			if err := ctx.Admin.CreateDB(dbnameB); err != nil {
				ctx.Fatalf("Failed to create db: %s", err)
			}
			replID := ctx.TestDBName()
			rep, err := client.Replicate(kt.CTX, dbnameA, dbnameB, kivik.Options{"_id": replID})
			if !ctx.IsExpectedSuccess(err) {
				return
			}
			defer rep.Delete(kt.CTX)
			ctx.CheckError(rep.Cancel(kt.CTX))
		})
		ctx.Run("MissingSource", func(ctx *kt.Context) {
			ctx.Parallel()
			replID := ctx.TestDBName()
			rep, err := client.Replicate(kt.CTX, dbname1, "foo", kivik.Options{"_id": replID})
			if !ctx.IsExpectedSuccess(err) {
				return
			}
			rep.Delete(kt.CTX)
		})
		ctx.Run("MissingTarget", func(ctx *kt.Context) {
			ctx.Parallel()
			replID := ctx.TestDBName()
			rep, err := client.Replicate(kt.CTX, "foo", dbname2, kivik.Options{"_id": replID})
			if !ctx.IsExpectedSuccess(err) {
				return
			}
			rep.Delete(kt.CTX)
		})
	})
}
