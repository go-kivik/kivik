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
	"database/sql"
	"testing"
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
		FROM "test_revs"
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
