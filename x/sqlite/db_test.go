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

package sqlite

import (
	"context"
	"testing"
)

func TestDBPut(t *testing.T) {
	t.Run("rev included", func(t *testing.T) {
		d := &drv{}
		client, err := d.NewClient(":memory:", nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := client.CreateDB(context.Background(), "foo", nil); err != nil {
			t.Fatal(err)
		}
		db, err := client.DB("foo", nil)
		if err != nil {
			t.Fatal(err)
		}
		const wantRev = "1-123"
		rev, err := db.Put(context.Background(), "foo", map[string]interface{}{"_id": "foo", "_rev": "1-123", "foo": "bar"}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if rev != wantRev {
			t.Errorf("Unexpected rev: %s", rev)
		}
	})
	t.Run("create doc", func(t *testing.T) {
		d := &drv{}
		client, err := d.NewClient(":memory:", nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := client.CreateDB(context.Background(), "foo", nil); err != nil {
			t.Fatal(err)
		}
		db, err := client.DB("foo", nil)
		if err != nil {
			t.Fatal(err)
		}
		const wantRev = "1-9bb58f26192e4ba00f01e2e7b136bbd8"
		rev, err := db.Put(context.Background(), "foo", map[string]interface{}{"_id": "foo", "foo": "bar"}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if rev != wantRev {
			t.Errorf("Unexpected rev: %s", rev)
		}
	})
	t.Run("update", func(t *testing.T) {
		d := &drv{}
		client, err := d.NewClient(":memory:", nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := client.CreateDB(context.Background(), "foo", nil); err != nil {
			t.Fatal(err)
		}
		db, err := client.DB("foo", nil)
		if err != nil {
			t.Fatal(err)
		}
		rev, err := db.Put(context.Background(), "foo", map[string]interface{}{"_id": "foo", "foo": "bar"}, nil)
		if err != nil {
			t.Fatal(err)
		}

		// Now the update
		rev2, err := db.Put(context.Background(), "foo", map[string]interface{}{"_id": "foo", "_rev": rev, "foo": "bar"}, nil)
		if err != nil {
			t.Fatal(err)
		}
		const wantRev = "2-9bb58f26192e4ba00f01e2e7b136bbd8"
		if rev2 != wantRev {
			t.Errorf("Unexpected rev: %s", rev)
		}
	})
}
