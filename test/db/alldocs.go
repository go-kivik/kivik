package db

import (
	"sort"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
	"github.com/pkg/errors"
)

func init() {
	kt.Register("AllDocs", allDocs)
}

func allDocs(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		testAllDocs(ctx, ctx.Admin)
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		testAllDocs(ctx, ctx.NoAuth)
	})
	ctx.RunRW(func(ctx *kt.Context) {
		testAllDocsRW(ctx)
	})
}

func testAllDocsRW(ctx *kt.Context) {
	if ctx.Admin == nil {
		// Can't do anything here without admin access
		return
	}
	dbName, expected, err := setUpAllDocsTest(ctx)
	if err != nil {
		ctx.Errorf("Failed to set up temp db: %s", err)
	}
	defer ctx.Admin.DestroyDB(dbName)
	ctx.Run("group", func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			doTest(ctx, ctx.Admin, dbName, 0, expected)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			doTest(ctx, ctx.NoAuth, dbName, 0, expected)
		})
	})
}

func setUpAllDocsTest(ctx *kt.Context) (dbName string, docIDs []string, err error) {
	dbName = ctx.TestDBName()
	if err = ctx.Admin.CreateDB(dbName); err != nil {
		return dbName, nil, errors.Wrap(err, "failed to create db")
	}
	db, err := ctx.Admin.DB(dbName)
	if err != nil {
		return dbName, nil, errors.Wrap(err, "failed to connect to db")
	}
	docIDs = make([]string, 10)
	for i := range docIDs {
		id := ctx.TestDBName()
		doc := struct {
			ID string `json:"id"`
		}{
			ID: id,
		}
		if _, err := db.Put(doc.ID, doc); err != nil {
			return dbName, nil, errors.Wrap(err, "failed to create doc")
		}
		docIDs[i] = id
	}
	sort.Strings(docIDs)
	return dbName, docIDs, nil

}

func testAllDocs(ctx *kt.Context, client *kivik.Client) {
	if !ctx.IsSet("databases") {
		ctx.Errorf("databases not set; Did you configure this test?")
		return
	}
	for _, dbName := range ctx.StringSlice("databases") {
		ctx.Run(dbName, func(ctx *kt.Context) {
			doTest(ctx, client, dbName, ctx.Int("offset"), ctx.StringSlice("expected"))
		})
	}
}

func doTest(ctx *kt.Context, client *kivik.Client, dbName string, expOffset int, expected []string) {
	ctx.Parallel()
	db, err := client.DB(dbName)
	// Errors may be deferred here, so only return if we actually get
	// an error.
	if err != nil && !ctx.IsExpectedSuccess(err) {
		return
	}

	docs := []struct {
		ID string `json:"id"`
	}{}
	offset, total, _, err := db.AllDocs(&docs, nil)
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	if offset != expOffset {
		ctx.Errorf("offset: Expected %d, got %d", expOffset, offset)
	}
	if total != len(expected) {
		ctx.Errorf("total: Expected %d, got %d", len(expected), total)
	}
	docIDs := make([]string, 0, len(docs))
	for _, doc := range docs {
		docIDs = append(docIDs, doc.ID)
	}
	sort.Strings(docIDs)
	if d := diff.TextSlices(expected, docIDs); d != "" {
		ctx.Errorf("Unexpected document IDs returned:\n%s\n", d)
	}
}
