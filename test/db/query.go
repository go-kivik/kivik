package db

import (
	"context"
	"sort"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
	"github.com/pkg/errors"
)

func init() {
	kt.Register("Query", query)
}

func query(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		testQueryRW(ctx)
	})
}

func testQueryRW(ctx *kt.Context) {
	if ctx.Admin == nil {
		// Can't do anything here without admin access
		return
	}
	dbName, expected, err := setUpQueryTest(ctx)
	if err != nil {
		ctx.Errorf("Failed to set up temp db: %s", err)
	}
	defer ctx.Admin.DestroyDB(context.Background(), dbName, ctx.Options("db"))
	ctx.Run("group", func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			doQueryTest(ctx, ctx.Admin, dbName, 0, expected)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			doQueryTest(ctx, ctx.NoAuth, dbName, 0, expected)
		})
	})
}

var ddoc = map[string]interface{}{
	"_id":      "_design/testddoc",
	"language": "javascript",
	"views": map[string]interface{}{
		"testview": map[string]interface{}{
			"map": `function(doc) {
                if (doc.include) {
                    emit(doc._id, doc.index);
                }
            }`,
		},
	},
}

func setUpQueryTest(ctx *kt.Context) (dbName string, docIDs []string, err error) {
	dbName = ctx.TestDB()
	db, err := ctx.Admin.DB(context.Background(), dbName, ctx.Options("db"))
	if err != nil {
		return dbName, nil, errors.Wrap(err, "failed to connect to db")
	}
	if _, err := db.Put(context.Background(), ddoc["_id"].(string), ddoc); err != nil {
		return dbName, nil, errors.Wrap(err, "failed to create design doc")
	}
	docIDs = make([]string, 10)
	for i := range docIDs {
		id := ctx.TestDBName()
		doc := struct {
			ID      string `json:"id"`
			Include bool   `json:"include"`
			Index   int    `json:"index"`
		}{
			ID:      id,
			Include: true,
			Index:   i,
		}
		if _, err := db.Put(context.Background(), doc.ID, doc); err != nil {
			return dbName, nil, errors.Wrap(err, "failed to create doc")
		}
		docIDs[i] = id
	}
	sort.Strings(docIDs)
	return dbName, docIDs, nil

}

func doQueryTest(ctx *kt.Context, client *kivik.Client, dbName string, expOffset int64, expected []string) {
	ctx.Run("WithDocs", func(ctx *kt.Context) {
		doQueryTestWithDocs(ctx, client, dbName, expOffset, expected)
	})
	ctx.Run("WithoutDocs", func(ctx *kt.Context) {
		doQueryTestWithoutDocs(ctx, client, dbName, expOffset, expected)
	})
}

func doQueryTestWithoutDocs(ctx *kt.Context, client *kivik.Client, dbName string, expOffset int64, expected []string) {
	ctx.Parallel()
	db, err := client.DB(context.Background(), dbName, ctx.Options("db"))
	// Errors may be deferred here, so only return if we actually get
	// an error.
	if err != nil && !ctx.IsExpectedSuccess(err) {
		return
	}

	rows, err := db.Query(context.Background(), "testddoc", "testview")
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	docIDs := make([]string, 0, len(expected))
	var scanTested bool
	for rows.Next() {
		docIDs = append(docIDs, rows.ID())
		if !scanTested {
			scanTested = true
			ctx.Run("ScanDoc", func(ctx *kt.Context) {
				var i interface{}
				ctx.CheckError(rows.ScanDoc(&i))
			})
			ctx.Run("ScanValue", func(ctx *kt.Context) {
				var i interface{}
				ctx.CheckError(rows.ScanValue(&i))
			})
		}
	}
	if rows.Err() != nil {
		ctx.Fatalf("Failed to fetch row: %s", rows.Err())
	}
	if d := diff.TextSlices(expected, docIDs); d != nil {
		ctx.Errorf("Unexpected document IDs returned:\n%s\n", d)
	}
	if expOffset != rows.Offset() {
		ctx.Errorf("offset: Expected %d, got %d", expOffset, rows.Offset())
	}
	if int64(len(expected)) != rows.TotalRows() {
		ctx.Errorf("total rows: Expected %d, got %d", len(expected), rows.TotalRows())
	}
}

func doQueryTestWithDocs(ctx *kt.Context, client *kivik.Client, dbName string, expOffset int64, expected []string) {
	ctx.Parallel()
	db, err := client.DB(context.Background(), dbName, ctx.Options("db"))
	// Errors may be deferred here, so only return if we actually get
	// an error.
	if err != nil && !ctx.IsExpectedSuccess(err) {
		return
	}
	opts := map[string]interface{}{
		"include_docs": true,
		"update_seq":   true,
	}

	rows, err := db.Query(context.Background(), "testddoc", "testview", opts)
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	docIDs := make([]string, 0, len(expected))
	for rows.Next() {
		var doc struct {
			ID    string `json:"_id"`
			Rev   string `json:"_rev"`
			Index int    `json:"index"`
		}
		if err := rows.ScanDoc(&doc); err != nil {
			ctx.Errorf("Failed to scan doc: %s", err)
		}
		var value int
		if err := rows.ScanValue(&value); err != nil {
			ctx.Errorf("Failed to scan value: %s", err)
		}
		if value != doc.Index {
			ctx.Errorf("doc._rev = %d, but value = %d", doc.Index, value)
		}
		if doc.ID != rows.ID() {
			ctx.Errorf("doc._id = %s, but rows.ID = %s", doc.ID, rows.ID())
		}
		docIDs = append(docIDs, rows.ID())
	}
	if rows.Err() != nil {
		ctx.Fatalf("Failed to fetch row: %s", rows.Err())
	}
	if d := diff.TextSlices(expected, docIDs); d != nil {
		ctx.Errorf("Unexpected document IDs returned:\n%s\n", d)
	}
	if expOffset != rows.Offset() {
		ctx.Errorf("offset: Expected %d, got %d", expOffset, rows.Offset())
	}
	ctx.Run("UpdateSeq", func(ctx *kt.Context) {
		if rows.UpdateSeq() == "" {
			ctx.Errorf("Expected updated sequence")
		}
	})
	if int64(len(expected)) != rows.TotalRows() {
		ctx.Errorf("total rows: Expected %d, got %d", len(expected), rows.TotalRows())
	}
}
