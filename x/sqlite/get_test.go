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
	t.Parallel()
	type test struct {
		setup      func(*testing.T, driver.DB)
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
	tests.Add("success", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		id: "foo",
		wantDoc: map[string]string{
			"_id":  "foo",
			"_rev": "1-9bb58f26192e4ba00f01e2e7b136bbd8",
			"foo":  "bar",
		},
	})
	tests.Add("get specific rev", test{
		setup: func(t *testing.T, d driver.DB) {
			rev, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "baz"}, kivik.Rev(rev))
			if err != nil {
				t.Fatal(err)
			}
		},
		id:      "foo",
		options: kivik.Rev("1-9bb58f26192e4ba00f01e2e7b136bbd8"),
		wantDoc: map[string]string{
			"_id":  "foo",
			"_rev": "1-9bb58f26192e4ba00f01e2e7b136bbd8",
			"foo":  "bar",
		},
	})
	tests.Add("specific rev not found", test{
		id:         "foo",
		options:    kivik.Rev("1-9bb58f26192e4ba00f01e2e7b136bbd8"),
		wantStatus: http.StatusNotFound,
		wantErr:    "not found",
	})
	tests.Add("include conflicts", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, kivik.Params(map[string]interface{}{
				"new_edits": false,
				"rev":       "1-abc",
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "baz"}, kivik.Params(map[string]interface{}{
				"new_edits": false,
				"rev":       "1-xyz",
			}))
			if err != nil {
				t.Fatal(err)
			}
		},
		id:      "foo",
		options: kivik.Param("conflicts", true),
		wantDoc: map[string]interface{}{
			"_id":        "foo",
			"_rev":       "1-xyz",
			"foo":        "baz",
			"_conflicts": []string{"1-abc"},
		},
	})
	tests.Add("include only leaf conflicts", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, kivik.Params(map[string]interface{}{
				"new_edits": false,
				"rev":       "1-abc",
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "baz"}, kivik.Params(map[string]interface{}{
				"new_edits": false,
				"rev":       "1-xyz",
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "qux"}, kivik.Rev("1-xyz"))
			if err != nil {
				t.Fatal(err)
			}
		},
		id:      "foo",
		options: kivik.Param("conflicts", true),
		wantDoc: map[string]interface{}{
			"_id":        "foo",
			"_rev":       "2-8ecd3d54a4d763ebc0b6e6666d9af066",
			"foo":        "qux",
			"_conflicts": []string{"1-abc"},
		},
	})
	tests.Add("deleted document", test{
		setup: func(t *testing.T, d driver.DB) {
			rev, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Delete(context.Background(), "foo", kivik.Rev(rev))
			if err != nil {
				t.Fatal(err)
			}
		},
		id:         "foo",
		wantStatus: http.StatusNotFound,
		wantErr:    "not found",
	})
	tests.Add("deleted document by rev", test{
		setup: func(t *testing.T, d driver.DB) {
			rev, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Delete(context.Background(), "foo", kivik.Rev(rev))
			if err != nil {
				t.Fatal(err)
			}
		},
		id:      "foo",
		options: kivik.Rev("2-df2a4fe30cde39c357c8d1105748d1b9"),
		wantDoc: map[string]interface{}{
			"_id":      "foo",
			"_rev":     "2-df2a4fe30cde39c357c8d1105748d1b9",
			"_deleted": true,
		},
	})
	tests.Add("deleted document with data by rev", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{"_deleted": true, "foo": "bar"}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		id:      "foo",
		options: kivik.Rev("1-6872a0fc474ada5c46ce054b92897063"),
		wantDoc: map[string]interface{}{
			"_id":      "foo",
			"_rev":     "1-6872a0fc474ada5c46ce054b92897063",
			"_deleted": true,
			"foo":      "bar",
		},
	})
	tests.Add("include conflicts, skip deleted conflicts", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{"foo": "moo", "_deleted": true}, kivik.Params(map[string]interface{}{
				"new_edits": false,
				"rev":       "1-qwe",
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, kivik.Params(map[string]interface{}{
				"new_edits": false,
				"rev":       "1-abc",
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "baz"}, kivik.Params(map[string]interface{}{
				"new_edits": false,
				"rev":       "1-xyz",
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "qux"}, kivik.Rev("1-xyz"))
			if err != nil {
				t.Fatal(err)
			}
		},
		id:      "foo",
		options: kivik.Param("conflicts", true),
		wantDoc: map[string]interface{}{
			"_id":        "foo",
			"_rev":       "2-8ecd3d54a4d763ebc0b6e6666d9af066",
			"foo":        "qux",
			"_conflicts": []string{"1-abc"},
		},
	})
	tests.Add("include deleted conflicts", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{"foo": "moo", "_deleted": true}, kivik.Params(map[string]interface{}{
				"new_edits": false,
				"rev":       "1-qwe",
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, kivik.Params(map[string]interface{}{
				"new_edits": false,
				"rev":       "1-abc",
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "baz"}, kivik.Params(map[string]interface{}{
				"new_edits": false,
				"rev":       "1-xyz",
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "qux"}, kivik.Rev("1-xyz"))
			if err != nil {
				t.Fatal(err)
			}
		},
		id:      "foo",
		options: kivik.Param("deleted_conflicts", true),
		wantDoc: map[string]interface{}{
			"_id":                "foo",
			"_rev":               "2-8ecd3d54a4d763ebc0b6e6666d9af066",
			"foo":                "qux",
			"_deleted_conflicts": []string{"1-qwe"},
		},
	})
	tests.Add("include all conflicts", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{"foo": "moo", "_deleted": true}, kivik.Params(map[string]interface{}{
				"new_edits": false,
				"rev":       "1-qwe",
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, kivik.Params(map[string]interface{}{
				"new_edits": false,
				"rev":       "1-abc",
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "baz"}, kivik.Params(map[string]interface{}{
				"new_edits": false,
				"rev":       "1-xyz",
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "qux"}, kivik.Rev("1-xyz"))
			if err != nil {
				t.Fatal(err)
			}
		},
		id: "foo",
		options: kivik.Params(map[string]interface{}{
			"conflicts":         true,
			"deleted_conflicts": true,
		}),
		wantDoc: map[string]interface{}{
			"_id":                "foo",
			"_rev":               "2-8ecd3d54a4d763ebc0b6e6666d9af066",
			"foo":                "qux",
			"_deleted_conflicts": []string{"1-qwe"},
			"_conflicts":         []string{"1-abc"},
		},
	})
	tests.Add("include revs_info", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{"foo": "moo", "_deleted": true}, kivik.Params(map[string]interface{}{
				"new_edits": false,
				"rev":       "1-qwe",
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, kivik.Params(map[string]interface{}{
				"new_edits": false,
				"rev":       "1-abc",
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "baz"}, kivik.Params(map[string]interface{}{
				"new_edits": false,
				"rev":       "1-xyz",
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "qux"}, kivik.Rev("1-xyz"))
			if err != nil {
				t.Fatal(err)
			}
		},
		id: "foo",
		options: kivik.Params(map[string]interface{}{
			"revs_info": true,
		}),
		wantDoc: map[string]interface{}{
			"_id":  "foo",
			"_rev": "2-8ecd3d54a4d763ebc0b6e6666d9af066",
			"foo":  "qux",
			"_revs_info": []map[string]string{
				{"rev": "2-8ecd3d54a4d763ebc0b6e6666d9af066", "status": "available"},
				{"rev": "1-xyz", "status": "available"},
			},
		},
	})
	tests.Add("include meta", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{"foo": "moo", "_deleted": true}, kivik.Params(map[string]interface{}{
				"new_edits": false,
				"rev":       "1-qwe",
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, kivik.Params(map[string]interface{}{
				"new_edits": false,
				"rev":       "1-abc",
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "baz"}, kivik.Params(map[string]interface{}{
				"new_edits": false,
				"rev":       "1-xyz",
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]string{"foo": "qux"}, kivik.Rev("1-xyz"))
			if err != nil {
				t.Fatal(err)
			}
		},
		id:      "foo",
		options: kivik.Param("meta", true),
		wantDoc: map[string]interface{}{
			"_id":  "foo",
			"_rev": "2-8ecd3d54a4d763ebc0b6e6666d9af066",
			"foo":  "qux",
			"_revs_info": []map[string]string{
				{"rev": "2-8ecd3d54a4d763ebc0b6e6666d9af066", "status": "available"},
				{"rev": "1-xyz", "status": "available"},
			},
			"_conflicts":         []string{"1-abc"},
			"_deleted_conflicts": []string{"1-qwe"},
		},
	})
	tests.Add("get latest winning leaf", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{"foo": "aaa", "_rev": "1-aaa"}, kivik.Params(map[string]interface{}{
				"new_edits": false,
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "bbb",
				"_revisions": map[string]interface{}{
					"ids":   []string{"bbb", "aaa"},
					"start": 2,
				},
			}, kivik.Params(map[string]interface{}{
				"new_edits": false,
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "ddd",
				"_revisions": map[string]interface{}{
					"ids":   []string{"yyy", "aaa"},
					"start": 2,
				},
			}, kivik.Params(map[string]interface{}{
				"new_edits": false,
			}))
			if err != nil {
				t.Fatal(err)
			}
		},
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
	})
	tests.Add("get latest non-winning leaf", test{
		setup: func(t *testing.T, d driver.DB) {
			// common root doc
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{"foo": "aaa", "_rev": "1-aaa"}, kivik.Params(map[string]interface{}{
				"new_edits": false,
			}))
			if err != nil {
				t.Fatal(err)
			}
			// losing branch
			_, err = d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "bbb",
				"_revisions": map[string]interface{}{
					"ids":   []string{"ccc", "bbb", "aaa"},
					"start": 3,
				},
			}, kivik.Params(map[string]interface{}{
				"new_edits": false,
			}))
			if err != nil {
				t.Fatal(err)
			}

			// winning branch
			_, err = d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "ddd",
				"_revisions": map[string]interface{}{
					"ids":   []string{"xxx", "yyy", "aaa"},
					"start": 3,
				},
			}, kivik.Params(map[string]interface{}{
				"new_edits": false,
			}))
			if err != nil {
				t.Fatal(err)
			}
		},
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
	})
	tests.Add("get latest rev with deleted leaf, reverts to the winning branch", test{
		setup: func(t *testing.T, d driver.DB) {
			// common root doc
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{"foo": "aaa", "_rev": "1-aaa"}, kivik.Params(map[string]interface{}{
				"new_edits": false,
			}))
			if err != nil {
				t.Fatal(err)
			}
			// losing branch
			_, err = d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "bbb",
				"_revisions": map[string]interface{}{
					"ids":   []string{"ccc", "bbb", "aaa"},
					"start": 3,
				},
			}, kivik.Params(map[string]interface{}{
				"new_edits": false,
			}))
			if err != nil {
				t.Fatal(err)
			}
			// now delete the losing leaf
			_, err = d.Delete(context.Background(), "foo", kivik.Rev("3-ccc"))
			if err != nil {
				t.Fatal(err)
			}

			// winning branch
			_, err = d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "ddd",
				"_revisions": map[string]interface{}{
					"ids":   []string{"xxx", "yyy", "aaa"},
					"start": 3,
				},
			}, kivik.Params(map[string]interface{}{
				"new_edits": false,
			}))
			if err != nil {
				t.Fatal(err)
			}
		},
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
	})
	tests.Add("revs=true, losing leaf", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{"foo": "aaa", "_rev": "1-aaa"}, kivik.Params(map[string]interface{}{
				"new_edits": false,
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "bbb",
				"_revisions": map[string]interface{}{
					"ids":   []string{"bbb", "aaa"},
					"start": 2,
				},
			}, kivik.Params(map[string]interface{}{
				"new_edits": false,
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "ddd",
				"_revisions": map[string]interface{}{
					"ids":   []string{"yyy", "aaa"},
					"start": 2,
				},
			}, kivik.Params(map[string]interface{}{
				"new_edits": false,
			}))
			if err != nil {
				t.Fatal(err)
			}
		},
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
	})
	tests.Add("local_seq=true", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{"foo": "aaa", "_rev": "1-aaa"}, kivik.Params(map[string]interface{}{
				"new_edits": false,
			}))
			if err != nil {
				t.Fatal(err)
			}
		},
		id:      "foo",
		options: kivik.Param("local_seq", true),
		wantDoc: map[string]interface{}{
			"_id":        "foo",
			"_rev":       "1-aaa",
			"foo":        "aaa",
			"_local_seq": float64(1),
		},
	})
	tests.Add("local_seq=true & specified rev", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{"foo": "aaa", "_rev": "1-aaa"}, kivik.Params(map[string]interface{}{
				"new_edits": false,
			}))
			if err != nil {
				t.Fatal(err)
			}
		},
		id:      "foo",
		options: kivik.Params(map[string]interface{}{"local_seq": true, "rev": "1-aaa"}),
		wantDoc: map[string]interface{}{
			"_id":        "foo",
			"_rev":       "1-aaa",
			"foo":        "aaa",
			"_local_seq": float64(1),
		},
	})
	tests.Add("local_seq=true & specified rev & latest=true", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{"foo": "aaa", "_rev": "1-aaa"}, kivik.Params(map[string]interface{}{
				"new_edits": false,
			}))
			if err != nil {
				t.Fatal(err)
			}
		},
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
	})
	tests.Add("attachments=false, doc with attachments", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"content_type": "text/plain",
						"data":         "YXR0LnR4dA==",
					},
				},
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		id:      "foo",
		options: kivik.Param("attachments", false),
		wantDoc: map[string]interface{}{
			"_id":  "foo",
			"_rev": "1-5f3e7150f872a1dd295f44b1e4a9fa41",
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
	})
	tests.Add("attachments=true, doc with attachments", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"content_type": "text/plain",
						"data":         "YXR0LnR4dA==",
					},
				},
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		id:      "foo",
		options: kivik.Param("attachments", true),
		wantDoc: map[string]interface{}{
			"_id":  "foo",
			"_rev": "1-5f3e7150f872a1dd295f44b1e4a9fa41",
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
	})
	tests.Add("attachments=true, doc without attachments", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "aaa",
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		id:      "foo",
		options: kivik.Param("attachments", true),
		wantDoc: map[string]interface{}{
			"_id":  "foo",
			"_rev": "1-8655eafbc9513d4857258c6d48f40399",
			"foo":  "aaa",
		},
	})
	tests.Add("attachments=false, do not return deleted attachments", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"content_type": "text/plain",
						"data":         "YXR0LnR4dA==",
					},
					"att2.txt": map[string]interface{}{
						"content_type": "text/plain",
						"data":         "YXR0LnR4dA==",
					},
				},
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"stub": true,
					},
				},
			}, kivik.Rev("1-a4791acb8fcc7d205b4c582e0c9e3dc0"))
			if err != nil {
				t.Fatal(err)
			}
		},
		id:      "foo",
		options: kivik.Param("attachments", false),
		wantDoc: map[string]interface{}{
			"_id":  "foo",
			"_rev": "2-683854351c2ead3ccc353da6980070c4",
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
	})
	tests.Add("attachments=false, fetch atts added at different revs", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"content_type": "text/plain",
						"data":         "YXR0LnR4dA==",
					},
				},
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"stub": true,
					},
					"att2.txt": map[string]interface{}{
						"content_type": "text/plain",
						"data":         "YXR0LnR4dA==",
					},
				},
			}, kivik.Rev("1-5f3e7150f872a1dd295f44b1e4a9fa41"))
			if err != nil {
				t.Fatal(err)
			}
		},
		id:      "foo",
		options: kivik.Param("attachments", false),
		wantDoc: map[string]interface{}{
			"_id":  "foo",
			"_rev": "2-9dc3adaa7b08ac0ae246cded87669883",
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
	})
	tests.Add("attachments=false, fetch only atts that existed at time of specific rev", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"content_type": "text/plain",
						"data":         "YXR0LnR4dA==",
					},
				},
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"stub": true,
					},
					"att2.txt": map[string]interface{}{
						"content_type": "text/plain",
						"data":         "YXR0LnR4dA==",
					},
				},
			}, kivik.Rev("1-5f3e7150f872a1dd295f44b1e4a9fa41"))
			if err != nil {
				t.Fatal(err)
			}
		},
		id: "foo",
		options: kivik.Params(map[string]interface{}{
			"attachments": false,
			"rev":         "1-5f3e7150f872a1dd295f44b1e4a9fa41",
		}),
		wantDoc: map[string]interface{}{
			"_id":  "foo",
			"_rev": "1-5f3e7150f872a1dd295f44b1e4a9fa41",
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
	})
	tests.Add("attachments=false, fetch updated attachment", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"content_type": "text/plain",
						"data":         "YXR0LnR4dA==",
					},
				},
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"data": "dmVyc2lvbiAyCg==",
					},
				},
			}, kivik.Rev("1-5f3e7150f872a1dd295f44b1e4a9fa41"))
			if err != nil {
				t.Fatal(err)
			}
		},
		id: "foo",
		options: kivik.Params(map[string]interface{}{
			"attachments": false,
		}),
		wantDoc: map[string]interface{}{
			"_id":  "foo",
			"_rev": "2-f8b1930b7bdd8d48f4701188f999d326",
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
	})
	tests.Add("atts_since", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"content_type": "text/plain",
						"data":         "YXR0LnR4dA==",
					},
				},
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"stub": true,
					},
					"att2.txt": map[string]interface{}{
						"content_type": "text/plain",
						"data":         "YXR0LnR4dA==",
					},
				},
			}, kivik.Rev("1-5f3e7150f872a1dd295f44b1e4a9fa41"))
			if err != nil {
				t.Fatal(err)
			}
		},
		id:      "foo",
		options: kivik.Param("atts_since", []string{"1-5f3e7150f872a1dd295f44b1e4a9fa41"}),
		wantDoc: map[string]interface{}{
			"_id":  "foo",
			"_rev": "2-9dc3adaa7b08ac0ae246cded87669883",
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
					"data":         "YXR0LnR4dA==",
				},
			},
		},
	})
	tests.Add("atts_since with invalid rev format", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"content_type": "text/plain",
						"data":         "YXR0LnR4dA==",
					},
				},
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		id:         "foo",
		options:    kivik.Param("atts_since", []string{"this is an invalid rev"}),
		wantStatus: http.StatusBadRequest,
		wantErr:    `strconv.ParseInt: parsing "this is an invalid rev": invalid syntax`,
	})
	tests.Add("atts_since with non-existent rev", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]interface{}{
				"foo": "aaa",
				"_attachments": map[string]interface{}{
					"att.txt": map[string]interface{}{
						"content_type": "text/plain",
						"data":         "YXR0LnR4dA==",
					},
				},
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		id:      "foo",
		options: kivik.Param("atts_since", []string{"1-asdfasdf"}),
		wantDoc: map[string]interface{}{
			"_id":  "foo",
			"_rev": "1-5f3e7150f872a1dd295f44b1e4a9fa41",
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
	})
	tests.Add("after PutAttachment", test{
		setup: func(t *testing.T, d driver.DB) {
			_, err := d.Put(context.Background(), "foo", map[string]string{
				"foo": "aaa",
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
			att := driver.Attachment{
				ContentType: "text/plain",
				Filename:    "att.txt",
				Content:     io.NopCloser(strings.NewReader("test")),
			}

			_, err = d.PutAttachment(context.Background(), "foo", &att, kivik.Rev("1-8655eafbc9513d4857258c6d48f40399"))
			if err != nil {
				t.Fatal(err)
			}
		},
		id: "foo",
		wantDoc: map[string]interface{}{
			"_id":  "foo",
			"_rev": "2-8655eafbc9513d4857258c6d48f40399",
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
	})

	/*
		TODO:
		att_encoding_info = true
		open_revs = [] // TODO: driver.OpenRever
	*/

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		db := newDB(t)
		if tt.setup != nil {
			tt.setup(t, db)
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
		var gotDoc interface{}
		if err := json.NewDecoder(doc.Body).Decode(&gotDoc); err != nil {
			t.Fatal(err)
		}
		if d := testy.DiffAsJSON(tt.wantDoc, gotDoc); d != nil {
			t.Errorf("Unexpected doc: %s", d)
		}
	})
}
