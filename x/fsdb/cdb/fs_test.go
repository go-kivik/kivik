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

package cdb

import (
	"errors"
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/x/fsdb/filesystem"
)

func TestFSOpenDocID(t *testing.T) {
	type tt struct {
		fs      filesystem.Filesystem
		root    string
		docID   string
		options kivik.Option
		status  int
		err     string
	}
	tests := testy.NewTable()
	tests.Add("not found", tt{
		root:   "notfound",
		status: http.StatusNotFound,
		err:    "missing",
	})
	tests.Add("main rev only", tt{
		root:  "testdata/open",
		docID: "foo",
	})
	tests.Add("main rev only, yaml", tt{
		root:  "testdata/open",
		docID: "bar",
	})
	tests.Add("no id in doc", tt{
		root:  "testdata/open",
		docID: "noid",
	})
	tests.Add("forbidden", tt{
		fs: &filesystem.MockFS{
			OpenFunc: func(_ string) (filesystem.File, error) {
				return nil, statusError{status: http.StatusForbidden, error: errors.New("permission denied")}
			},
		},
		root:   "doesntmatter",
		docID:  "foo",
		status: http.StatusForbidden,
		err:    "permission denied",
	})
	tests.Add("attachment", tt{
		root:  "testdata/open",
		docID: "att",
	})
	tests.Add("attachments from multiple revs", tt{
		root:  "testdata/open",
		docID: "splitatts",
	})
	tests.Add("no rev", tt{
		root:  "testdata/open",
		docID: "norev",
	})
	tests.Add("no main rev", tt{
		root:  "testdata/open",
		docID: "nomain",
	})
	tests.Add("json auto rev number", tt{
		root:  "testdata/open",
		docID: "jsonautorevnum",
	})
	tests.Add("yaml auto rev number", tt{
		root:  "testdata/open",
		docID: "yamlautorevnum",
	})
	tests.Add("json auto rev string", tt{
		root:  "testdata/open",
		docID: "jsonautorevstr",
	})
	tests.Add("yaml auto rev string", tt{
		root:  "testdata/open",
		docID: "yamlautorevstr",
	})
	tests.Add("multiple revs, winner selected", tt{
		root:  "testdata/open",
		docID: "multiplerevs",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		fs := New(tt.root, tt.fs)
		opts := tt.options
		if opts == nil {
			opts = kivik.Params(nil)
		}
		result, err := fs.OpenDocID(tt.docID, opts)
		if d := internal.StatusErrorDiff(tt.err, tt.status, err); d != "" {
			t.Error(d)
		}
		if err != nil {
			return
		}
		result.Options = map[string]any{
			"revs":          true,
			"attachments":   true,
			"header:accept": "application/json",
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
	})
}

func TestRestoreAttachments(t *testing.T) {
	type tt struct {
		r      *Revision
		status int
		err    string
	}
	tests := testy.NewTable()
	tests.Add("missing attachment", tt{
		r: &Revision{
			options: map[string]any{
				"attachments": true,
			},
			RevMeta: RevMeta{
				fs:         filesystem.Default(),
				RevHistory: &RevHistory{},
				Attachments: map[string]*Attachment{
					"notfound.txt": {
						fs:   filesystem.Default(),
						path: "/somewhere/notfound.txt",
					},
				},
			},
		},
		status: http.StatusInternalServerError,
		err:    "attachment 'notfound.txt': missing",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		err := tt.r.restoreAttachments()
		if d := internal.StatusErrorDiff(tt.err, tt.status, err); d != "" {
			t.Error(d)
		}
	})
}
