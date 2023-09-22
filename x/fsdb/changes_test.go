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
	"io"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

func TestChanges(t *testing.T) {
	type tt struct {
		db      *db
		options kivik.Option
		status  int
		err     string
	}
	tests := testy.NewTable()
	tests.Add("success", tt{
		db: &db{
			client: &client{root: "testdata"},
			dbPath: "testdata/db_foo",
			dbName: "db_foo",
		},
	})
	tests.Add("repl failure", tt{
		db: &db{
			client: &client{root: ""},
			dbPath: "./testdata/source",
			dbName: "source",
		},
		err:    `open testdata/source: [Nn]o such file or directory`,
		status: 500,
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		changes, err := tt.db.Changes(context.TODO(), tt.options)
		testy.StatusErrorRE(t, tt.err, tt.status, err)
		t.Cleanup(func() {
			_ = changes.Close()
		})
		result := make(map[string]driver.Change)
		ch := &driver.Change{}
		for {
			if err := changes.Next(ch); err != nil {
				if err == io.EOF {
					break
				}
				t.Fatal(err)
			}
			result[ch.ID] = *ch
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
	})
}
