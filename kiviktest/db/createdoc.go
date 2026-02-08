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
	"testing"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.Register("CreateDoc", createDoc)
}

func createDoc(t *testing.T, c *kt.Context) {
	t.Helper()
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		dbname := c.TestDB(t)
		c.RunAdmin(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			testCreate(t, c, c.Admin, dbname)
		})
		c.RunNoAuth(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			testCreate(t, c, c.NoAuth, dbname)
		})
	})
}

func testCreate(t *testing.T, c *kt.Context, client *kivik.Client, dbname string) { //nolint:thelper
	db := c.DB(t, client, dbname)
	c.Run(t, "WithoutID", func(t *testing.T) {
		t.Parallel()
		err := kt.Retry(func() error {
			_, _, err := db.CreateDoc(context.Background(), map[string]string{"foo": "bar"})
			return err
		})
		c.CheckError(t, err)
	})
	c.Run(t, "WithID", func(t *testing.T) {
		t.Parallel()
		id := kt.TestDBName(t)
		var docID string
		err := kt.Retry(func() error {
			var err error
			docID, _, err = db.CreateDoc(context.Background(), map[string]string{"foo": "bar", "_id": id})
			return err
		})
		if !c.IsExpectedSuccess(t, err) {
			return
		}
		if id != docID {
			t.Errorf("CreateDoc didn't honor provided ID. Expected '%s', Got '%s'", id, docID)
		}
	})
}
