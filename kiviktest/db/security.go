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

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.Register("Security", security)
	kt.Register("SetSecurity", setSecurity)
}

var sec = &kivik.Security{
	Admins: kivik.Members{
		Names: []string{"bob", "alice"},
		Roles: []string{"hipsters"},
	},
	Members: kivik.Members{
		Names: []string{"fred"},
		Roles: []string{"beatniks"},
	},
}

func security(t *testing.T, c *kt.Context) {
	t.Helper()
	c.RunAdmin(t, func(t *testing.T) {
		t.Helper()
		for _, dbname := range c.MustStringSlice(t, "databases") {
			func(dbname string) {
				c.Run(t, dbname, func(t *testing.T) {
					t.Parallel()
					testGetSecurity(t, c, c.Admin, dbname, nil)
				})
			}(dbname)
		}
	})
	c.RunNoAuth(t, func(t *testing.T) {
		t.Helper()
		for _, dbname := range c.MustStringSlice(t, "databases") {
			func(dbname string) {
				c.Run(t, dbname, func(t *testing.T) {
					t.Parallel()
					testGetSecurity(t, c, c.NoAuth, dbname, nil)
				})
			}(dbname)
		}
	})
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		dbname := c.TestDB(t)
		db := c.Admin.DB(dbname, c.Options(t, "db"))
		if err := db.Err(); err != nil {
			t.Fatalf("Failed to open db: %s", err)
		}
		err := kt.Retry(func() error {
			return db.SetSecurity(context.Background(), sec)
		})
		if err != nil {
			t.Fatalf("Failed to set security: %s", err)
		}
		c.RunAdmin(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			testGetSecurity(t, c, c.Admin, dbname, sec)
		})
		c.RunNoAuth(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			testGetSecurity(t, c, c.NoAuth, dbname, sec)
		})
	})
}

func setSecurity(t *testing.T, c *kt.Context) {
	t.Helper()
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		c.RunAdmin(t, func(t *testing.T) {
			t.Helper()
			testSetSecurityTests(t, c, c.Admin)
		})
		c.RunNoAuth(t, func(t *testing.T) {
			t.Helper()
			testSetSecurityTests(t, c, c.NoAuth)
		})
	})
}

func testSetSecurityTests(t *testing.T, c *kt.Context, client *kivik.Client) {
	t.Helper()
	c.Run(t, "Exists", func(t *testing.T) {
		t.Parallel()
		dbname := c.TestDB(t)
		testSetSecurity(t, c, client, dbname)
	})
	c.Run(t, "NotExists", func(t *testing.T) {
		t.Parallel()
		dbname := kt.TestDBName(t)
		testSetSecurity(t, c, client, dbname)
	})
}

func testSetSecurity(t *testing.T, c *kt.Context, client *kivik.Client, dbname string) { //nolint:thelper
	db := client.DB(dbname, c.Options(t, "db"))
	if err := db.Err(); err != nil {
		t.Fatalf("Failed to open db: %s", err)
	}
	err := kt.Retry(func() error {
		return db.SetSecurity(context.Background(), sec)
	})
	c.CheckError(t, err)
}

func testGetSecurity(t *testing.T, c *kt.Context, client *kivik.Client, dbname string, expected *kivik.Security) { //nolint:thelper
	sec, err := func() (*kivik.Security, error) {
		db := client.DB(dbname, c.Options(t, "db"))
		if err := db.Err(); err != nil {
			return nil, err
		}
		return db.Security(context.Background())
	}()
	if !c.IsExpectedSuccess(t, err) {
		return
	}
	if expected != nil {
		if d := testy.DiffAsJSON(expected, sec); d != nil {
			t.Errorf("Security document differs from expected:\n%s\n", d)
		}
	}
}
