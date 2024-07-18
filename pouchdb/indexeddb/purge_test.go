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

//go:build js
// +build js

package indexeddb

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gopherjs/gopherjs/js"
	"gitlab.com/flimzy/testy"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
	_ "github.com/go-kivik/kivik/v4/pouchdb" // PouchDB driver we're testing
	"github.com/go-kivik/kivik/v4/pouchdb/internal"
)

func TestPurge(t *testing.T) {
	pouchVer := internal.PouchDBVersion(t)
	if strings.HasPrefix(pouchVer, "7.") {
		t.Skipf("Skipping PouchDB 8+ test for PouchDB %v", pouchVer)
	}

	js.Global.Call("require", "fake-indexeddb/auto")
	indexedDBPlugin := js.Global.Call("require", "pouchdb-adapter-indexeddb")
	pouchDB := js.Global.Get("PouchDB")
	pouchDB.Call("plugin", indexedDBPlugin)
	idbPouch := pouchDB.Call("defaults", map[string]interface{}{
		"adapter": "indexeddb",
	})
	js.Global.Set("PouchDB", idbPouch)

	t.Run("not found", func(t *testing.T) {
		const wantErr = "not_found: missing"
		client, err := kivik.New("pouch", "")
		if err != nil {
			t.Errorf("Failed to connect to PouchDB/memdown driver: %s", err)
			return
		}
		dbname := kt.TestDBName(t)
		ctx := context.Background()
		t.Cleanup(func() {
			_ = client.DestroyDB(ctx, dbname)
		})
		if e := client.CreateDB(ctx, dbname); e != nil {
			t.Fatalf("Failed to create db: %s", e)
		}
		_, err = client.DB(dbname).Purge(ctx, map[string][]string{"foo": {"1-xxx"}})
		if !testy.ErrorMatches(wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
	})
	t.Run("success", func(t *testing.T) {
		client, err := kivik.New("pouch", "")
		if err != nil {
			t.Errorf("Failed to connect to PouchDB/memdown driver: %s", err)
			return
		}
		const docID = "test"
		dbname := kt.TestDBName(t)
		ctx := context.Background()
		t.Cleanup(func() {
			_ = client.DestroyDB(ctx, dbname)
		})
		if e := client.CreateDB(ctx, dbname); e != nil {
			t.Fatalf("Failed to create db: %s", e)
		}
		db := client.DB(dbname)
		rev, err := db.Put(ctx, docID, map[string]string{"foo": "bar"})
		if err != nil {
			t.Fatal(err)
		}
		result, err := db.Purge(ctx, map[string][]string{docID: {rev}})
		if err != nil {
			t.Fatal(err)
		}
		want := &kivik.PurgeResult{
			Seq: 0,
			Purged: map[string][]string{
				docID: {rev},
			},
		}
		if d := cmp.Diff(want, result); d != "" {
			t.Error(d)
		}
	})
}
