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
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

func TestDBGet(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		setup      func(*testing.T, driver.DB)
		id         string
		options    driver.Options
		wantDoc    interface{}
		wantStatus int
		wantErr    string
	}{
		{
			name:       "not found",
			id:         "foo",
			wantStatus: http.StatusNotFound,
			wantErr:    "not found",
		},
		{
			name: "success",
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
		},
		{
			name: "get specific rev",
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
		},
		{
			name:       "specific rev not found",
			id:         "foo",
			options:    kivik.Rev("1-9bb58f26192e4ba00f01e2e7b136bbd8"),
			wantStatus: http.StatusNotFound,
			wantErr:    "not found",
		},
		{
			name: "include conflicts",
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
		},
		{
			name: "include only leaf conflicts",
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
		},
		{
			name: "deleted document",
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
		},
		{
			name: "deleted document by rev",
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
		},
		{
			name: "deleted document with data by rev",
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
		},
		{
			name: "include conflicts, skip deleted conflicts",
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
		},
		{
			name: "include deleted conflicts",
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
		},
		{
			name: "include all conflicts",
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
		},
		{
			name: "include revs_info",
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
		},
		{
			name: "include meta",
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
		},
		{
			name: "get latest winning leaf",
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
		},
		{
			name: "get latest non-winning leaf",
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
		},
		{
			name: "get latest rev with deleted leaf, reverts to the winning branch",
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
		},
		{
			name: "revs=true, losing leaf",
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
		},
		{
			name: "local_seq=true",
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
		},
		{
			name: "local_seq=true & specified rev",
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
		},
		{
			name: "local_seq=true & specified rev & latest=true",
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
		},
		{
			name: "attachments=false, doc with attachments",
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
		},
		{
			name: "attachments=true, doc with attachments",
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
		},
		{
			name: "attachments=true, doc without attachments",
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
		},
		{
			name: "attachments=false, do not return deleted attachments",
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
		},
		/*
			TODO:
			attachments = true
				- with specific rev, exclude newer attachments
				- exclude deleted attachments
				- fetch correct version of attachment
			att_encoding_info = true
			atts_since = [revs]
			open_revs = [] // TODO: driver.OpenRever
		*/
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
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
}
