package client

import (
	"context"
	"strings"
	"time"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("Replicate", replicate)
}

func replicate(ctx *kt.Context) {
	defer lockReplication(ctx)()
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			testReplication(ctx, ctx.Admin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			testReplication(ctx, ctx.NoAuth)
		})
	})
}

func callReplicate(ctx *kt.Context, client *kivik.Client, target, source, repID string, opts kivik.Options) (*kivik.Replication, error) {
	opts = replicationOptions(ctx, client, target, source, repID, opts)
	return client.Replicate(context.Background(), target, source, opts)
}

func testReplication(ctx *kt.Context, client *kivik.Client) {
	prefix := ctx.String("prefix")
	switch prefix {
	case "":
		prefix = strings.TrimSuffix(client.DSN(), "/") + "/"
	case "none":
		prefix = ""
	}
	dbtarget := prefix + ctx.TestDB()
	dbsource := prefix + ctx.TestDB()
	defer ctx.Admin.DestroyDB(context.Background(), dbtarget, ctx.Options("db"))
	defer ctx.Admin.DestroyDB(context.Background(), dbsource, ctx.Options("db"))

	db, err := ctx.Admin.DB(context.Background(), dbsource)
	if err != nil {
		ctx.Fatalf("Failed to open db: %s", err)
	}

	// Create 10 docs for testing sync
	for i := 0; i < 10; i++ {
		id := ctx.TestDBName()
		doc := struct {
			ID string `json:"id"`
		}{
			ID: id,
		}
		if _, err := db.Put(context.Background(), doc.ID, doc); err != nil {
			ctx.Fatalf("Failed to create doc: %s", err)
		}
	}

	ctx.Run("group", func(ctx *kt.Context) {
		ctx.Run("ValidReplication", func(ctx *kt.Context) {
			ctx.Parallel()
			tries := 3
			success := false
			for i := 0; i < tries; i++ {
				success = doReplicationTest(ctx, client, dbtarget, dbsource)
				if success {
					break
				}
			}
			if !success {
				ctx.Errorf("Replication failied after %d tries", tries)
			}
		})
		ctx.Run("MissingSource", func(ctx *kt.Context) {
			ctx.Parallel()
			doReplicationTest(ctx, client, dbtarget, ctx.MustString("NotFoundDB"))
		})
		ctx.Run("MissingTarget", func(ctx *kt.Context) {
			ctx.Parallel()
			doReplicationTest(ctx, client, ctx.MustString("NotFoundDB"), dbsource)
		})
		ctx.Run("Cancel", func(ctx *kt.Context) {
			ctx.Parallel()
			replID := ctx.TestDBName()
			rep, err := callReplicate(ctx, client, dbtarget, "http://foo:foo@192.168.2.254/foo", replID, kivik.Options{"continuous": true})
			if !ctx.IsExpectedSuccess(err) {
				return
			}
			defer rep.Delete(context.Background())
			timeout := time.Duration(ctx.MustInt("timeoutSeconds")) * time.Second
			cx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			ctx.CheckError(rep.Delete(context.Background()))
			for rep.IsActive() {
				if rep.State() == kivik.ReplicationStarted {
					return
				}
				select {
				case <-cx.Done():
					break
				default:
				}
				if err := rep.Update(cx); err != nil {
					if kivik.StatusCode(err) == kivik.StatusNotFound {
						// NotFound expected after the replication is cancelled
						break
					}
					ctx.Fatalf("Failed to read update: %s", err)
					break
				}
			}
			if err := cx.Err(); err != nil {
				ctx.Fatalf("context was cancelled: %s", err)
			}
			if err := rep.Err(); err != nil {
				ctx.Fatalf("Replication cancellation failed: %s", err)
			}
		})
	})
}

func doReplicationTest(ctx *kt.Context, client *kivik.Client, dbtarget, dbsource string) (success bool) {
	success = true
	replID := ctx.TestDBName()
	rep, err := callReplicate(ctx, client, dbtarget, dbsource, replID, nil)
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	defer rep.Delete(context.Background())
	timeout := time.Duration(ctx.MustInt("timeoutSeconds")) * time.Second
	cx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var updateErr error
	for rep.IsActive() {
		select {
		case <-cx.Done():
			ctx.Fatalf("context cancelled waiting for replication: %s", cx.Err())
			return
		default:
		}
		if updateErr = rep.Update(cx); updateErr != nil {
			break
		}
	}
	if updateErr != nil {
		ctx.Fatalf("Replication update failed: %s", updateErr)
	}
	ctx.Run("Results", func(ctx *kt.Context) {
		err := rep.Err()
		if kivik.StatusCode(err) == kivik.StatusRequestTimeout {
			success = false // Allow retrying
			return
		}
		if !ctx.IsExpectedSuccess(err) {
			return
		}
		switch ctx.String("mode") {
		case "pouchdb":
			if rep.ReplicationID() != "" {
				ctx.Errorf("Did not expect replication ID")
			}
		default:
			if rep.ReplicationID() == "" {
				ctx.Errorf("Expected a replication ID")
			}
		}
		if rep.Source != dbsource {
			ctx.Errorf("Unexpected source. Expected: %s, Actual: %s\n", dbsource, rep.Source)
		}
		if rep.Target != dbtarget {
			ctx.Errorf("Unexpected target. Expected: %s, Actual: %s\n", dbtarget, rep.Target)
		}
		if rep.State() != kivik.ReplicationComplete {
			ctx.Errorf("Replication failed to complete. Final state: %s\n", rep.State())
		}
		if (rep.Progress() - float64(100)) > 0.0001 {
			ctx.Errorf("Expected 100%% completion, got %%%02.2f", rep.Progress())
		}
	})
	return success
}
