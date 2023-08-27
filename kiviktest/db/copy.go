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

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kiviktest/v4/kt"
)

func init() {
	kt.Register("Copy", _copy)
}

func _copy(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		dbname := ctx.TestDB()
		defer ctx.DestroyDB(dbname)
		db := ctx.Admin.DB(dbname, ctx.Options("db"))
		if err := db.Err(); err != nil {
			ctx.Fatalf("Failed to open db: %s", err)
		}

		doc := map[string]string{
			"_id":  "foo",
			"name": "Robert",
		}
		rev, err := db.Put(context.Background(), doc["_id"], doc)
		if err != nil {
			ctx.Fatalf("Failed to create source doc: %s", err)
		}
		doc["_rev"] = rev

		ddoc := map[string]string{
			"_id":  "_design/foo",
			"name": "Robert",
		}
		rev, err = db.Put(context.Background(), ddoc["_id"], ddoc)
		if err != nil {
			ctx.Fatalf("Failed to create source design doc: %s", err)
		}
		ddoc["_rev"] = rev

		local := map[string]string{
			"_id":  "_local/foo",
			"name": "Robert",
		}
		rev, err = db.Put(context.Background(), local["_id"], local)
		if err != nil {
			ctx.Fatalf("Failed to create source design doc: %s", err)
		}
		local["_rev"] = rev

		ctx.Run("group", func(ctx *kt.Context) {
			ctx.RunAdmin(func(ctx *kt.Context) {
				copyTest(ctx, ctx.Admin, dbname, doc)
				copyTest(ctx, ctx.Admin, dbname, ddoc)
				copyTest(ctx, ctx.Admin, dbname, local)
			})
			ctx.RunNoAuth(func(ctx *kt.Context) {
				copyTest(ctx, ctx.NoAuth, dbname, doc)
				copyTest(ctx, ctx.NoAuth, dbname, ddoc)
				copyTest(ctx, ctx.NoAuth, dbname, local)
			})
		})
	})
}

func copyTest(ctx *kt.Context, client *kivik.Client, dbname string, source map[string]string) {
	ctx.Run(source["_id"], func(ctx *kt.Context) {
		ctx.Parallel()
		db := client.DB(dbname, ctx.Options("db"))
		if err := db.Err(); err != nil {
			ctx.Fatalf("Failed to open db: %s", err)
		}
		targetID := ctx.TestDBName()
		rev, err := db.Copy(context.Background(), targetID, source["_id"])
		if !ctx.IsExpectedSuccess(err) {
			return
		}
		ctx.Run("RevCopy", func(ctx *kt.Context) {
			cp := map[string]string{
				"_id":  targetID,
				"name": "Bob",
				"_rev": rev,
			}
			if _, e := db.Put(context.Background(), targetID, cp); e != nil {
				ctx.Fatalf("Failed to update copy: %s", e)
			}
			targetID2 := ctx.TestDBName()
			if _, e := db.Copy(context.Background(), targetID2, targetID, kivik.Options{"rev": rev}); e != nil {
				ctx.Fatalf("Failed to copy doc with rev option: %s", e)
			}
			var readCopy map[string]string
			if err = db.Get(context.Background(), targetID2).ScanDoc(&readCopy); err != nil {
				ctx.Fatalf("Failed to scan copy: %s", err)
			}
			if readCopy["name"] != "Robert" {
				ctx.Errorf("Copy-with-rev failed. Name = %s, expected %s", readCopy["name"], "Robert")
			}
		})
	})
}
