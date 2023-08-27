// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package db

import (
	"context"
	"sort"
	"time"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kiviktest/v4/kt"
)

func init() {
	kt.Register("Changes", changes)
}

func changes(ctx *kt.Context) {
	ctx.Run("Normal", func(ctx *kt.Context) {
		ctx.RunRW(func(ctx *kt.Context) {
			ctx.Run("group", func(ctx *kt.Context) {
				ctx.RunAdmin(func(ctx *kt.Context) {
					testNormalChanges(ctx, ctx.Admin)
				})
				ctx.RunNoAuth(func(ctx *kt.Context) {
					testNormalChanges(ctx, ctx.NoAuth)
				})
			})
		})
	})
	ctx.Run("Continuous", func(ctx *kt.Context) {
		ctx.RunRW(func(ctx *kt.Context) {
			ctx.Run("group", func(ctx *kt.Context) {
				ctx.RunAdmin(func(ctx *kt.Context) {
					testContinuousChanges(ctx, ctx.Admin)
				})
				ctx.RunNoAuth(func(ctx *kt.Context) {
					testContinuousChanges(ctx, ctx.NoAuth)
				})
			})
		})
	})
}

const maxWait = 5 * time.Second

type cDoc struct {
	ID    string `json:"_id"`
	Rev   string `json:"_rev,omitempty"`
	Value string `json:"value"`
}

func testContinuousChanges(ctx *kt.Context, client *kivik.Client) {
	ctx.Parallel()
	dbname := ctx.TestDB()
	defer ctx.DestroyDB(dbname)
	db := client.DB(dbname, ctx.Options("db"))
	if err := db.Err(); err != nil {
		ctx.Fatalf("failed to connect to db: %s", err)
	}
	changes := db.Changes(context.Background(), ctx.Options("options"))
	if !ctx.IsExpectedSuccess(changes.Err()) {
		return
	}
	expected := make([]string, 0, 3)
	doc := cDoc{
		ID:    ctx.TestDBName(),
		Value: "foo",
	}
	rev, err := ctx.Admin.DB(dbname).Put(context.Background(), doc.ID, doc)
	if err != nil {
		ctx.Fatalf("Failed to create doc: %s", err)
	}
	expected = append(expected, rev)
	doc.Rev = rev
	doc.Value = "bar"
	rev, err = db.Put(context.Background(), doc.ID, doc)
	if err != nil {
		ctx.Fatalf("Failed to update doc: %s", err)
	}
	expected = append(expected, rev)
	doc.Rev = rev
	time.Sleep(10 * time.Millisecond) // Pause to ensure that the update counts as a separate rev; especially problematic on PouchDB
	rev, err = db.Delete(context.Background(), doc.ID, doc.Rev)
	if err != nil {
		ctx.Fatalf("Failed to delete doc: %s", err)
	}
	expected = append(expected, rev)
	revs := make([]string, 0, 3)
	errChan := make(chan error)
	go func() {
		for changes.Next() {
			revs = append(revs, changes.Changes()...)
			if len(revs) >= len(expected) {
				_ = changes.Close()
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
		_ = changes.Close()
		ctx.Errorf("Failed to read changes in %s", maxWait)
	}
	if err = changes.Err(); err != nil {
		ctx.Errorf("iteration failed: %s", err)
	}
	expectedRevs := make(map[string]struct{})
	for _, rev := range expected {
		expectedRevs[rev] = struct{}{}
	}
	for _, rev := range revs {
		if _, ok := expectedRevs[rev]; !ok {
			ctx.Errorf("Unexpected rev in changes feed: %s", rev)
		}
	}
	if d := testy.DiffTextSlices(expected, revs); d != nil {
		ctx.Errorf("Unexpected revisions:\n%s", d)
	}
	if err = changes.Close(); err != nil {
		ctx.Errorf("Error closing changes feed: %s", err)
	}
}

func testNormalChanges(ctx *kt.Context, client *kivik.Client) {
	ctx.Parallel()
	dbname := ctx.TestDB()
	defer ctx.DestroyDB(dbname)
	db := client.DB(dbname, ctx.Options("db"))
	if err := db.Err(); err != nil {
		ctx.Fatalf("failed to connect to db: %s", err)
	}
	adb := ctx.Admin.DB(dbname)
	expected := make([]string, 0, 3)

	// Doc: foo
	doc := cDoc{
		ID:    ctx.TestDBName(),
		Value: "foo",
	}
	rev, err := adb.Put(context.Background(), doc.ID, doc)
	if err != nil {
		ctx.Fatalf("Failed to create doc: %s", err)
	}
	expected = append(expected, rev)

	// Doc: bar
	doc = cDoc{
		ID:    ctx.TestDBName(),
		Value: "bar",
	}
	rev, err = adb.Put(context.Background(), doc.ID, doc)
	if err != nil {
		ctx.Fatalf("Failed to create doc: %s", err)
	}
	doc.Rev = rev
	doc.Value = "baz"
	rev, err = adb.Put(context.Background(), doc.ID, doc)
	if err != nil {
		ctx.Fatalf("Failed to update doc: %s", err)
	}
	expected = append(expected, rev)

	// Doc: baz
	doc = cDoc{
		ID:    ctx.TestDBName(),
		Value: "bar",
	}
	rev, err = adb.Put(context.Background(), doc.ID, doc)
	if err != nil {
		ctx.Fatalf("Failed to create doc: %s", err)
	}
	doc.Rev = rev
	rev, err = adb.Delete(context.Background(), doc.ID, doc.Rev)
	if err != nil {
		ctx.Fatalf("Failed to delete doc: %s", err)
	}
	expected = append(expected, rev)

	changes := db.Changes(context.Background(), ctx.Options("options"))
	if !ctx.IsExpectedSuccess(changes.Err()) {
		return
	}

	revs := make([]string, 0, 3)
	for changes.Next() {
		revs = append(revs, changes.Changes()...)
		if len(revs) >= len(expected) {
			_ = changes.Close()
		}
	}
	if err = changes.Err(); err != nil {
		ctx.Errorf("iteration failed: %s", err)
	}
	expectedRevs := make(map[string]struct{})
	for _, rev := range expected {
		expectedRevs[rev] = struct{}{}
	}
	for _, rev := range revs {
		if _, ok := expectedRevs[rev]; !ok {
			ctx.Errorf("Unexpected rev in changes feed: %s", rev)
		}
	}
	sort.Strings(expected)
	sort.Strings(revs)
	if d := testy.DiffTextSlices(expected, revs); d != nil {
		ctx.Errorf("Unexpected revisions:\n%s", d)
	}
	if err = changes.Close(); err != nil {
		ctx.Errorf("Error closing changes feed: %s", err)
	}
}
