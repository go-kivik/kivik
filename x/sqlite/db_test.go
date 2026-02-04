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
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/mock"
)

type leaf struct {
	ID          string
	Rev         int
	RevID       string
	ParentRev   *int
	ParentRevID *string
}

func readRevisions(t *testing.T, db *sql.DB) []leaf {
	t.Helper()
	rows, err := db.Query(`
		SELECT id, rev, rev_id, parent_rev, parent_rev_id
		FROM "kivik$test$revs"
		ORDER BY id, rev, rev_id
	`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var leaves []leaf
	for rows.Next() {
		var l leaf
		if err := rows.Scan(&l.ID, &l.Rev, &l.RevID, &l.ParentRev, &l.ParentRevID); err != nil {
			t.Fatal(err)
		}
		leaves = append(leaves, l)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return leaves
}

func Test_db_not_found(t *testing.T) {
	t.Parallel()
	const (
		wantStatus = http.StatusNotFound
		wantErr    = "database not found: db_not_found"
	)

	type test struct {
		call func(*db) error
	}

	tests := testy.NewTable()
	tests.Add("Changes", test{
		call: func(d *db) error {
			_, err := d.Changes(context.Background(), mock.NilOption)
			return err
		},
	})
	tests.Add("Changes, since=now", test{
		call: func(d *db) error {
			_, err := d.Changes(context.Background(), kivik.Param("since", "now"))
			return err
		},
	})
	tests.Add("Changes, longpoll", test{
		call: func(d *db) error {
			_, err := d.Changes(context.Background(), kivik.Param("feed", "longpoll"))
			return err
		},
	})
	tests.Add("CreateDoc", test{
		call: func(d *db) error {
			_, _, err := d.CreateDoc(context.Background(), map[string]string{}, mock.NilOption)
			return err
		},
	})
	tests.Add("DeleteAttachment", test{
		call: func(d *db) error {
			_, err := d.DeleteAttachment(context.Background(), "doc", "att", kivik.Rev("1-x"))
			return err
		},
	})
	tests.Add("Delete", test{
		call: func(d *db) error {
			_, err := d.Delete(context.Background(), "doc", kivik.Rev("1-x"))
			return err
		},
	})
	tests.Add("Find", test{
		call: func(d *db) error {
			_, err := d.Find(context.Background(), json.RawMessage(`{"selector":{}}`), mock.NilOption)
			return err
		},
	})
	tests.Add("GetAttachment", test{
		call: func(d *db) error {
			_, err := d.GetAttachment(context.Background(), "doc", "att", mock.NilOption)
			return err
		},
	})
	tests.Add("GetAttachmentMeta", test{
		call: func(d *db) error {
			_, err := d.GetAttachmentMeta(context.Background(), "doc", "att", mock.NilOption)
			return err
		},
	})
	tests.Add("Get", test{
		call: func(d *db) error {
			_, err := d.Get(context.Background(), "doc", mock.NilOption)
			return err
		},
	})
	tests.Add("GetRev", test{
		call: func(d *db) error {
			_, err := d.GetRev(context.Background(), "doc", mock.NilOption)
			return err
		},
	})
	tests.Add("OpenRevs", test{
		call: func(d *db) error {
			_, err := d.OpenRevs(context.Background(), "doc", nil, mock.NilOption)
			return err
		},
	})
	tests.Add("Purge", test{
		call: func(d *db) error {
			_, err := d.Purge(context.Background(), map[string][]string{"doc": {"1-x"}})
			return err
		},
	})
	tests.Add("PutAttachment", test{
		call: func(d *db) error {
			_, err := d.PutAttachment(context.Background(), "doc", &driver.Attachment{}, mock.NilOption)
			return err
		},
	})
	tests.Add("Put", test{
		call: func(d *db) error {
			_, err := d.Put(context.Background(), "doc", map[string]string{}, mock.NilOption)
			return err
		},
	})
	tests.Add("Put, new_edits=false", test{
		call: func(d *db) error {
			_, err := d.Put(context.Background(), "doc", map[string]any{
				"_rev": "1-x",
			}, kivik.Param("new_edits", false))
			return err
		},
	})
	tests.Add("Put, new_edits=false + _revisions", test{
		call: func(d *db) error {
			_, err := d.Put(context.Background(), "doc", map[string]any{
				"_revisions": map[string]interface{}{
					"start": 1,
					"ids":   []string{"x"},
				},
			}, kivik.Param("new_edits", false))
			return err
		},
	})
	tests.Add("Put, _revisoins + new_edits=true", test{
		call: func(d *db) error {
			_, err := d.Put(context.Background(), "doc", map[string]any{
				"_revisions": map[string]interface{}{
					"start": 1,
					"ids":   []string{"x", "y", "z"},
				},
			}, mock.NilOption)
			return err
		},
	})
	tests.Add("AllDocs", test{
		call: func(d *db) error {
			_, err := d.AllDocs(context.Background(), mock.NilOption)
			return err
		},
	})
	tests.Add("Query", test{
		call: func(d *db) error {
			_, err := d.Query(context.Background(), "ddoc", "view", mock.NilOption)
			return err
		},
	})
	tests.Add("Query, group=true", test{
		call: func(d *db) error {
			_, err := d.Query(context.Background(), "ddoc", "view", kivik.Param("group", true))
			return err
		},
	})
	tests.Add("RevsDiff", test{
		call: func(d *db) error {
			_, err := d.RevsDiff(context.Background(), map[string][]string{"doc": {"1-x"}})
			return err
		},
	})
	/*
		TODO:
	*/

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		client, err := (drv{}).NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}
		d, err := client.DB("db_not_found", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}

		err = tt.call(d.(*db))
		if err == nil {
			t.Fatal("Expected error")
		}
		if wantErr != err.Error() {
			t.Errorf("Unexpected error: %s", err)
		}
		if status := kivik.HTTPStatus(err); status != wantStatus {
			t.Errorf("Unexpected status: %d", status)
		}
	})
}

func TestDBCompact(t *testing.T) {
	t.Parallel()
	d := newDB(t)
	_ = d.tPut("doc1", map[string]string{"foo": "bar"})

	err := d.Compact(context.Background())
	if err != nil {
		t.Fatalf("Compact failed: %s", err)
	}
}
