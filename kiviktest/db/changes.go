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
	"testing"
	"time"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.RegisterV2("Changes", changes)
}

func changes(t *testing.T, c *kt.ContextCore) {
	t.Helper()
	c.Run(t, "Normal", func(t *testing.T) {
		c.RunRW(t, func(t *testing.T) {
			t.Helper()
			c.RunAdmin(t, func(t *testing.T) {
				t.Helper()
				testNormalChanges(t, c, c.Admin)
			})
			c.RunNoAuth(t, func(t *testing.T) {
				t.Helper()
				testNormalChanges(t, c, c.NoAuth)
			})
		})
	})
	c.Run(t, "Continuous", func(t *testing.T) {
		c.RunRW(t, func(t *testing.T) {
			t.Helper()
			c.RunAdmin(t, func(t *testing.T) {
				t.Helper()
				testContinuousChanges(t, c, c.Admin)
			})
			c.RunNoAuth(t, func(t *testing.T) {
				t.Helper()
				testContinuousChanges(t, c, c.NoAuth)
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

func testContinuousChanges(t *testing.T, c *kt.ContextCore, client *kivik.Client) { //nolint:thelper
	t.Parallel()
	dbname := c.TestDB(t)
	db := client.DB(dbname, c.Options(t, "db"))
	if err := db.Err(); err != nil {
		t.Fatalf("failed to connect to db: %s", err)
	}
	changes := db.Changes(context.Background(), c.Options(t, "options"))

	const maxChanges = 3
	expected := make([]string, 0, maxChanges)
	doc := cDoc{
		ID:    kt.TestDBName(t),
		Value: "foo",
	}
	rev, err := c.Admin.DB(dbname).Put(context.Background(), doc.ID, doc)
	if err != nil {
		t.Fatalf("Failed to create doc: %s", err)
	}
	expected = append(expected, rev)
	doc.Rev = rev
	doc.Value = "bar"
	rev, err = c.Admin.DB(dbname).Put(context.Background(), doc.ID, doc)
	if err != nil {
		t.Fatalf("Failed to update doc: %s", err)
	}
	expected = append(expected, rev)
	doc.Rev = rev
	const delay = 10 * time.Millisecond
	time.Sleep(delay) // Pause to ensure that the update counts as a separate rev; especially problematic on PouchDB
	rev, err = c.Admin.DB(dbname).Delete(context.Background(), doc.ID, doc.Rev)
	if err != nil {
		t.Fatalf("Failed to delete doc: %s", err)
	}
	expected = append(expected, rev)
	const maxRevs = 3
	revs := make([]string, 0, maxRevs)
	done := make(chan struct{})
	go func() {
		for changes.Next() {
			revs = append(revs, changes.Changes()...)
			if len(revs) >= len(expected) {
				_ = changes.Close()
			}
		}
		close(done)
	}()
	timer := time.NewTimer(maxWait)
	select {
	case <-done:
		timer.Stop()
	case <-timer.C:
		_ = changes.Close()
		t.Errorf("Failed to read changes in %s", maxWait)
	}
	if !c.IsExpectedSuccess(t, changes.Err()) {
		return
	}
	expectedRevs := make(map[string]struct{})
	for _, rev := range expected {
		expectedRevs[rev] = struct{}{}
	}
	for _, rev := range revs {
		if _, ok := expectedRevs[rev]; !ok {
			t.Errorf("Unexpected rev in changes feed: %s", rev)
		}
	}
	if d := testy.DiffTextSlices(expected, revs); d != nil {
		t.Errorf("Unexpected revisions:\n%s", d)
	}
	if err = changes.Close(); err != nil {
		t.Errorf("Error closing changes feed: %s", err)
	}
}

func testNormalChanges(t *testing.T, c *kt.ContextCore, client *kivik.Client) { //nolint:thelper
	t.Parallel()
	dbname := c.TestDB(t)
	db := client.DB(dbname, c.Options(t, "db"))
	if err := db.Err(); err != nil {
		t.Fatalf("failed to connect to db: %s", err)
	}
	adb := c.Admin.DB(dbname)
	const maxChanges = 3
	expected := make([]string, 0, maxChanges)

	// Doc: foo
	doc := cDoc{
		ID:    kt.TestDBName(t),
		Value: "foo",
	}
	rev, err := adb.Put(context.Background(), doc.ID, doc)
	if err != nil {
		t.Fatalf("Failed to create doc: %s", err)
	}
	expected = append(expected, rev)

	// Doc: bar
	doc = cDoc{
		ID:    kt.TestDBName(t),
		Value: "bar",
	}
	rev, err = adb.Put(context.Background(), doc.ID, doc)
	if err != nil {
		t.Fatalf("Failed to create doc: %s", err)
	}
	doc.Rev = rev
	doc.Value = "baz"
	rev, err = adb.Put(context.Background(), doc.ID, doc)
	if err != nil {
		t.Fatalf("Failed to update doc: %s", err)
	}
	expected = append(expected, rev)

	// Doc: baz
	doc = cDoc{
		ID:    kt.TestDBName(t),
		Value: "bar",
	}
	rev, err = adb.Put(context.Background(), doc.ID, doc)
	if err != nil {
		t.Fatalf("Failed to create doc: %s", err)
	}
	doc.Rev = rev
	rev, err = adb.Delete(context.Background(), doc.ID, doc.Rev)
	if err != nil {
		t.Fatalf("Failed to delete doc: %s", err)
	}
	expected = append(expected, rev)

	changes := db.Changes(context.Background(), c.Options(t, "options"))

	const maxRevs = 3
	revs := make([]string, 0, maxRevs)
	for changes.Next() {
		revs = append(revs, changes.Changes()...)
		if len(revs) >= len(expected) {
			_ = changes.Close()
		}
	}
	if !c.IsExpectedSuccess(t, changes.Err()) {
		return
	}
	expectedRevs := make(map[string]struct{})
	for _, rev := range expected {
		expectedRevs[rev] = struct{}{}
	}
	for _, rev := range revs {
		if _, ok := expectedRevs[rev]; !ok {
			t.Errorf("Unexpected rev in changes feed: %s", rev)
		}
	}
	sort.Strings(expected)
	sort.Strings(revs)
	if d := testy.DiffTextSlices(expected, revs); d != nil {
		t.Errorf("Unexpected revisions:\n%s", d)
	}
	if err = changes.Close(); err != nil {
		t.Errorf("Error closing changes feed: %s", err)
	}
}
