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
	"errors"
	"io"
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal"
	"github.com/go-kivik/kivik/v4/x/fsdb/filesystem"
)

func TestGet(t *testing.T) {
	type tt struct {
		fs           filesystem.Filesystem
		setup        func(*testing.T, *db)
		final        func(*testing.T, *db)
		path, dbname string
		id           string
		options      kivik.Option
		expected     *driver.Result
		status       int
		err          string
	}
	tests := testy.NewTable()
	tests.Add("no id", tt{
		path:   "",
		dbname: "foo",
		status: http.StatusBadRequest,
		err:    "no docid specified",
	})
	tests.Add("not found", tt{
		dbname: "asdf",
		id:     "foo",
		status: http.StatusNotFound,
		err:    `^missing$`,
	})
	tests.Add("forbidden", func(t *testing.T) interface{} {
		return tt{
			fs: &filesystem.MockFS{
				OpenFunc: func(_ string) (filesystem.File, error) {
					return nil, statusError{status: http.StatusForbidden, error: errors.New("permission denied")}
				},
			},
			dbname: "foo",
			id:     "foo",
			status: http.StatusForbidden,
			err:    "permission denied$",
		}
	})
	tests.Add("success, no attachments", tt{
		path:   "testdata",
		dbname: "db_foo",
		id:     "noattach",
		expected: &driver.Result{
			Rev: "1-xxxxxxxxxx",
		},
	})
	tests.Add("success, attachment stub", tt{
		path:   "testdata",
		dbname: "db_foo",
		id:     "withattach",
		expected: &driver.Result{
			Rev: "2-yyyyyyyyy",
		},
	})
	tests.Add("success, include mp attachments", tt{
		path:    "testdata",
		dbname:  "db_foo",
		id:      "withattach",
		options: kivik.Param("attachments", true),
		expected: &driver.Result{
			Rev: "2-yyyyyyyyy",
		},
	})
	tests.Add("success, include inline attachments", tt{
		path:   "testdata",
		dbname: "db_foo",
		id:     "withattach",
		options: kivik.Params(map[string]interface{}{
			"attachments":   true,
			"header:accept": "application/json",
		}),
		expected: &driver.Result{
			Rev: "2-yyyyyyyyy",
		},
	})
	tests.Add("specify current rev", tt{
		path:    "testdata",
		dbname:  "db_foo",
		id:      "noattach",
		options: kivik.Rev("1-xxxxxxxxxx"),
		expected: &driver.Result{
			Rev: "1-xxxxxxxxxx",
		},
	})
	tests.Add("specify old rev", tt{
		path:    "testdata",
		dbname:  "db_foo",
		id:      "withattach",
		options: kivik.Rev("1-xxxxxxxxxx"),
		expected: &driver.Result{
			Rev: "1-xxxxxxxxxx",
		},
	})
	tests.Add("autorev", tt{
		path:   "testdata",
		dbname: "db_foo",
		id:     "autorev",
		expected: &driver.Result{
			Rev: "6-",
		},
	})
	tests.Add("intrev", tt{
		path:   "testdata",
		dbname: "db_foo",
		id:     "intrev",
		expected: &driver.Result{
			Rev: "6-",
		},
	})
	tests.Add("norev", tt{
		path:   "testdata",
		dbname: "db_foo",
		id:     "norev",
		expected: &driver.Result{
			Rev: "1-",
		},
	})
	tests.Add("noid", tt{
		path:   "testdata",
		dbname: "db_foo",
		id:     "noid",
		expected: &driver.Result{
			Rev: "6-",
		},
	})
	tests.Add("wrong id", tt{
		path:   "testdata",
		dbname: "db_foo",
		id:     "wrongid",
		expected: &driver.Result{
			Rev: "6-",
		},
	})
	tests.Add("yaml", tt{
		path:   "testdata",
		dbname: "db_foo",
		id:     "yamltest",
		expected: &driver.Result{
			Rev: "3-",
		},
	})
	tests.Add("specify current rev yaml", tt{
		path:    "testdata",
		dbname:  "db_foo",
		id:      "yamltest",
		options: kivik.Rev("3-"),
		expected: &driver.Result{
			Rev: "3-",
		},
	})
	tests.Add("specify old rev yaml", tt{
		path:    "testdata",
		dbname:  "db_foo",
		id:      "yamltest",
		options: kivik.Rev("2-xxx"),
		expected: &driver.Result{
			Rev: "2-xxx",
		},
	})
	tests.Add("specify bogus rev yaml", tt{
		path:    "testdata",
		dbname:  "db_foo",
		id:      "yamltest",
		options: kivik.Rev("1-oink"),
		status:  http.StatusNotFound,
		err:     "missing",
	})
	tests.Add("ddoc yaml", tt{
		path:   "testdata",
		dbname: "db_foo",
		id:     "_design/users",
		expected: &driver.Result{
			Rev: "2-",
		},
	})
	tests.Add("ddoc rev yaml", tt{
		path:    "testdata",
		dbname:  "db_foo",
		id:      "_design/users",
		options: kivik.Rev("2-"),
		expected: &driver.Result{
			Rev: "2-",
		},
	})
	tests.Add("revs", tt{
		path:    "testdata",
		dbname:  "db_foo",
		id:      "wrongid",
		options: kivik.Param("revs", true),
		expected: &driver.Result{
			Rev: "6-",
		},
	})
	tests.Add("revs real", tt{
		path:    "testdata",
		dbname:  "db_foo",
		id:      "noattach",
		options: kivik.Param("revs", true),
		expected: &driver.Result{
			Rev: "1-xxxxxxxxxx",
		},
	})
	tests.Add("note--XkWjFv13acvjJTt-CGJJ8hXlWE", tt{
		path:   "testdata",
		dbname: "db_att",
		id:     "note--XkWjFv13acvjJTt-CGJJ8hXlWE",
		expected: &driver.Result{
			Rev: "1-fbaabe005e0f4e5685a68f857c0777d6",
		},
	})
	tests.Add("note--XkWjFv13acvjJTt-CGJJ8hXlWE + attachments", tt{
		path:    "testdata",
		dbname:  "db_att",
		id:      "note--XkWjFv13acvjJTt-CGJJ8hXlWE",
		options: kivik.Param("attachments", true),
		expected: &driver.Result{
			Rev: "1-fbaabe005e0f4e5685a68f857c0777d6",
		},
	})
	tests.Add("revs_info=true", tt{
		path:    "testdata",
		dbname:  "db_foo",
		id:      "autorev",
		options: kivik.Param("revs_info", true),
		expected: &driver.Result{
			Rev: "6-",
		},
	})
	tests.Add("revs, explicit", tt{
		path:    "testdata",
		dbname:  "db_foo",
		id:      "withrevs",
		options: kivik.Param("revs", true),
		expected: &driver.Result{
			Rev: "8-asdf",
		},
	})
	tests.Add("specify current rev, revs_info=true", tt{
		path:   "testdata",
		dbname: "db_foo",
		id:     "yamltest",
		options: kivik.Params(map[string]interface{}{
			"rev":       "3-",
			"revs_info": true,
		}),
		expected: &driver.Result{
			Rev: "3-",
		},
	})
	tests.Add("specify conflicting rev, revs_info=true", tt{
		path:   "testdata",
		dbname: "db_foo",
		id:     "yamltest",
		options: kivik.Params(map[string]interface{}{
			"rev":       "2-xxx",
			"revs_info": true,
		}),
		expected: &driver.Result{
			Rev: "2-xxx",
		},
	})
	tests.Add("specify rev, revs=true", tt{
		path:   "testdata",
		dbname: "db_foo",
		id:     "withrevs",
		options: kivik.Params(map[string]interface{}{
			"rev":  "8-asdf",
			"revs": true,
		}),
		expected: &driver.Result{
			Rev: "8-asdf",
		},
	})
	tests.Add("interrupted put", tt{
		// This tests a put which was aborted, leaving the attachments in
		// {db}/.{docid}/{rev}/{filename}, while the winning rev is at
		// the friendlier location of {db}/{docid}.{ext}
		path:    "testdata",
		dbname:  "db_foo",
		id:      "abortedput",
		options: kivik.Param("attachments", true),
		expected: &driver.Result{
			Rev: "2-yyyyyyyyy",
		},
	})
	tests.Add("no winner, tied rev", tt{
		// This tests a put which was aborted, leaving the attachments in
		// {db}/.{docid}/{rev}/{filename}, while the winning rev is at
		// the friendlier location of {db}/{docid}.{ext}
		path:   "testdata",
		dbname: "get_nowinner",
		id:     "foo",
		expected: &driver.Result{
			Rev: "1-yyy",
		},
	})
	tests.Add("no winner, greater rev", tt{
		// This tests a put which was aborted, leaving the attachments in
		// {db}/.{docid}/{rev}/{filename}, while the winning rev is at
		// the friendlier location of {db}/{docid}.{ext}
		path:   "testdata",
		dbname: "get_nowinner",
		id:     "bar",
		expected: &driver.Result{
			Rev: "2-yyy",
		},
	})
	tests.Add("atts split between winning and revs dir", tt{
		path:    "testdata",
		dbname:  "get_split_atts",
		id:      "foo",
		options: kivik.Param("attachments", true),
		expected: &driver.Result{
			Rev: "2-zzz",
		},
	})
	tests.Add("atts split between two revs", tt{
		path:    "testdata",
		dbname:  "get_split_atts",
		id:      "bar",
		options: kivik.Param("attachments", true),
		expected: &driver.Result{
			Rev: "2-yyy",
		},
	})
	tests.Add("non-standard filenames", tt{
		path:   "testdata",
		dbname: "db_nonascii",
		id:     "note-i_ɪ",
		expected: &driver.Result{
			Rev: "1-",
		},
	})
	tests.Add("deleted doc", tt{
		path:   "testdata",
		dbname: "db_foo",
		id:     "deleted",
		status: http.StatusNotFound,
		err:    "deleted",
	})
	tests.Add("deleted doc, specific rev", tt{
		path:    "testdata",
		dbname:  "db_foo",
		id:      "deleted",
		options: kivik.Rev("3-"),
		expected: &driver.Result{
			Rev: "3-",
		},
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
		if tt.setup != nil {
			tt.setup(t, db)
		}
		opts := tt.options
		if opts == nil {
			opts = kivik.Params(nil)
		}
		result, err := db.Get(context.Background(), tt.id, opts)
		if d := internal.StatusErrorDiffRE(tt.err, tt.status, err); d != "" {
			t.Error(d)
		}
		if err != nil {
			return
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t), result.Body); d != nil {
			t.Errorf("document:\n%s", d)
		}
		if result.Attachments != nil {
			attachments := result.Attachments
			t.Cleanup(func() {
				_ = attachments.Close()
			})
			att := &driver.Attachment{}
			for {
				if err := result.Attachments.Next(att); err != nil {
					if err == io.EOF {
						break
					}
					t.Fatal(err)
				}
				if d := testy.DiffText(&testy.File{Path: "testdata/" + testy.Stub(t) + "_" + att.Filename}, att.Content); d != nil {
					t.Errorf("Attachment %s content:\n%s", att.Filename, d)
				}
				_ = att.Content.Close()
				att.Content = nil
				if d := testy.DiffAsJSON(&testy.File{Path: "testdata/" + testy.Stub(t) + "_" + att.Filename + "_struct"}, att); d != nil {
					t.Errorf("Attachment %s struct:\n%s", att.Filename, d)
				}
			}
		}
		result.Body = nil
		result.Attachments = nil
		if d := testy.DiffInterface(tt.expected, result); d != nil {
			t.Error(d)
		}
		if tt.final != nil {
			tt.final(t, db)
		}
	})
}
