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

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kiviktest/v4/kt"
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

func security(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		for _, dbname := range ctx.MustStringSlice("databases") {
			func(dbname string) {
				ctx.Run(dbname, func(ctx *kt.Context) {
					ctx.Parallel()
					testGetSecurity(ctx, ctx.Admin, dbname, nil)
				})
			}(dbname)
		}
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		for _, dbname := range ctx.MustStringSlice("databases") {
			func(dbname string) {
				ctx.Run(dbname, func(ctx *kt.Context) {
					ctx.Parallel()
					testGetSecurity(ctx, ctx.NoAuth, dbname, nil)
				})
			}(dbname)
		}
	})
	ctx.RunRW(func(ctx *kt.Context) {
		dbname := ctx.TestDB()
		defer ctx.DestroyDB(dbname)
		db := ctx.Admin.DB(dbname, ctx.Options("db"))
		if err := db.Err(); err != nil {
			ctx.Fatalf("Failed to open db: %s", err)
		}
		err := kt.Retry(func() error {
			return db.SetSecurity(context.Background(), sec)
		})
		if err != nil {
			ctx.Fatalf("Failed to set security: %s", err)
		}
		ctx.Run("group", func(ctx *kt.Context) {
			ctx.RunAdmin(func(ctx *kt.Context) {
				ctx.Parallel()
				testGetSecurity(ctx, ctx.Admin, dbname, sec)
			})
			ctx.RunNoAuth(func(ctx *kt.Context) {
				ctx.Parallel()
				testGetSecurity(ctx, ctx.NoAuth, dbname, sec)
			})
		})
	})
}

func setSecurity(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			testSetSecurityTests(ctx, ctx.Admin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			testSetSecurityTests(ctx, ctx.NoAuth)
		})
	})
}

func testSetSecurityTests(ctx *kt.Context, client *kivik.Client) {
	ctx.Run("Exists", func(ctx *kt.Context) {
		ctx.Parallel()
		dbname := ctx.TestDB()
		defer ctx.DestroyDB(dbname)
		testSetSecurity(ctx, client, dbname)
	})
	ctx.Run("NotExists", func(ctx *kt.Context) {
		ctx.Parallel()
		dbname := ctx.TestDBName()
		testSetSecurity(ctx, client, dbname)
	})
}

func testSetSecurity(ctx *kt.Context, client *kivik.Client, dbname string) {
	db := client.DB(dbname, ctx.Options("db"))
	if err := db.Err(); err != nil {
		ctx.Fatalf("Failed to open db: %s", err)
	}
	err := kt.Retry(func() error {
		return db.SetSecurity(context.Background(), sec)
	})
	ctx.CheckError(err)
}

func testGetSecurity(ctx *kt.Context, client *kivik.Client, dbname string, expected *kivik.Security) {
	sec, err := func() (*kivik.Security, error) {
		db := client.DB(dbname, ctx.Options("db"))
		if err := db.Err(); err != nil {
			return nil, err
		}
		return db.Security(context.Background())
	}()
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	if expected != nil {
		if d := testy.DiffAsJSON(expected, sec); d != nil {
			ctx.Errorf("Security document differs from expected:\n%s\n", d)
		}
	}
}
