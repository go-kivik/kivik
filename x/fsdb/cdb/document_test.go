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
	"context"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/x/fsdb/filesystem"
)

const isGopherJS117 = runtime.GOARCH == "js"

func TestDocumentPersist(t *testing.T) {
	if isGopherJS117 {
		t.Skip("Tests broken for GopherJS 1.17")
	}
	type tt struct {
		path   string
		doc    *Document
		status int
		err    string
	}
	tests := testy.NewTable()
	tests.Add("nil doc", func(t *testing.T) interface{} {
		var tmpdir string
		tests.Cleanup(testy.TempDir(t, &tmpdir))

		return tt{
			path:   tmpdir,
			status: http.StatusBadRequest,
			err:    "document has no revisions",
		}
	})
	tests.Add("no revs", func(t *testing.T) interface{} {
		var tmpdir string
		tests.Cleanup(testy.TempDir(t, &tmpdir))

		cdb := New(tmpdir, filesystem.Default())

		return tt{
			path:   tmpdir,
			doc:    cdb.NewDocument("foo"),
			status: http.StatusBadRequest,
			err:    "document has no revisions",
		}
	})
	tests.Add("new doc, one rev", func(t *testing.T) interface{} {
		var tmpdir string
		tests.Cleanup(testy.TempDir(t, &tmpdir))

		cdb := New(tmpdir, filesystem.Default())
		doc := cdb.NewDocument("foo")
		rev, err := cdb.NewRevision(map[string]string{
			"value": "bar",
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := doc.AddRevision(context.TODO(), rev, kivik.Params(nil)); err != nil {
			t.Fatal(err)
		}

		return tt{
			path: tmpdir,
			doc:  doc,
		}
	})
	tests.Add("update existing doc", func(t *testing.T) interface{} {
		tmpdir := testy.CopyTempDir(t, "testdata/persist.update", 0)
		tests.Cleanup(func() error {
			return os.RemoveAll(tmpdir)
		})

		cdb := New(tmpdir)
		doc, err := cdb.OpenDocID("foo", kivik.Params(nil))
		if err != nil {
			t.Fatal(err)
		}
		rev, err := cdb.NewRevision(map[string]interface{}{
			"_rev":  "1-xxx",
			"value": "bar",
			"_revisions": map[string]interface{}{
				"start": 2,
				"ids":   []string{"yyy", "xxx"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := doc.AddRevision(context.TODO(), rev, kivik.Params(nil)); err != nil {
			t.Fatal(err)
		}

		return tt{
			path: tmpdir,
			doc:  doc,
		}
	})
	tests.Add("update existing doc with attachments", func(t *testing.T) interface{} {
		tmpdir := testy.CopyTempDir(t, "testdata/persist.att", 0)
		tests.Cleanup(func() error {
			return os.RemoveAll(tmpdir)
		})

		cdb := New(tmpdir)
		doc, err := cdb.OpenDocID("bar", kivik.Params(nil))
		if err != nil {
			t.Fatal(err)
		}
		rev, err := cdb.NewRevision(map[string]interface{}{
			"_rev":  "1-xxx",
			"value": "bar",
			"_revisions": map[string]interface{}{
				"start": 2,
				"ids":   []string{"yyy", "xxx"},
			},
			"_attachments": map[string]interface{}{
				"bar.txt": map[string]interface{}{
					"content_type": "text/plain",
					"data":         []byte("Additional content"),
				},
				"foo.txt": map[string]interface{}{
					"content_type": "text/plain",
					"stub":         true,
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := doc.addRevision(context.TODO(), rev, kivik.Params(nil)); err != nil {
			t.Fatal(err)
		}

		return tt{
			path: tmpdir,
			doc:  doc,
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		err := tt.doc.persist(context.TODO())
		testy.StatusError(t, tt.err, tt.status, err)
		if d := testy.DiffInterface(testy.Snapshot(t), tt.doc, tmpdirRE(tt.path)); d != nil {
			t.Error(d)
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t, "fs"), testy.JSONDir{
			Path:        tt.path,
			NoMD5Sum:    true,
			FileContent: true,
		}); d != nil {
			t.Error(d)
		}
	})
}

func tmpdirRE(path string) testy.Replacement {
	return testy.Replacement{
		Regexp:      regexp.MustCompile(`len=\d+\) "` + regexp.QuoteMeta(path)),
		Replacement: `len=X) "<tmpdir>`,
	}
}

func TestDocumentAddRevision(t *testing.T) {
	if isGopherJS117 {
		t.Skip("Tests broken for GopherJS 1.17")
	}

	type tt struct {
		path     string
		doc      *Document
		rev      *Revision
		options  kivik.Option
		status   int
		err      string
		expected string
	}
	tests := testy.NewTable()
	tests.Add("stub with bad digest", func(t *testing.T) interface{} {
		tmpdir := testy.CopyTempDir(t, "testdata/persist.att", 0)
		tests.Cleanup(func() error {
			return os.RemoveAll(tmpdir)
		})

		cdb := New(tmpdir)
		doc, err := cdb.OpenDocID("bar", kivik.Params(nil))
		if err != nil {
			t.Fatal(err)
		}
		rev, err := cdb.NewRevision(map[string]interface{}{
			"_rev":  "1-xxx",
			"value": "bar",
			"_revisions": map[string]interface{}{
				"start": 2,
				"ids":   []string{"yyy", "xxx"},
			},
			"_attachments": map[string]interface{}{
				"foo.txt": map[string]interface{}{
					"content_type": "text/plain",
					"stub":         true,
					"digest":       "md5-asdf",
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		return tt{
			path:   tmpdir,
			doc:    doc,
			rev:    rev,
			status: http.StatusBadRequest,
			err:    "invalid attachment data for foo.txt",
		}
	})
	tests.Add("stub with wrong revpos", func(t *testing.T) interface{} {
		tmpdir := testy.CopyTempDir(t, "testdata/persist.att", 0)
		tests.Cleanup(func() error {
			return os.RemoveAll(tmpdir)
		})

		cdb := New(tmpdir)
		doc, err := cdb.OpenDocID("bar", kivik.Params(nil))
		if err != nil {
			t.Fatal(err)
		}
		rev, err := cdb.NewRevision(map[string]interface{}{
			"_rev":  "1-xxx",
			"value": "bar",
			"_revisions": map[string]interface{}{
				"start": 2,
				"ids":   []string{"yyy", "xxx"},
			},
			"_attachments": map[string]interface{}{
				"foo.txt": map[string]interface{}{
					"content_type": "text/plain",
					"stub":         true,
					"revpos":       6,
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		return tt{
			path:   tmpdir,
			doc:    doc,
			rev:    rev,
			status: http.StatusBadRequest,
			err:    "invalid attachment data for foo.txt",
		}
	})
	tests.Add("stub with 0 revpos", func(t *testing.T) interface{} {
		tmpdir := testy.CopyTempDir(t, "testdata/persist.att", 0)
		tests.Cleanup(func() error {
			return os.RemoveAll(tmpdir)
		})

		cdb := New(tmpdir)
		doc, err := cdb.OpenDocID("bar", kivik.Params(nil))
		if err != nil {
			t.Fatal(err)
		}
		rev, err := cdb.NewRevision(map[string]interface{}{
			"_rev":  "1-xxx",
			"value": "bar",
			"_revisions": map[string]interface{}{
				"start": 2,
				"ids":   []string{"yyy", "xxx"},
			},
			"_attachments": map[string]interface{}{
				"foo.txt": map[string]interface{}{
					"content_type": "text/plain",
					"stub":         true,
					"revpos":       0,
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		return tt{
			path:   tmpdir,
			doc:    doc,
			rev:    rev,
			status: http.StatusBadRequest,
			err:    "invalid attachment data for foo.txt",
		}
	})
	tests.Add("upload attachment", func(t *testing.T) interface{} {
		var tmpdir string
		tests.Cleanup(testy.TempDir(t, &tmpdir))

		cdb := New(tmpdir)
		doc := cdb.NewDocument("foo")
		rev, err := cdb.NewRevision(map[string]interface{}{
			"value": "bar",
			"_revisions": map[string]interface{}{
				"start": 2,
				"ids":   []string{"yyy", "xxx"},
			},
			"_attachments": map[string]interface{}{
				"!foo.txt": map[string]interface{}{
					"content_type": "text/plain",
					"data":         []byte("some test content"),
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		return tt{
			path:     tmpdir,
			doc:      doc,
			rev:      rev,
			expected: "1-1472ad25836971f236294ad7b19d9f65",
		}
	})
	tests.Add("re-upload identical attachment", func(t *testing.T) interface{} {
		tmpdir := testy.CopyTempDir(t, "testdata/persist.att", 0)
		tests.Cleanup(func() error {
			return os.RemoveAll(tmpdir)
		})

		cdb := New(tmpdir)
		doc, err := cdb.OpenDocID("bar", kivik.Params(nil))
		if err != nil {
			t.Fatal(err)
		}
		rev, err := cdb.NewRevision(map[string]interface{}{
			"_rev": "1-xxx",
			"_attachments": map[string]interface{}{
				"foo.txt": map[string]interface{}{
					"content_type": "text/plain",
					"data":         []byte("Test content\n"),
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		return tt{
			path:     tmpdir,
			doc:      doc,
			rev:      rev,
			expected: "2-61afc657ebc34041a2568f5d5ab9fc71",
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		opts := tt.options
		if opts == nil {
			opts = kivik.Params(nil)
		}
		revid, err := tt.doc.addRevision(context.TODO(), tt.rev, opts)
		testy.StatusError(t, tt.err, tt.status, err)
		if revid != tt.expected {
			t.Errorf("Unexpected revd: %s", revid)
		}
		if d := testy.DiffInterface(testy.Snapshot(t), tt.doc, tmpdirRE(tt.path)); d != nil {
			t.Error(d)
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t, "fs"), testy.JSONDir{
			Path:        tt.path,
			NoMD5Sum:    true,
			FileContent: true,
		}); d != nil {
			t.Error(d)
		}
	})
}
