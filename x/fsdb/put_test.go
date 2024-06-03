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
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/x/fsdb/filesystem"
)

func TestPut(t *testing.T) {
	if isGopherJS117 {
		t.Skip("Tests broken for GopherJS 1.17")
	}

	type tt struct {
		fs       filesystem.Filesystem
		path     string
		dbname   string
		id       string
		doc      interface{}
		options  kivik.Option
		status   int
		err      string
		expected string
	}
	tests := testy.NewTable()
	tests.Add("invalid docID", tt{
		path:   "doesntmatter",
		dbname: "doesntmatter",
		id:     "_foo",
		status: http.StatusBadRequest,
		err:    "only reserved document ids may start with underscore",
	})
	tests.Add("invalid document", tt{
		path:   "doesntmatter",
		dbname: "doesntmatter",
		id:     "foo",
		doc:    make(chan int),
		status: http.StatusBadRequest,
		err:    "json: unsupported type: chan int",
	})
	tests.Add("create with revid", func(t *testing.T) interface{} {
		tmpdir := tempDir(t)
		tests.Cleanup(cleanTmpdir(tmpdir))
		if err := os.Mkdir(filepath.Join(tmpdir, "foo"), 0o777); err != nil {
			t.Fatal(err)
		}

		return tt{
			path:   tmpdir,
			dbname: "foo",
			id:     "foo",
			doc:    map[string]string{"foo": "bar", "_rev": "1-xxx"},
			status: http.StatusConflict,
			err:    "document update conflict",
		}
	})
	tests.Add("simple create", func(t *testing.T) interface{} {
		tmpdir := tempDir(t)
		tests.Cleanup(cleanTmpdir(tmpdir))
		if err := os.Mkdir(filepath.Join(tmpdir, "foo"), 0o777); err != nil {
			t.Fatal(err)
		}

		return tt{
			path:     tmpdir,
			dbname:   "foo",
			id:       "foo",
			doc:      map[string]string{"foo": "bar"},
			expected: "1-04edfaf9abdaed3c0accf6c463e78fd4",
		}
	})
	tests.Add("update conflict, doc key", func(t *testing.T) interface{} {
		tmpdir := testy.CopyTempDir(t, "testdata/db_put", 1)
		tests.Cleanup(cleanTmpdir(tmpdir))

		return tt{
			path:   tmpdir,
			dbname: "db_put",
			id:     "foo",
			doc:    map[string]string{"foo": "bar", "_rev": "2-asdf"},
			status: http.StatusConflict,
			err:    "document update conflict",
		}
	})
	tests.Add("update conflict, options", func(t *testing.T) interface{} {
		tmpdir := testy.CopyTempDir(t, "testdata/db_put", 1)
		tests.Cleanup(cleanTmpdir(tmpdir))

		return tt{
			path:    tmpdir,
			dbname:  "db_put",
			id:      "foo",
			doc:     map[string]string{"foo": "bar"},
			options: kivik.Rev("2-asdf"),
			status:  http.StatusConflict,
			err:     "document update conflict",
		}
	})
	tests.Add("no explicit rev", func(t *testing.T) interface{} {
		tmpdir := testy.CopyTempDir(t, "testdata/db_put", 1)
		tests.Cleanup(cleanTmpdir(tmpdir))

		return tt{
			path:   tmpdir,
			dbname: "db_put",
			id:     "foo",
			doc:    map[string]string{"foo": "bar"},
			status: http.StatusConflict,
			err:    "document update conflict",
		}
	})
	tests.Add("revs mismatch", tt{
		path:    "/tmp",
		dbname:  "doesntmatter",
		id:      "foo",
		doc:     map[string]string{"foo": "bar", "_rev": "2-asdf"},
		options: kivik.Rev("3-asdf"),
		status:  http.StatusBadRequest,
		err:     "document rev from request body and query string have different values",
	})
	tests.Add("proper update", func(t *testing.T) interface{} {
		tmpdir := testy.CopyTempDir(t, "testdata/db_put", 1)
		tests.Cleanup(cleanTmpdir(tmpdir))

		return tt{
			path:     tmpdir,
			dbname:   "db_put",
			id:       "foo",
			doc:      map[string]string{"foo": "quxx", "_rev": "1-beea34a62a215ab051862d1e5d93162e"},
			expected: "2-ff3a4f106331244679a6cac83a74ae48",
		}
	})
	tests.Add("design doc", func(t *testing.T) interface{} {
		tmpdir := tempDir(t)
		tests.Cleanup(cleanTmpdir(tmpdir))
		if err := os.Mkdir(filepath.Join(tmpdir, "foo"), 0o777); err != nil {
			t.Fatal(err)
		}

		return tt{
			path:     tmpdir,
			dbname:   "foo",
			id:       "_design/foo",
			doc:      map[string]string{"foo": "bar"},
			expected: "1-04edfaf9abdaed3c0accf6c463e78fd4",
		}
	})
	tests.Add("invalid doc id", tt{
		path:   "/tmp",
		dbname: "doesntmatter",
		id:     "_oink",
		doc:    map[string]string{"foo": "bar"},
		status: http.StatusBadRequest,
		err:    "only reserved document ids may start with underscore",
	})
	tests.Add("invalid attachments", tt{
		path:   "/tmp",
		dbname: "doesntmatter",
		id:     "foo",
		doc: map[string]interface{}{
			"foo":          "bar",
			"_attachments": 123,
		},
		status: http.StatusBadRequest,
		err:    "json: cannot unmarshal number into Go struct field RevMeta._attachments of type map[string]*cdb.Attachment",
	})
	tests.Add("attachment", func(t *testing.T) interface{} {
		tmpdir := tempDir(t)
		tests.Cleanup(cleanTmpdir(tmpdir))
		if err := os.Mkdir(filepath.Join(tmpdir, "foo"), 0o777); err != nil {
			t.Fatal(err)
		}

		return tt{
			path:   tmpdir,
			dbname: "foo",
			id:     "foo",
			doc: map[string]interface{}{
				"foo": "bar",
				"_attachments": map[string]interface{}{
					"foo.txt": map[string]interface{}{
						"content_type": "text/plain",
						"data":         []byte("Testing"),
					},
				},
			},
			expected: "1-c706e75b505ddddeed04b959cfcb0ace",
		}
	})
	tests.Add("new_edits=false, no rev", func(t *testing.T) interface{} {
		tmpdir := tempDir(t)
		tests.Cleanup(cleanTmpdir(tmpdir))
		if err := os.Mkdir(filepath.Join(tmpdir, "foo"), 0o777); err != nil {
			t.Fatal(err)
		}

		return tt{
			path:   tmpdir,
			dbname: "foo",
			id:     "foo",
			doc: map[string]string{
				"foo": "bar",
			},
			options: kivik.Param("new_edits", false),
			status:  http.StatusBadRequest,
			err:     "_rev required with new_edits=false",
		}
	})
	tests.Add("new_edits=false, rev already exists", func(t *testing.T) interface{} {
		tmpdir := testy.CopyTempDir(t, "testdata/db_put", 1)
		tests.Cleanup(cleanTmpdir(tmpdir))

		return tt{
			path:   tmpdir,
			dbname: "db_put",
			id:     "foo",
			doc: map[string]string{
				"_rev": "1-beea34a62a215ab051862d1e5d93162e",
				"foo":  "bar",
			},
			options:  kivik.Param("new_edits", false),
			expected: "1-beea34a62a215ab051862d1e5d93162e",
		}
	})
	tests.Add("new_edits=false", func(t *testing.T) interface{} {
		tmpdir := testy.CopyTempDir(t, "testdata/db_put", 1)
		tests.Cleanup(cleanTmpdir(tmpdir))

		return tt{
			path:   tmpdir,
			dbname: "db_put",
			id:     "foo",
			doc: map[string]string{
				"_rev": "1-other",
				"foo":  "bar",
			},
			options:  kivik.Param("new_edits", false),
			expected: "1-other",
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		if tt.path == "" {
			t.Fatalf("path must be set")
		}
		fs := tt.fs
		if fs == nil {
			fs = filesystem.Default()
		}
		c := &client{root: tt.path, fs: fs}
		db, err := c.newDB(tt.dbname)
		if err != nil {
			t.Fatal(err)
		}
		opts := tt.options
		if opts == nil {
			opts = kivik.Params(nil)
		}
		rev, err := db.Put(context.Background(), tt.id, tt.doc, opts)
		if d := internal.StatusErrorDiff(tt.err, tt.status, err); d != "" {
			t.Error(d)
		}
		if err != nil {
			return
		}
		if rev != tt.expected {
			t.Errorf("Unexpected rev returned: %s", rev)
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
