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

//go:build !js
// +build !js

package sqlite

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

func TestDBGet(t *testing.T) {
	type test struct {
		db         driver.DB
		id         string
		options    driver.Options
		wantDoc    interface{}
		wantStatus int
		wantErr    string
	}
	tests := testy.NewTable()
	tests.Add("not found", test{
		id:         "foo",
		wantStatus: http.StatusNotFound,
		wantErr:    "not found",
	})
	tests.Add("success", func(t *testing.T) interface{} {
		db := newDB(t)
		rev := db.tPut("foo", map[string]string{"foo": "bar"})

		return test{
			db: db,
			id: "foo",
			wantDoc: map[string]string{
				"_id":  "foo",
				"_rev": rev,
				"foo":  "bar",
			},
		}
	})
	tests.Add("get specific rev", func(t *testing.T) interface{} {
		db := newDB(t)
		rev := db.tPut("foo", map[string]string{"foo": "bar"})
		_ = db.tPut("foo", map[string]string{"foo": "baz"}, kivik.Rev(rev))

		return test{
			db:      db,
			id:      "foo",
			options: kivik.Rev(rev),
			wantDoc: map[string]string{
				"_id":  "foo",
				"_rev": rev,
				"foo":  "bar",
			},
		}
	})
	tests.Add("specific rev not found", test{
		id:         "foo",
		options:    kivik.Rev("1-9bb58f26192e4ba00f01e2e7b136bbd8"),
		wantStatus: http.StatusNotFound,
		wantErr:    "not found",
	})
	tests.Add("include conflicts", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]string{"foo": "bar"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
			"rev":       "1-abc",
		}))
		_ = db.tPut("foo", map[string]string{"foo": "baz"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
			"rev":       "1-xyz",
		}))

		return test{
			db:      db,
			id:      "foo",
			options: kivik.Param("conflicts", true),
			wantDoc: map[string]interface{}{
				"_id":        "foo",
				"_rev":       "1-xyz",
				"foo":        "baz",
				"_conflicts": []string{"1-abc"},
			},
		}
	})
	tests.Add("include only leaf conflicts", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]string{"foo": "bar"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
			"rev":       "1-abc",
		}))
		_ = db.tPut("foo", map[string]string{"foo": "baz"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
			"rev":       "1-xyz",
		}))
		rev := db.tPut("foo", map[string]string{"foo": "qux"}, kivik.Rev("1-xyz"))

		return test{
			db:      db,
			id:      "foo",
			options: kivik.Param("conflicts", true),
			wantDoc: map[string]interface{}{
				"_id":        "foo",
				"_rev":       rev,
				"foo":        "qux",
				"_conflicts": []string{"1-abc"},
			},
		}
	})
	tests.Add("deleted document", func(t *testing.T) interface{} {
		db := newDB(t)
		rev := db.tPut("foo", map[string]string{"foo": "bar"})
		_ = db.tDelete("foo", kivik.Rev(rev))

		return test{
			db:         db,
			id:         "foo",
			wantStatus: http.StatusNotFound,
			wantErr:    "not found",
		}
	})
	tests.Add("deleted document by rev", func(t *testing.T) interface{} {
		db := newDB(t)
		rev := db.tPut("foo", map[string]string{"foo": "bar"})
		rev = db.tDelete("foo", kivik.Rev(rev))

		return test{
			db:      db,
			id:      "foo",
			options: kivik.Rev(rev),
			wantDoc: map[string]interface{}{
				"_id":      "foo",
				"_rev":     rev,
				"_deleted": true,
			},
		}
	})
	tests.Add("deleted document with data by rev", func(t *testing.T) interface{} {
		db := newDB(t)
		rev := db.tPut("foo", map[string]interface{}{"_deleted": true, "foo": "bar"})

		return test{
			db:      db,
			id:      "foo",
			options: kivik.Rev(rev),
			wantDoc: map[string]interface{}{
				"_id":      "foo",
				"_rev":     rev,
				"_deleted": true,
				"foo":      "bar",
			},
		}
	})
	tests.Add("include conflicts, skip deleted conflicts", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]interface{}{"foo": "moo", "_deleted": true}, kivik.Params(map[string]interface{}{
			"new_edits": false,
			"rev":       "1-qwe",
		}))
		_ = db.tPut("foo", map[string]string{"foo": "bar"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
			"rev":       "1-abc",
		}))
		_ = db.tPut("foo", map[string]string{"foo": "baz"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
			"rev":       "1-xyz",
		}))
		rev := db.tPut("foo", map[string]string{"foo": "qux"}, kivik.Rev("1-xyz"))

		return test{
			db:      db,
			id:      "foo",
			options: kivik.Param("conflicts", true),
			wantDoc: map[string]interface{}{
				"_id":        "foo",
				"_rev":       rev,
				"foo":        "qux",
				"_conflicts": []string{"1-abc"},
			},
		}
	})
	tests.Add("include deleted conflicts", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]interface{}{"foo": "moo", "_deleted": true}, kivik.Params(map[string]interface{}{
			"new_edits": false,
			"rev":       "1-qwe",
		}))
		_ = db.tPut("foo", map[string]string{"foo": "bar"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
			"rev":       "1-abc",
		}))
		_ = db.tPut("foo", map[string]string{"foo": "baz"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
			"rev":       "1-xyz",
		}))
		rev := db.tPut("foo", map[string]string{"foo": "qux"}, kivik.Rev("1-xyz"))

		return test{
			db:      db,
			id:      "foo",
			options: kivik.Param("deleted_conflicts", true),
			wantDoc: map[string]interface{}{
				"_id":                "foo",
				"_rev":               rev,
				"foo":                "qux",
				"_deleted_conflicts": []string{"1-qwe"},
			},
		}
	})
	tests.Add("include all conflicts", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]interface{}{"foo": "moo", "_deleted": true}, kivik.Params(map[string]interface{}{
			"new_edits": false,
			"rev":       "1-qwe",
		}))
		_ = db.tPut("foo", map[string]string{"foo": "bar"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
			"rev":       "1-abc",
		}))
		_ = db.tPut("foo", map[string]string{"foo": "baz"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
			"rev":       "1-xyz",
		}))
		rev := db.tPut("foo", map[string]string{"foo": "qux"}, kivik.Rev("1-xyz"))

		return test{
			db: db,
			id: "foo",
			options: kivik.Params(map[string]interface{}{
				"conflicts":         true,
				"deleted_conflicts": true,
			}),
			wantDoc: map[string]interface{}{
				"_id":                "foo",
				"_rev":               rev,
				"foo":                "qux",
				"_deleted_conflicts": []string{"1-qwe"},
				"_conflicts":         []string{"1-abc"},
			},
		}
	})
	tests.Add("include revs_info", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]interface{}{"foo": "moo", "_deleted": true}, kivik.Params(map[string]interface{}{
			"new_edits": false,
			"rev":       "1-qwe",
		}))
		_ = db.tPut("foo", map[string]string{"foo": "bar"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
			"rev":       "1-abc",
		}))
		_ = db.tPut("foo", map[string]string{"foo": "baz"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
			"rev":       "1-xyz",
		}))
		rev := db.tPut("foo", map[string]string{"foo": "qux"}, kivik.Rev("1-xyz"))

		return test{
			db: db,
			id: "foo",
			options: kivik.Params(map[string]interface{}{
				"revs_info": true,
			}),
			wantDoc: map[string]interface{}{
				"_id":  "foo",
				"_rev": rev,
				"foo":  "qux",
				"_revs_info": []map[string]string{
					{"rev": rev, "status": "available"},
					{"rev": "1-xyz", "status": "available"},
				},
			},
		}
	})
	tests.Add("include meta", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]interface{}{"foo": "moo", "_deleted": true}, kivik.Params(map[string]interface{}{
			"new_edits": false,
			"rev":       "1-qwe",
		}))
		_ = db.tPut("foo", map[string]string{"foo": "bar"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
			"rev":       "1-abc",
		}))
		_ = db.tPut("foo", map[string]string{"foo": "baz"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
			"rev":       "1-xyz",
		}))
		rev := db.tPut("foo", map[string]string{"foo": "qux"}, kivik.Rev("1-xyz"))

		return test{
			db:      db,
			id:      "foo",
			options: kivik.Param("meta", true),
			wantDoc: map[string]interface{}{
				"_id":  "foo",
				"_rev": rev,
				"foo":  "qux",
				"_revs_info": []map[string]string{
					{"rev": rev, "status": "available"},
					{"rev": "1-xyz", "status": "available"},
				},
				"_conflicts":         []string{"1-abc"},
				"_deleted_conflicts": []string{"1-qwe"},
			},
		}
	})
	tests.Add("get latest winning leaf", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]interface{}{"foo": "aaa", "_rev": "1-aaa"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))
		_ = db.tPut("foo", map[string]interface{}{
			"foo": "bbb",
			"_revisions": map[string]interface{}{
				"ids":   []string{"bbb", "aaa"},
				"start": 2,
			},
		}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))
		_ = db.tPut("foo", map[string]interface{}{
			"foo": "ddd",
			"_revisions": map[string]interface{}{
				"ids":   []string{"yyy", "aaa"},
				"start": 2,
			},
		}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))

		return test{
			db: db,
			id: "foo",
			options: kivik.Params(map[string]interface{}{
				"latest": true,
				"rev":    "1-aaa",
			}),
			wantDoc: map[string]interface{}{
				"_id":  "foo",
				"_rev": "2-yyy",
				"foo":  "ddd",
			},
		}
	})
	tests.Add("get latest non-winning leaf", func(t *testing.T) interface{} {
		db := newDB(t)
		// common root doc
		_ = db.tPut("foo", map[string]interface{}{"foo": "aaa", "_rev": "1-aaa"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))
		// losing branch
		_ = db.tPut("foo", map[string]interface{}{
			"foo": "bbb",
			"_revisions": map[string]interface{}{
				"ids":   []string{"ccc", "bbb", "aaa"},
				"start": 3,
			},
		}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))

		// winning branch
		_ = db.tPut("foo", map[string]interface{}{
			"foo": "ddd",
			"_revisions": map[string]interface{}{
				"ids":   []string{"xxx", "yyy", "aaa"},
				"start": 3,
			},
		}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))

		return test{
			db: db,
			id: "foo",
			options: kivik.Params(map[string]interface{}{
				"latest": true,
				"rev":    "2-bbb",
			}),
			wantDoc: map[string]interface{}{
				"_id":  "foo",
				"_rev": "3-ccc",
				"foo":  "bbb",
			},
		}
	})
	tests.Add("get latest rev with deleted leaf, reverts to the winning branch", func(t *testing.T) interface{} {
		db := newDB(t)
		// common root doc
		_ = db.tPut("foo", map[string]interface{}{"foo": "aaa", "_rev": "1-aaa"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))
		// losing branch
		_ = db.tPut("foo", map[string]interface{}{
			"foo": "bbb",
			"_revisions": map[string]interface{}{
				"ids":   []string{"ccc", "bbb", "aaa"},
				"start": 3,
			},
		}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))
		// now delete the losing leaf
		_ = db.tDelete("foo", kivik.Rev("3-ccc"))

		// winning branch
		_ = db.tPut("foo", map[string]interface{}{
			"foo": "ddd",
			"_revisions": map[string]interface{}{
				"ids":   []string{"xxx", "yyy", "aaa"},
				"start": 3,
			},
		}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))

		return test{
			db: db,
			id: "foo",
			options: kivik.Params(map[string]interface{}{
				"latest": true,
				"rev":    "2-bbb",
			}),
			wantDoc: map[string]interface{}{
				"_id":  "foo",
				"_rev": "3-xxx",
				"foo":  "ddd",
			},
		}
	})
	tests.Add("revs=true, losing leaf", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]interface{}{"foo": "aaa", "_rev": "1-aaa"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))
		_ = db.tPut("foo", map[string]interface{}{
			"foo": "bbb",
			"_revisions": map[string]interface{}{
				"ids":   []string{"bbb", "aaa"},
				"start": 2,
			},
		}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))
		_ = db.tPut("foo", map[string]interface{}{
			"foo": "ddd",
			"_revisions": map[string]interface{}{
				"ids":   []string{"yyy", "aaa"},
				"start": 2,
			},
		}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))

		return test{
			db: db,
			id: "foo",
			options: kivik.Params(map[string]interface{}{
				"revs": true,
				"rev":  "2-bbb",
			}),
			wantDoc: map[string]interface{}{
				"_id":  "foo",
				"_rev": "2-bbb",
				"foo":  "bbb",
				"_revisions": map[string]interface{}{
					"start": 2,
					"ids":   []string{"bbb", "aaa"},
				},
			},
		}
	})
	tests.Add("local_seq=true", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]interface{}{"foo": "aaa", "_rev": "1-aaa"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))

		return test{
			db:      db,
			id:      "foo",
			options: kivik.Param("local_seq", true),
			wantDoc: map[string]interface{}{
				"_id":        "foo",
				"_rev":       "1-aaa",
				"foo":        "aaa",
				"_local_seq": float64(1),
			},
		}
	})
	tests.Add("local_seq=true & specified rev", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]interface{}{"foo": "aaa", "_rev": "1-aaa"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))

		return test{
			db:      db,
			id:      "foo",
			options: kivik.Params(map[string]interface{}{"local_seq": true, "rev": "1-aaa"}),
			wantDoc: map[string]interface{}{
				"_id":        "foo",
				"_rev":       "1-aaa",
				"foo":        "aaa",
				"_local_seq": float64(1),
			},
		}
	})
	tests.Add("local_seq=true & specified rev & latest=true", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]interface{}{"foo": "aaa", "_rev": "1-aaa"}, kivik.Params(map[string]interface{}{
			"new_edits": false,
		}))

		return test{
			db: db,
			id: "foo",
			options: kivik.Params(map[string]interface{}{
				"local_seq": true,
				"rev":       "1-aaa",
				"latest":    true,
			}),
			wantDoc: map[string]interface{}{
				"_id":        "foo",
				"_rev":       "1-aaa",
				"foo":        "aaa",
				"_local_seq": float64(1),
			},
		}
	})
	tests.Add("attachments=false, doc with attachments", func(t *testing.T) interface{} {
		db := newDB(t)
		rev := db.tPut("foo", map[string]interface{}{
			"foo":          "aaa",
			"_attachments": newAttachments().add("att.txt", "att.txt"),
		})

		return test{
			db:      db,
			id:      "foo",
			options: kivik.Param("attachments", false),
			wantDoc: map[string]interface{}{
				"_id":  "foo",
				"_rev": rev,
				"foo":  "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"content_type": "text/plain",
						"digest":       "md5-a4NyknGw7YOh+a5ezPdZ4A==",
						"length":       float64(7),
						"revpos":       float64(1),
						"stub":         true,
					},
				},
			},
		}
	})
	tests.Add("attachments=true, doc with attachments", func(t *testing.T) interface{} {
		db := newDB(t)
		rev := db.tPut("foo", map[string]interface{}{
			"foo":          "aaa",
			"_attachments": newAttachments().add("att.txt", "att.txt"),
		})

		return test{
			db:      db,
			id:      "foo",
			options: kivik.Param("attachments", true),
			wantDoc: map[string]interface{}{
				"_id":  "foo",
				"_rev": rev,
				"foo":  "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"content_type": "text/plain",
						"digest":       "md5-a4NyknGw7YOh+a5ezPdZ4A==",
						"length":       float64(7),
						"revpos":       float64(1),
						"data":         "YXR0LnR4dA==",
					},
				},
			},
		}
	})
	tests.Add("attachments=true, doc without attachments", func(t *testing.T) interface{} {
		db := newDB(t)
		rev := db.tPut("foo", map[string]interface{}{
			"foo": "aaa",
		})

		return test{
			db:      db,
			id:      "foo",
			options: kivik.Param("attachments", true),
			wantDoc: map[string]interface{}{
				"_id":  "foo",
				"_rev": rev,
				"foo":  "aaa",
			},
		}
	})
	tests.Add("attachments=false, do not return deleted attachments", func(t *testing.T) interface{} {
		db := newDB(t)
		rev1 := db.tPut("foo", map[string]interface{}{
			"foo": "aaa",
			"_attachments": newAttachments().
				add("att.txt", "att.txt").
				add("att2.txt", "att2.txt"),
		})
		rev2 := db.tPut("foo", map[string]interface{}{
			"foo":          "aaa",
			"_attachments": newAttachments().addStub("att.txt"),
		}, kivik.Rev(rev1))

		return test{
			db:      db,
			id:      "foo",
			options: kivik.Param("attachments", false),
			wantDoc: map[string]interface{}{
				"_id":  "foo",
				"_rev": rev2,
				"foo":  "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"content_type": "text/plain",
						"digest":       "md5-a4NyknGw7YOh+a5ezPdZ4A==",
						"revpos":       float64(1),
						"length":       float64(7),
						"stub":         true,
					},
				},
			},
		}
	})
	tests.Add("attachments=false, fetch atts added at different revs", func(t *testing.T) interface{} {
		db := newDB(t)
		rev1 := db.tPut("foo", map[string]interface{}{
			"foo":          "aaa",
			"_attachments": newAttachments().add("att.txt", "att.txt"),
		})
		rev2 := db.tPut("foo", map[string]interface{}{
			"foo": "aaa",
			"_attachments": newAttachments().
				addStub("att.txt").
				add("att2.txt", "att.txt"),
		}, kivik.Rev(rev1))

		return test{
			db:      db,
			id:      "foo",
			options: kivik.Param("attachments", false),
			wantDoc: map[string]interface{}{
				"_id":  "foo",
				"_rev": rev2,
				"foo":  "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"content_type": "text/plain",
						"digest":       "md5-a4NyknGw7YOh+a5ezPdZ4A==",
						"revpos":       float64(1),
						"length":       float64(7),
						"stub":         true,
					},
					"att2.txt": map[string]interface{}{
						"content_type": "text/plain",
						"digest":       "md5-a4NyknGw7YOh+a5ezPdZ4A==",
						"revpos":       float64(2),
						"length":       float64(7),
						"stub":         true,
					},
				},
			},
		}
	})
	tests.Add("attachments=false, fetch only atts that existed at time of specific rev", func(t *testing.T) interface{} {
		db := newDB(t)
		rev1 := db.tPut("foo", map[string]interface{}{
			"foo":          "aaa",
			"_attachments": newAttachments().add("att.txt", "att.txt"),
		})
		_ = db.tPut("foo", map[string]interface{}{
			"foo": "aaa",
			"_attachments": newAttachments().
				addStub("att.txt").
				add("att2.txt", "att.txt"),
		}, kivik.Rev(rev1))

		return test{
			db: db,
			id: "foo",
			options: kivik.Params(map[string]interface{}{
				"attachments": false,
				"rev":         rev1,
			}),
			wantDoc: map[string]interface{}{
				"_id":  "foo",
				"_rev": rev1,
				"foo":  "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"content_type": "text/plain",
						"digest":       "md5-a4NyknGw7YOh+a5ezPdZ4A==",
						"revpos":       float64(1),
						"length":       float64(7),
						"stub":         true,
					},
				},
			},
		}
	})
	tests.Add("attachments=false, fetch updated attachment", func(t *testing.T) interface{} {
		db := newDB(t)
		rev1 := db.tPut("foo", map[string]interface{}{
			"foo":          "aaa",
			"_attachments": newAttachments().add("att.txt", "att.txt"),
		})
		rev2 := db.tPut("foo", map[string]interface{}{
			"foo": "aaa",
			"_attachments": map[string]interface{}{
				"att.txt": map[string]interface{}{
					"data": "dmVyc2lvbiAyCg==",
				},
			},
		}, kivik.Rev(rev1))

		return test{
			db: db,
			id: "foo",
			options: kivik.Params(map[string]interface{}{
				"attachments": false,
			}),
			wantDoc: map[string]interface{}{
				"_id":  "foo",
				"_rev": rev2,
				"foo":  "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"content_type": "application/octet-stream",
						"digest":       "md5-sE0LKdS6wHgf6ETjKMXirA==",
						"revpos":       float64(2),
						"length":       float64(10),
						"stub":         true,
					},
				},
			},
		}
	})
	tests.Add("atts_since", func(t *testing.T) interface{} {
		db := newDB(t)
		rev1 := db.tPut("foo", map[string]interface{}{
			"foo":          "aaa",
			"_attachments": newAttachments().add("att.txt", "att.txt"),
		})
		rev2 := db.tPut("foo", map[string]interface{}{
			"foo": "aaa",
			"_attachments": newAttachments().
				addStub("att.txt").
				add("att2.txt", "second\n"),
		}, kivik.Rev(rev1))
		rev3 := db.tPut("foo", map[string]interface{}{
			"foo": "aaa",
			"_attachments": newAttachments().
				addStub("att.txt").
				addStub("att2.txt").
				add("att3.txt", "THREE\n"),
		}, kivik.Rev(rev2))
		rev4 := db.tPut("foo", map[string]interface{}{
			"foo": "aaa",
			"_attachments": newAttachments().
				addStub("att.txt").
				addStub("att2.txt").
				addStub("att3.txt").
				add("att4.txt", "IV\n"),
		}, kivik.Rev(rev3))

		return test{
			db:      db,
			id:      "foo",
			options: kivik.Param("atts_since", []string{rev1}),
			wantDoc: map[string]interface{}{
				"_id":  "foo",
				"_rev": rev4,
				"foo":  "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"content_type": "text/plain",
						"digest":       "md5-a4NyknGw7YOh+a5ezPdZ4A==",
						"revpos":       float64(1),
						"length":       float64(7),
						"stub":         true,
					},
					"att2.txt": map[string]interface{}{
						"content_type": "text/plain",
						"digest":       "md5-WdDRn8RcppIw2Fj2ClVX+A==",
						"revpos":       float64(2),
						"length":       float64(7),
						"data":         "c2Vjb25kCg==",
					},
					"att3.txt": map[string]interface{}{
						"content_type": "text/plain",
						"digest":       "md5-+kNHBLBGcQJJi0WTQ1EEsA==",
						"revpos":       float64(3),
						"length":       float64(6),
						"data":         "VEhSRUUK",
					},
					"att4.txt": map[string]interface{}{
						"content_type": "text/plain",
						"digest":       "md5-CVDeT3sXzxU0dMLAuMJKXA==",
						"revpos":       float64(4),
						"length":       float64(3),
						"data":         "SVYK",
					},
				},
			},
		}
	})
	tests.Add("atts_since with invalid rev format", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]interface{}{
			"foo":          "aaa",
			"_attachments": newAttachments().add("att.txt", "att.txt"),
		})

		return test{
			db:         db,
			id:         "foo",
			options:    kivik.Param("atts_since", []string{"this is an invalid rev"}),
			wantStatus: http.StatusBadRequest,
			wantErr:    `invalid rev format`,
		}
	})
	tests.Add("atts_since with non-existent rev", func(t *testing.T) interface{} {
		db := newDB(t)
		rev := db.tPut("foo", map[string]interface{}{
			"foo":          "aaa",
			"_attachments": newAttachments().add("att.txt", "att.txt"),
		})

		return test{
			db:      db,
			id:      "foo",
			options: kivik.Param("atts_since", []string{"1-asdfasdf"}),
			wantDoc: map[string]interface{}{
				"_id":  "foo",
				"_rev": rev,
				"foo":  "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"content_type": "text/plain",
						"digest":       "md5-a4NyknGw7YOh+a5ezPdZ4A==",
						"revpos":       float64(1),
						"length":       float64(7),
						"stub":         true,
					},
				},
			},
		}
	})
	tests.Add("after PutAttachment", func(t *testing.T) interface{} {
		db := newDB(t)
		rev1 := db.tPut("foo", map[string]string{
			"foo": "aaa",
		})
		att := driver.Attachment{
			ContentType: "text/plain",
			Filename:    "att.txt",
			Content:     io.NopCloser(strings.NewReader("test")),
		}

		rev2, err := db.PutAttachment(context.Background(), "foo", &att, kivik.Rev(rev1))
		if err != nil {
			t.Fatal(err)
		}

		return test{
			db: db,
			id: "foo",
			wantDoc: map[string]interface{}{
				"_id":  "foo",
				"_rev": rev2,
				"foo":  "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"content_type": "text/plain",
						"digest":       "md5-CY9rzUYh03PK3k6DJie09g==",
						"revpos":       float64(2),
						"length":       float64(4),
						"stub":         true,
					},
				},
			},
		}
	})

	/*
		TODO:
		att_encoding_info = true
	*/

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		db := tt.db
		if db == nil {
			db = newDB(t)
		}
		opts := tt.options
		if opts == nil {
			opts = mock.NilOption
		}
		doc, err := db.Get(context.Background(), tt.id, opts)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status: %d", status)
		}
		if err != nil {
			return
		}
		body, _ := io.ReadAll(doc.Body)
		var gotDoc interface{}
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&gotDoc); err != nil {
			t.Fatal(err)
		}
		if d := testy.DiffAsJSON(tt.wantDoc, gotDoc); d != nil {
			t.Errorf("Unexpected doc: %s", d)
		}
	})
}
