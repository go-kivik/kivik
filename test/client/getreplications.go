package client

import (
	"sync"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
	"golang.org/x/net/context"
)

func init() {
	kt.Register("GetReplications", getReplications)
}

// masterMU protects the map
var masterMU sync.Mutex

// We can only run one set of replication tests at a time
var replicationMUs = make(map[*kivik.Client]*sync.Mutex)

func lockReplication(ctx *kt.Context) func() {
	masterMU.Lock()
	defer masterMU.Unlock()
	if _, ok := replicationMUs[ctx.Admin]; !ok {
		replicationMUs[ctx.Admin] = &sync.Mutex{}
	}
	replicationMUs[ctx.Admin].Lock()
	return func() { replicationMUs[ctx.Admin].Unlock() }
}

func getReplications(ctx *kt.Context) {
	defer lockReplication(ctx)()
	ctx.RunAdmin(func(ctx *kt.Context) {
		ctx.Parallel()
		testGetReplications(ctx, ctx.Admin, []struct{}{})
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		ctx.Parallel()
		testGetReplications(ctx, ctx.NoAuth, []struct{}{})
	})
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			ctx.Parallel()

		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			ctx.Parallel()

		})
	})
}

func testRWGetReplications(ctx *kt.Context, client *kivik.Client) {
	dbname1 := ctx.TestDB()
	dbname2 := ctx.TestDB()
	defer ctx.Admin.DestroyDB(context.Background(), dbname1, ctx.Options("db"))
	defer ctx.Admin.DestroyDB(context.Background(), dbname2, ctx.Options("db"))
	ctx.Run("group", func(ctx *kt.Context) {
		ctx.Run("ValidReplication", func(ctx *kt.Context) {
			// TODO
		})
		ctx.Run("MissingSource", func(ctx *kt.Context) {
			// TODO
		})
		ctx.Run("MissingTarget", func(ctx *kt.Context) {
			// TODO
		})
	})
}

func testGetReplications(ctx *kt.Context, client *kivik.Client, expected interface{}) {
	reps, err := client.GetReplications(context.Background())
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	if d := diff.AsJSON(expected, reps); d != nil {
		ctx.Errorf("GetReplications results differ:\n%s\n", d)
	}
}
