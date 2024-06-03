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

package sqlite

import (
	"context"
	"net/http"
	"regexp"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestDBCreateDoc(t *testing.T) {
	t.Parallel()
	type test struct {
		db   *testDB
		doc  interface{}
		opts driver.Options

		wantDocID  string
		wantRev    string
		check      func(*testing.T)
		wantErr    string
		wantStatus int
	}

	tests := testy.NewTable()
	tests.Add("create doc with specified doc id", test{
		doc:       map[string]string{"_id": "foo"},
		wantDocID: "foo",
		wantRev:   "1-" + (&docData{ID: "foo", Doc: []byte(`{}`)}).RevID(),
	})
	tests.Add("create doc with different doc id", test{
		doc:       map[string]string{"_id": "bar"},
		wantDocID: "bar",
		wantRev:   "1-.*",
	})
	tests.Add("invalid document should return 400", test{
		doc:        make(chan int),
		wantErr:    "json: unsupported type: chan int",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("doc without ID", func(t *testing.T) interface{} {
		db := newDB(t)

		return test{
			db:        db,
			doc:       map[string]interface{}{"foo": "bar"},
			wantDocID: "(?i)^[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12}$",
			wantRev:   "1-.*",
			check: func(t *testing.T) {
				var docID, doc string
				err := db.underlying().QueryRow(`
					SELECT id, doc
					FROM test
					WHERE rev=1
				`).Scan(&docID, &doc)
				if err != nil {
					t.Fatalf("Failed to query for doc: %s", err)
				}
				if docID == "" {
					t.Errorf("Expected doc ID, got empty string")
				}
				if doc != `{"foo":"bar"}` {
					t.Errorf("Unexpected doc content: %s", doc)
				}
			},
		}
	})
	tests.Add("creating doc with existing id causes conflict", func(t *testing.T) interface{} {
		db := newDB(t)
		_ = db.tPut("foo", map[string]string{"foo": "bar"})

		return test{
			db:         db,
			doc:        map[string]string{"_id": "foo"},
			wantErr:    "document update conflict",
			wantStatus: http.StatusConflict,
		}
	})
	tests.Add("create doc with attachment", func(t *testing.T) interface{} {
		db := newDB(t)

		return test{
			db: db,
			doc: map[string]interface{}{
				"foo":          "bar",
				"_attachments": newAttachments().add("foo.txt", "bar"),
			},
			wantDocID: "(?i)^[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12}$",
			wantRev:   "1-.*",
			check: func(t *testing.T) {
				var attCount int
				err := db.underlying().QueryRow(`
					SELECT COUNT(*)
					FROM test_attachments
				`).Scan(&attCount)
				if err != nil {
					t.Fatalf("Failed to query for doc: %s", err)
				}
				if attCount != 1 {
					t.Errorf("Expected 1 attachment, got %d", attCount)
				}
			},
		}
	})
	tests.Add("create a deleted document", func(t *testing.T) interface{} {
		db := newDB(t)

		return test{
			db: db,
			doc: map[string]interface{}{
				"foo":      "bar",
				"_deleted": true,
			},
			wantDocID: "(?i)^[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12}$",
			wantRev:   "1-.*",
			check: func(t *testing.T) {
				var attCount int
				err := db.underlying().QueryRow(`
					SELECT COUNT(*)
					FROM test
					WHERE deleted=true
				`).Scan(&attCount)
				if err != nil {
					t.Fatalf("Failed to query for doc: %s", err)
				}
				if attCount != 1 {
					t.Errorf("Expected 1 deleted document, got %d", attCount)
				}
			},
		}
	})
	/*
		TODO:
		- different UUID configuration options????
		- retry in case of duplicate random uuid ???
		- support batch mode?
	*/

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		db := tt.db
		if db == nil {
			db = newDB(t)
		}
		opts := tt.opts
		if opts == nil {
			opts = mock.NilOption
		}
		docID, rev, err := db.CreateDoc(context.Background(), tt.doc, opts)
		if d := errors.StatusErrorDiff(tt.wantErr, tt.wantStatus, err); d != "" {
			t.Errorf("Unexpected error: %s", d)
		}
		if err != nil {
			return
		}
		if !regexp.MustCompile(tt.wantDocID).MatchString(docID) {
			t.Errorf("Unexpected doc ID. Expected %s, got %s", tt.wantDocID, docID)
		}
		if !regexp.MustCompile(tt.wantRev).MatchString(rev) {
			t.Errorf("Unexpected rev. Expected %s, got %s", tt.wantRev, rev)
		}
		if tt.check != nil {
			tt.check(t)
		}
	})
}
