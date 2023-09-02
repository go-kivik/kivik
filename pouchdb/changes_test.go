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
	"testing"

	"gitlab.com/flimzy/testy"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func TestChanges(t *testing.T) {
	type tst struct {
		opts            map[string]interface{}
		status          int
		err             string
		changesErr      string
		expectedIDs     []string
		expectedLastSeq string
		expectedPending int64
	}
	tests := testy.NewTable()

	tests.Run(t, func(t *testing.T, test tst) {
		ctx := context.Background()
		client, err := kivik.New("pouch", "")
		if err != nil {
			t.Fatalf("Failed to connect to PouchDB/memdown driver: %s", err)
		}
		dbname := kt.TestDBName(t)
		defer client.DestroyDB(ctx, dbname) // nolint: errcheck
		if err := client.CreateDB(ctx, dbname); err != nil {
			t.Fatalf("Failed to create db: %s", err)
		}
		db := client.DB(dbname)
		changes := db.Changes(ctx, test.opts)
		testy.StatusError(t, test.err, test.status, changes.Err())
		results := []string{}
		for changes.Next() {
			results = append(results, changes.ID())
		}
		testy.Error(t, test.changesErr, changes.Err())
		if d := testy.DiffTextSlices(test.expectedIDs, results); d != nil {
			t.Error(d)
		}
		meta, err := changes.Metadata()
		if err != nil {
			t.Fatal(err)
		}
		if ls := meta.LastSeq; ls != test.expectedLastSeq {
			t.Errorf("Unexpected last_seq: %s", ls)
		}
		if p := meta.Pending; p != test.expectedPending {
			t.Errorf("Unexpected pending count: %d", p)
		}
	})
}
