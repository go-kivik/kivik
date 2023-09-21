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
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/x/fsdb/filesystem"
)

func TestSecurity(t *testing.T) {
	type tt struct {
		fs           filesystem.Filesystem
		path, dbname string
		status       int
		err          string
	}
	tests := testy.NewTable()
	tests.Add("no security object", tt{
		dbname: "foo",
	})
	tests.Add("json security obj", tt{
		path:   "testdata",
		dbname: "db_foo",
	})
	tests.Add("yaml security obj", tt{
		path:   "testdata",
		dbname: "db_bar",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		dir := tt.path
		if dir == "" {
			dir = tempDir(t)
			defer rmdir(t, dir)
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
		sec, err := db.Security(context.Background())
		testy.StatusErrorRE(t, tt.err, tt.status, err)
		if d := testy.DiffAsJSON(testy.Snapshot(t), sec); d != nil {
			t.Error(d)
		}
	})
}
