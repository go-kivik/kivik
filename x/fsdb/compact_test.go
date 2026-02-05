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
	"errors"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"gitlab.com/flimzy/testy"

	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/x/fsdb/filesystem"
)

const isGopherJS117 = runtime.GOARCH == "js"

func TestCompact(t *testing.T) {
	if isGopherJS117 {
		t.Skip("Tests broken for GopherJS 1.17")
	}
	type tt struct {
		fs     filesystem.Filesystem
		path   string
		dbname string
		status int
		err    string
	}
	tests := testy.NewTable()
	tests.Add("directory does not exist", tt{
		path:   "testdata",
		dbname: "notfound",
		status: http.StatusNotFound,
		err:    "^open testdata/notfound: [Nn]o such file or directory$",
	})
	tests.Add("empty directory", func(t *testing.T) interface{} {
		tmpdir := tempDir(t)
		t.Cleanup(func() {
			_ = os.RemoveAll(tmpdir)
		})
		if err := os.Mkdir(filepath.Join(tmpdir, "foo"), 0o666); err != nil {
			t.Fatal(err)
		}

		return tt{
			path:   tmpdir,
			dbname: "foo",
		}
	})
	tests.Add("permission denied", tt{
		fs: &filesystem.MockFS{
			OpenFunc: func(_ string) (filesystem.File, error) {
				return nil, statusError{status: http.StatusForbidden, error: errors.New("permission denied")}
			},
		},
		path:   "somepath",
		dbname: "doesntmatter",
		status: http.StatusForbidden,
		err:    "permission denied$",
	})
	tests.Add("no attachments", func(t *testing.T) interface{} {
		tmpdir := testy.CopyTempDir(t, "testdata/compact_noatt", 1)
		t.Cleanup(func() {
			_ = os.RemoveAll(tmpdir)
		})

		return tt{
			path:   tmpdir,
			dbname: "compact_noatt",
		}
	})
	tests.Add("non-winning revs only, no attachments", func(t *testing.T) interface{} {
		tmpdir := testy.CopyTempDir(t, "testdata/compact_nowinner_noatt", 1)
		t.Cleanup(func() {
			_ = os.RemoveAll(tmpdir)
		})

		return tt{
			path:   tmpdir,
			dbname: "compact_nowinner_noatt",
		}
	})
	tests.Add("clean up old revs", func(t *testing.T) interface{} {
		tmpdir := testy.CopyTempDir(t, "testdata/compact_oldrevs", 1)
		t.Cleanup(func() {
			_ = os.RemoveAll(tmpdir)
		})

		return tt{
			path:   tmpdir,
			dbname: "compact_oldrevs",
		}
	})
	tests.Add("clean up old revs with atts", func(t *testing.T) interface{} {
		tmpdir := testy.CopyTempDir(t, "testdata/compact_oldrevsatt", 1)
		t.Cleanup(func() {
			_ = os.RemoveAll(tmpdir)
		})

		return tt{
			path:   tmpdir,
			dbname: "compact_oldrevsatt",
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		fs := tt.fs
		if fs == nil {
			fs = filesystem.Default()
		}
		db := &db{
			client: &client{root: tt.path},
			dbPath: path.Join(tt.path, tt.dbname),
			dbName: tt.dbname,
		}
		err := db.compact(t.Context(), fs)
		if d := internal.StatusErrorDiffRE(tt.err, tt.status, err); d != "" {
			t.Error(d)
		}
		if err != nil {
			return
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t), testy.JSONDir{
			Path:        tt.path,
			NoMD5Sum:    true,
			FileContent: true,
		}); d != nil {
			t.Error(d)
		}
	})
}
