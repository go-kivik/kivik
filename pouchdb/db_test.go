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

package pouchdb

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/gopherjs/gopherjs/js"
	"gitlab.com/flimzy/testy"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	memPouch := js.Global.Get("PouchDB").Call("defaults", map[interface{}]interface{}{
		"db": js.Global.Call("require", "memdown"),
	})
	js.Global.Set("PouchDB", memPouch)
}

func TestPut(t *testing.T) {
	client, err := kivik.New("pouch", "")
	if err != nil {
		t.Errorf("Failed to connect to PouchDB/memdown driver: %s", err)
		return
	}
	dbname := kt.TestDBName(t)
	ctx := context.Background()
	defer client.DestroyDB(ctx, dbname) // nolint: errcheck
	if e := client.CreateDB(ctx, dbname); e != nil {
		t.Fatalf("Failed to create db: %s", e)
	}
	_, err = client.DB(dbname).Put(ctx, "foo", map[string]string{"_id": "bar"})
	testy.StatusError(t, "id argument must match _id field in document", http.StatusBadRequest, err)
}

func TestPurge(t *testing.T) {
	client, err := kivik.New("pouch", "")
	if err != nil {
		t.Errorf("Failed to connect to PouchDB/memdown driver: %s", err)
		return
	}
	v, _ := client.Version(context.Background())
	pouchVer := v.Version

	t.Run("PouchDB 7", func(t *testing.T) {
		if !strings.HasPrefix(pouchVer, "7.") {
			t.Skipf("Skipping PouchDB 7 test for PouchDB %v", pouchVer)
		}
		const wantErr = "kivik: purge supported by PouchDB 8 or newer"
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
			fmt.Fprintf(os.Stderr, "%s\n", err)
			t.Errorf("Unexpected error: %s", err)
		}
	})
	t.Run("no IndexedDB", func(t *testing.T) {
		if strings.HasPrefix(pouchVer, "7.") {
			t.Skipf("Skipping PouchDB 8 test for PouchDB %v", pouchVer)
		}
		const wantErr = "kivik: purge only supported with indexedDB adapter"
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
			fmt.Fprintf(os.Stderr, "%s\n", err)
			t.Errorf("Unexpected error: %s", err)
		}
	})
}
