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

package fs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/x/fsdb/filesystem"
)

func TestRevsDiff(t *testing.T) {
	type tt struct {
		ctx          context.Context
		fs           filesystem.Filesystem
		path, dbname string
		revMap       interface{}
		status       int
		err          string
		rowStatus    int
		rowErr       string
	}
	tests := testy.NewTable()
	tests.Add("invalid revMap", tt{
		dbname: "foo",
		revMap: make(chan int),
		status: http.StatusBadRequest,
		err:    "json: unsupported type: chan int",
	})
	tests.Add("empty map", tt{
		dbname: "foo",
		revMap: map[string][]string{},
	})
	tests.Add("real test", tt{
		path:   "testdata",
		dbname: "db_foo",
		revMap: map[string][]string{
			"yamltest": {"3-", "2-xxx", "1-oink"},
			"autorev":  {"6-", "5-", "4-"},
			"newdoc":   {"1-asdf"},
		},
	})
	tests.Add("cancelled context", func(*testing.T) interface{} {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		return tt{
			ctx:    ctx,
			path:   "testdata",
			dbname: "db_foo",
			revMap: map[string][]string{
				"yamltest": {"3-", "2-xxx", "1-oink"},
				"autorev":  {"6-", "5-", "4-"},
				"newdoc":   {"1-asdf"},
			},
			rowStatus: http.StatusInternalServerError,
			rowErr:    "context canceled",
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		dir := tt.path
		if dir == "" {
			dir = tempDir(t)
			t.Cleanup(func() {
				rmdir(t, dir)
			})
		}
		fs := tt.fs
		if fs == nil {
			fs = filesystem.Default()
		}
		c := &client{root: dir, fs: fs}
		db, err := c.newDB(tt.dbname)
		if err != nil {
			t.Fatal(err)
		}
		ctx := tt.ctx
		if ctx == nil {
			ctx = context.Background()
		}
		rows, err := db.RevsDiff(ctx, tt.revMap)
		if d := internal.StatusErrorDiffRE(tt.err, tt.status, err); d != "" {
			t.Error(d)
		}
		if err != nil {
			return
		}
		result := make(map[string]json.RawMessage)
		var row driver.Row
		var rowErr error
		for {
			err := rows.Next(&row)
			if err == io.EOF {
				break
			}
			if err != nil {
				rowErr = err
				break
			}
			var value json.RawMessage
			_ = json.NewDecoder(row.Value).Decode(&value)
			result[row.ID] = value
		}
		if d := internal.StatusErrorDiffRE(tt.rowErr, tt.rowStatus, rowErr); d != "" {
			t.Error(d)
		}
		if rowErr != nil {
			return
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
	})
}
