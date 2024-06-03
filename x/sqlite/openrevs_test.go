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
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestDBOpenRevs(t *testing.T) {
	t.Parallel()
	type test struct {
		db         *testDB
		docID      string
		revs       []string
		options    driver.Options
		want       []rowResult
		wantErr    string
		wantStatus int
	}
	tests := testy.NewTable()
	tests.Add("all revs, document not found", test{
		docID:      "not there",
		revs:       []string{"all"},
		wantErr:    "missing",
		wantStatus: http.StatusNotFound,
	})
	tests.Add("invalid rev format", func(t *testing.T) interface{} {
		d := newDB(t)
		docID := "foo"
		_ = d.tPut(docID, map[string]string{"foo": "bar"})

		return test{
			db:         d,
			docID:      docID,
			revs:       []string{"oink", "all"},
			wantErr:    "invalid rev format",
			wantStatus: http.StatusBadRequest,
		}
	})
	tests.Add("all, with single rev", func(t *testing.T) interface{} {
		d := newDB(t)
		docID := "foo"
		rev := d.tPut(docID, map[string]string{"foo": "bar"})

		return test{
			db:    d,
			docID: docID,
			revs:  []string{"all"},
			want: []rowResult{
				{ID: docID, Rev: rev, Doc: `{"_id":"` + docID + `","_rev":"` + rev + `","foo":"bar"}`},
			},
		}
	})
	tests.Add("all, with deleted rev", func(t *testing.T) interface{} {
		d := newDB(t)
		docID := "foo"
		rev := d.tPut(docID, map[string]string{"foo": "bar"})
		rev2 := d.tDelete(docID, kivik.Rev(rev))

		return test{
			db:    d,
			docID: docID,
			revs:  []string{"all"},
			want: []rowResult{
				{ID: docID, Rev: rev2, Doc: `{"_id":"` + docID + `","_rev":"` + rev2 + `","_deleted":true}`},
			},
		}
	})
	tests.Add("all, with conflicting leaves", func(t *testing.T) interface{} {
		d := newDB(t)
		docID := "foo"
		rev := d.tPut(docID, map[string]string{"_rev": "1-xyz", "foo": "bar"}, kivik.Param("new_edits", false))
		rev2 := d.tPut(docID, map[string]string{"_rev": "1-abc", "foo": "baz"}, kivik.Param("new_edits", false))

		return test{
			db:    d,
			docID: docID,
			revs:  []string{"all"},
			want: []rowResult{
				{ID: docID, Rev: rev2, Doc: `{"_id":"` + docID + `","_rev":"` + rev2 + `","foo":"baz"}`},
				{ID: docID, Rev: rev, Doc: `{"_id":"` + docID + `","_rev":"` + rev + `","foo":"bar"}`},
			},
		}
	})
	tests.Add("no revs provided returns winning leaf", func(t *testing.T) interface{} {
		d := newDB(t)
		docID := "foo"
		rev := d.tPut(docID, map[string]string{"_rev": "1-xyz", "foo": "bar"}, kivik.Param("new_edits", false))
		_ = d.tPut(docID, map[string]string{"_rev": "1-abc", "foo": "baz"}, kivik.Param("new_edits", false))

		return test{
			db:    d,
			docID: docID,
			revs:  []string{},
			want: []rowResult{
				{ID: docID, Rev: rev, Doc: `{"_id":"` + docID + `","_rev":"` + rev + `","foo":"bar"}`},
			},
		}
	})
	tests.Add("specific rev", func(t *testing.T) interface{} {
		d := newDB(t)
		docID := "foo"
		rev := d.tPut(docID, map[string]string{"foo": "bar"})
		rev2 := d.tPut(docID, map[string]string{"foo": "baz"}, kivik.Rev(rev))

		return test{
			db:    d,
			docID: docID,
			revs:  []string{rev2},
			want: []rowResult{
				{ID: docID, Rev: rev2, Doc: `{"_id":"` + docID + `","_rev":"` + rev2 + `","foo":"baz"}`},
			},
		}
	})
	tests.Add("specific revs, including non-leaf revs", func(t *testing.T) interface{} {
		d := newDB(t)
		docID := "foo"
		rev := d.tPut(docID, map[string]string{"foo": "bar"})
		rev2 := d.tPut(docID, map[string]string{"foo": "baz"}, kivik.Rev(rev))

		return test{
			db:    d,
			docID: docID,
			revs:  []string{rev, rev2},
			want: []rowResult{
				{ID: docID, Rev: rev, Doc: `{"_id":"` + docID + `","_rev":"` + rev + `","foo":"bar"}`},
				{ID: docID, Rev: rev2, Doc: `{"_id":"` + docID + `","_rev":"` + rev2 + `","foo":"baz"}`},
			},
		}
	})
	tests.Add("specific revs, including one that doesn't exist", func(t *testing.T) interface{} {
		d := newDB(t)
		docID := "foo"
		rev := d.tPut(docID, map[string]string{"foo": "bar"})
		rev2 := d.tPut(docID, map[string]string{"foo": "baz"}, kivik.Rev(rev))

		return test{
			db:    d,
			docID: docID,
			revs:  []string{rev, rev2, "99-asdf"},
			want: []rowResult{
				{ID: docID, Rev: rev, Doc: `{"_id":"` + docID + `","_rev":"` + rev + `","foo":"bar"}`},
				{ID: docID, Rev: rev2, Doc: `{"_id":"` + docID + `","_rev":"` + rev2 + `","foo":"baz"}`},
				{ID: docID, Rev: "99-asdf", Error: "missing"},
			},
		}
	})
	tests.Add("latest=true returns latest leaf of conflicting branch", func(t *testing.T) interface{} {
		d := newDB(t)
		docID := "foo"
		rev := d.tPut(docID, map[string]string{"step": "one"})
		r, _ := parseRev(rev)
		_ = d.tPut(docID, map[string]interface{}{
			"step": "two",
			"_revisions": map[string]interface{}{
				"start": 2,
				"ids":   []string{"abc", r.id},
			},
		}, kivik.Param("new_edits", false))
		_ = d.tPut(docID, map[string]interface{}{
			"step": "three",
			"_revisions": map[string]interface{}{
				"start": 3,
				"ids":   []string{"def", "abc"},
			},
		}, kivik.Param("new_edits", false))
		_ = d.tPut(docID, map[string]interface{}{
			"step": "four",
			"_revisions": map[string]interface{}{
				"start": 3,
				"ids":   []string{"jkl", "ghi", r.id},
			},
		}, kivik.Param("new_edits", false))

		return test{
			db:      d,
			docID:   docID,
			revs:    []string{"2-abc"},
			options: kivik.Param("latest", true),
			want: []rowResult{
				{ID: docID, Rev: "2-abc", Doc: `{"_id":"` + docID + `","_rev":"2-abc","step":"two"}`},
				{ID: docID, Rev: "3-def", Doc: `{"_id":"` + docID + `","_rev":"3-def","step":"three"}`},
			},
		}
	})
	tests.Add("latest=true returns multiple leaves when requested ancestor has them", func(t *testing.T) interface{} {
		d := newDB(t)
		docID := "foo"
		rev := d.tPut(docID, map[string]string{"step": "one"})
		r, _ := parseRev(rev)
		_ = d.tPut(docID, map[string]interface{}{
			"step": "three",
			"_revisions": map[string]interface{}{
				"start": 3,
				"ids":   []string{"def", "abc", r.id},
			},
		}, kivik.Param("new_edits", false))
		_ = d.tPut(docID, map[string]interface{}{
			"step": "four",
			"_revisions": map[string]interface{}{
				"start": 3,
				"ids":   []string{"jkl", "ghi", r.id},
			},
		}, kivik.Param("new_edits", false))

		return test{
			db:      d,
			docID:   docID,
			revs:    []string{rev},
			options: kivik.Param("latest", true),
			want: []rowResult{
				{ID: docID, Rev: rev, Doc: `{"_id":"` + docID + `","_rev":"` + rev + `","step":"one"}`},
				{ID: docID, Rev: "3-def", Doc: `{"_id":"` + docID + `","_rev":"3-def","step":"three"}`},
				{ID: docID, Rev: "3-jkl", Doc: `{"_id":"` + docID + `","_rev":"3-jkl","step":"four"}`},
			},
		}
	})
	tests.Add("specific rev and revs=true", func(t *testing.T) interface{} {
		d := newDB(t)
		docID := "foo"
		rev := d.tPut(docID, map[string]string{"foo": "bar"})
		rev2 := d.tPut(docID, map[string]string{"foo": "baz"}, kivik.Rev(rev))
		rev3 := d.tPut(docID, map[string]string{"foo": "qux"}, kivik.Rev(rev2))

		r, _ := parseRev(rev)
		r2, _ := parseRev(rev2)
		r3, _ := parseRev(rev3)

		return test{
			db:      d,
			docID:   docID,
			revs:    []string{rev3},
			options: kivik.Param("revs", true),
			want: []rowResult{
				{ID: docID, Rev: rev3, Doc: `{"_id":"` + docID + `","_rev":"` + rev3 + `","foo":"qux","_revisions":{"start":3,"ids":["` + r3.id + `","` + r2.id + `","` + r.id + `"]}}`},
			},
		}
	})
	tests.Add("revs=true with a single rev only", func(t *testing.T) interface{} {
		d := newDB(t)
		docID := "foo"
		rev := d.tPut(docID, map[string]interface{}{
			"foo": "bar",
		})

		r, _ := parseRev(rev)

		return test{
			db:      d,
			docID:   docID,
			revs:    []string{rev},
			options: kivik.Param("revs", true),
			want: []rowResult{
				{ID: docID, Rev: rev, Doc: `{"_id":"` + docID + `","_rev":"` + rev + `","foo":"bar","_revisions":{"start":1,"ids":["` + r.id + `"]}}`},
			},
		}
	})

	tests.Add("include attachment data", func(t *testing.T) interface{} {
		d := newDB(t)
		docID := "foo"
		rev := d.tPut(docID, map[string]interface{}{
			"foo": "bar",
			"_attachments": newAttachments().
				add("att.txt", "hello, world").
				add("att2.txt", "goodbye, world"),
		})

		return test{
			db:    d,
			docID: docID,
			revs:  []string{rev},
			want: []rowResult{
				{ID: docID, Rev: rev, Doc: `{"_id":"` + docID + `","_rev":"` + rev + `","foo":"bar","_attachments":{"att.txt":{"content_type":"text/plain","digest":"md5-5NfxtO0uQtFYmPSyewGdpA==","length":12,"revpos":1,"data":"aGVsbG8sIHdvcmxk"},"att2.txt":{"content_type":"text/plain","digest":"md5-18/CjiFEUaAYOzOUOD2UPQ==","length":14,"revpos":1,"data":"Z29vZGJ5ZSwgd29ybGQ="}}}`},
			},
		}
	})
	/*
		TODO:
		- revs=true + attachments
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

		rows, err := db.OpenRevs(context.Background(), tt.docID, tt.revs, opts)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status: %d", status)
		}
		if err != nil {
			return
		}

		checkRows(t, rows, tt.want)
	})
}
