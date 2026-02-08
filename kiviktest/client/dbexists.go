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

package client

import (
	"context"
	"testing"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.Register("DBExists", dbExists)
}

func dbExists(t *testing.T, c *kt.Context) {
	t.Helper()
	c.RunAdmin(t, func(t *testing.T) {
		t.Helper()
		for _, dbName := range c.MustStringSlice(t, "databases") {
			checkDBExists(t, c, c.Admin, dbName)
		}
	})
	c.RunNoAuth(t, func(t *testing.T) {
		t.Helper()
		for _, dbName := range c.MustStringSlice(t, "databases") {
			checkDBExists(t, c, c.NoAuth, dbName)
		}
	})
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		dbName := c.TestDB(t)
		c.RunAdmin(t, func(t *testing.T) {
			t.Helper()
			checkDBExists(t, c, c.Admin, dbName)
		})
		c.RunNoAuth(t, func(t *testing.T) {
			t.Helper()
			checkDBExists(t, c, c.NoAuth, dbName)
		})
	})
}

func checkDBExists(t *testing.T, c *kt.Context, client *kivik.Client, dbName string) { //nolint:thelper
	c.Run(t, dbName, func(t *testing.T) {
		t.Parallel()
		exists, err := client.DBExists(context.Background(), dbName)
		if !c.IsExpectedSuccess(t, err) {
			return
		}
		if c.MustBool(t, "exists") != exists {
			t.Errorf("Expected: %t, actual: %t", c.Bool(t, "exists"), exists)
		}
	})
}
