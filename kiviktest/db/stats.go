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
	"strconv"
	"testing"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.Register("Stats", stats)
}

func stats(t *testing.T, c *kt.Context) {
	t.Helper()
	c.RunAdmin(t, func(t *testing.T) {
		t.Helper()
		t.Parallel()
		roTests(t, c, c.Admin)
	})
	c.RunNoAuth(t, func(t *testing.T) {
		t.Helper()
		t.Parallel()
		roTests(t, c, c.NoAuth)
	})
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		t.Parallel()
		c.RunAdmin(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			rwTests(t, c, c.Admin)
		})
		c.RunNoAuth(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			rwTests(t, c, c.NoAuth)
		})
	})
}

func rwTests(t *testing.T, c *kt.Context, client *kivik.Client) { //nolint:thelper
	dbname := c.TestDB(t)
	db := c.Admin.DB(dbname, c.Options(t, "db"))
	if err := db.Err(); err != nil {
		t.Fatalf("Failed to connect to db: %s", err)
	}
	for i := 0; i < 10; i++ {
		id := strconv.Itoa(i)
		rev, err := db.Put(context.Background(), id, struct{}{})
		if err != nil {
			t.Fatalf("Failed to create document ID %s: %s", id, err)
		}
		const deleteThreshold = 5
		if i > deleteThreshold {
			if _, err = db.Delete(context.Background(), id, rev); err != nil {
				t.Fatalf("Failed to delete document ID %s: %s", id, err)
			}
		}
	}
	const docCount = 6
	testDBInfo(t, c, client, dbname, docCount)
}

func roTests(t *testing.T, c *kt.Context, client *kivik.Client) { //nolint:thelper
	for _, dbname := range c.MustStringSlice(t, "databases") {
		func(dbname string) {
			c.Run(t, dbname, func(t *testing.T) {
				t.Parallel()
				testDBInfo(t, c, client, dbname, 0)
			})
		}(dbname)
	}
}

func testDBInfo(t *testing.T, c *kt.Context, client *kivik.Client, dbname string, docCount int64) { //nolint:thelper
	stats, err := client.DB(dbname, c.Options(t, "db")).Stats(context.Background())
	if !c.IsExpectedSuccess(t, err) {
		return
	}
	if stats.Name != dbname {
		t.Errorf("Name: Expected '%s', actual '%s'", dbname, stats.Name)
	}
	if docCount > 0 {
		if docCount != stats.DocCount {
			t.Errorf("DocCount: Expected %d, actual %d", docCount, stats.DocCount)
		}
	}
}
