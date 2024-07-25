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

package sqlite

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-kivik/kivik/v4"
)

func Test_isLeafRev(t *testing.T) {
	t.Run("Doc not found", func(t *testing.T) {
		d := newDB(t)
		tx, err := d.underlying().Begin()
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback()
		_, err = d.DB.(*db).isLeafRev(context.Background(), tx, "foo", 1, "abc")
		status := kivik.HTTPStatus(err)
		if status != http.StatusNotFound {
			t.Errorf("Expected %d, got %d", http.StatusNotFound, status)
		}
	})
	t.Run("doc exists, but missing rev provided returns conflict", func(t *testing.T) {
		d := newDB(t)

		// setup
		_ = d.tPut("foo", map[string]string{"_id": "foo"})

		tx, err := d.underlying().Begin()
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback()
		_, err = d.DB.(*db).isLeafRev(context.Background(), tx, "foo", 3, "abc")
		status := kivik.HTTPStatus(err)
		if status != http.StatusConflict {
			t.Errorf("Expected %d, got %d", http.StatusConflict, status)
		}
	})
	t.Run("Not a leaf revision", func(t *testing.T) {
		d := newDB(t)

		// setup
		rev := d.tPut("foo", map[string]string{"_id": "foo"})
		_ = d.tPut("foo", map[string]string{"_id": "foo"}, kivik.Rev(rev))

		r, _ := parseRev(rev)

		tx, err := d.underlying().Begin()
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback()
		_, err = d.DB.(*db).isLeafRev(context.Background(), tx, "foo", r.rev, r.id)
		status := kivik.HTTPStatus(err)
		if status != http.StatusConflict {
			t.Errorf("Expected %d, got %d", http.StatusConflict, status)
		}
	})
	t.Run("Is a leaf revision", func(t *testing.T) {
		d := newDB(t)

		// setup
		rev := d.tPut("foo", map[string]string{"_id": "foo"})
		rev2 := d.tPut("foo", map[string]string{"_id": "foo"}, kivik.Rev(rev))

		r, _ := parseRev(rev2)

		tx, err := d.underlying().Begin()
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback()
		_, err = d.DB.(*db).isLeafRev(context.Background(), tx, "foo", r.rev, r.id)
		if err != nil {
			t.Fatal(err)
		}
	})
}
