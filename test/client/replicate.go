package client

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver/pouchdb/bindings"
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

func replicationOptions(ctx *kt.Context, client *kivik.Client, target, source, repID string, in map[string]interface{}) map[string]interface{} {
	if in == nil {
		in = make(map[string]interface{})
	}
	if ctx.String("mode") != "pouchdb" {
		in["_id"] = repID
		return in
	}
	in["source"] = bindings.GlobalPouchDB().New(source, ctx.Options("db"))
	in["target"] = bindings.GlobalPouchDB().New(target, ctx.Options("db"))
	return in
}

func testReplication(ctx *kt.Context, client *kivik.Client) {
	prefix := ctx.String("prefix")
	switch prefix {
	case "":
		prefix = strings.TrimSuffix(client.DSN(), "/") + "/"
	case "none":
		prefix = ""
	}
	dbname1 := prefix + ctx.TestDB()
	dbname2 := prefix + ctx.TestDB()
	defer ctx.Admin.DestroyDB(context.Background(), dbname1, ctx.Options("db"))
	defer ctx.Admin.DestroyDB(context.Background(), dbname2, ctx.Options("db"))
	ctx.Run("group", func(ctx *kt.Context) {
		ctx.Run("ValidReplication", func(ctx *kt.Context) {
			ctx.Parallel()
			replID := ctx.TestDBName()
			rep, err := callReplicate(ctx, client, dbname1, dbname2, replID, nil)
			if !ctx.IsExpectedSuccess(err) {
				return
			}
			rep.SetRateLimiter(kivik.ConstantRateLimiter(250 * time.Millisecond))
			defer rep.Delete(context.Background())
			done := make(chan struct{})
			cx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var updateErr error
			go func() {
				for rep.IsActive() {
					select {
					case <-cx.Done():
						return
					default:
					}
					if updateErr = rep.Update(cx); updateErr != nil {
						break
					}
				}
				done <- struct{}{}
			}()
			timeout := time.Duration(ctx.MustInt("timeoutSeconds")) * time.Second
			select {
			case <-time.After(timeout):
				ctx.Fatalf("Replication failed to complete in %s", timeout)
			case <-done:
				if updateErr != nil {
					ctx.Fatalf("Replication update failed: %s", updateErr)
				}
			}
			ctx.Run("ReplicationResults", func(ctx *kt.Context) {
				if !ctx.IsExpectedSuccess(rep.Err()) {
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
				if rep.Source != dbname2 {
					ctx.Errorf("Unexpected source. Expected: %s, Actual: %s\n", dbname2, rep.Source)
				}
				if rep.Target != dbname1 {
					ctx.Errorf("Unexpected target. Expected: %s, Actual: %s\n", dbname1, rep.Target)
				}
				if rep.State() != kivik.ReplicationComplete {
					ctx.Errorf("Replication failed to complete. Final state: %s\n", rep.State())
				}
			})
		})
		ctx.Run("Cancel", func(ctx *kt.Context) {
			ctx.Parallel()
			dbnameA := prefix + ctx.TestDB()
			dbnameB := prefix + ctx.TestDB()
			defer ctx.Admin.DestroyDB(context.Background(), dbnameA, ctx.Options("db"))
			defer ctx.Admin.DestroyDB(context.Background(), dbnameB, ctx.Options("db"))
			replID := ctx.TestDBName()
			rep, err := callReplicate(ctx, client, dbnameA, dbnameB, replID, kivik.Options{"continuous": true})
			if !ctx.IsExpectedSuccess(err) {
				return
			}
			defer rep.Delete(context.Background())
			done := make(chan struct{})
			cx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var updateErr error
			go func() {
				defer func() { done <- struct{}{} }()
				for rep.IsActive() {
					if rep.State() == kivik.ReplicationStarted {
						return
					}
					select {
					case <-cx.Done():
						updateErr = errors.New("replication completed normally")
					default:
					}
					if updateErr = rep.Update(cx); updateErr != nil {
						return
					}
				}
			}()
			timeout := time.Duration(ctx.MustInt("timeoutSeconds")) * time.Second
			select {
			case <-time.After(timeout):
				ctx.Fatalf("Replication failed to complete in %s", timeout)
			case <-done:
				if updateErr != nil {
					ctx.Fatalf("Replication cancellation failed: %s", updateErr)
				}
			}
			ctx.CheckError(rep.Delete(context.Background()))
		})
		ctx.Run("MissingSource", func(ctx *kt.Context) {
			ctx.Parallel()
			replID := ctx.TestDBName()
			rep, err := callReplicate(ctx, client, dbname1, "http://localhost:5984/foo", replID, nil)
			if !ctx.IsExpectedSuccess(err) {
				return
			}
			rep.Delete(context.Background())
		})
		ctx.Run("MissingTarget", func(ctx *kt.Context) {
			ctx.Parallel()
			replID := ctx.TestDBName()
			rep, err := callReplicate(ctx, client, "http://localhost:5984/foo", dbname2, replID, nil)
			if !ctx.IsExpectedSuccess(err) {
				return
			}
			rep.Delete(context.Background())
		})
	})
}
