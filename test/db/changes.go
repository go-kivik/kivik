package db

import (
	"time"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("Changes", changes)
}

func changes(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.Run("group", func(ctx *kt.Context) {
			ctx.RunAdmin(func(ctx *kt.Context) {
				testChanges(ctx, ctx.Admin)
			})
			ctx.RunNoAuth(func(ctx *kt.Context) {
				testChanges(ctx, ctx.NoAuth)
			})
		})
	})
}

const maxWait = 5 * time.Second

func testChanges(ctx *kt.Context, client *kivik.Client) {
	ctx.Parallel()
	dbname := ctx.TestDBName()
	defer ctx.Admin.DestroyDB(dbname)
	if err := ctx.Admin.CreateDB(dbname); err != nil {
		ctx.Fatalf("Failed to create db: %s", err)
	}
	db, err := client.DB(dbname)
	if err != nil {
		ctx.Fatalf("failed to connect to db: %s", err)
	}
	changes, err := db.Changes(nil)
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	type Doc struct {
		ID    string `json:"_id"`
		Rev   string `json:"_rev,omitempty"`
		Value string `json:"value"`
	}
	expected := make([]string, 0, 3)
	doc := Doc{
		ID:    ctx.TestDBName(),
		Value: "foo",
	}
	rev, err := db.Put(doc.ID, doc)
	if err != nil {
		ctx.Fatalf("Failed to create doc: %s", err)
	}
	expected = append(expected, rev)
	doc.Rev = rev
	doc.Value = "bar"
	rev, err = db.Put(doc.ID, doc)
	if err != nil {
		ctx.Fatalf("Failed to update doc: %s", err)
	}
	expected = append(expected, rev)
	doc.Rev = rev
	rev, err = db.Delete(doc.ID, doc.Rev)
	if err != nil {
		ctx.Fatalf("Failed to delete doc: %s", err)
	}
	expected = append(expected, rev)
	revs := make([]string, 0, 3)
	errChan := make(chan error)
	go func() {
		for changes.Next() {
			for _, ch := range changes.Changes() {
				revs = append(revs, ch)
			}
			if len(revs) >= len(expected) {
				changes.Close()
			}
		}
		if err = changes.Err(); err != nil {
			errChan <- err
		}
		close(errChan)
	}()
	timer := time.NewTimer(maxWait)
	select {
	case chErr, ok := <-errChan:
		if ok {
			ctx.Errorf("Error reading changes: %s", chErr)
		}
	case <-timer.C:
		ctx.Errorf("Failed to read changes in %s", maxWait)
	}
	if err = changes.Err(); err != nil {
		ctx.Errorf("iteration failed: %s", err)
	}
	if d := diff.AsJSON(expected, revs); d != "" {
		ctx.Errorf("Changes revs not as expected:\n%s\n", d)
	}
	if err = changes.Close(); err != nil {
		ctx.Errorf("Error closing changes feed: %s", err)
	}
}
