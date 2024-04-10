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
	"fmt"
	"net/http"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

func (d *db) Purge(ctx context.Context, request map[string][]string) (*driver.PurgeResult, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	stmts := newStmtCache()
	result := &driver.PurgeResult{}

	for docID, revs := range request {
		for _, rev := range revs {
			r, err := parseRev(rev)
			if err != nil {
				return nil, err
			}
			_, err = d.isLeafRev(ctx, tx, docID, r.rev, r.id)
			if err != nil {
				switch kivik.HTTPStatus(err) {
				case http.StatusNotFound, http.StatusConflict:
					// Non-leaf rev, do nothing
					continue
				default:
					return nil, err
				}
			}

			stmt, err := stmts.prepare(ctx, tx, d.query(`
				DELETE FROM {{ .Revs }}
				WHERE id = $1 AND rev = $2 AND rev_id = $3
			`))
			if err != nil {
				return nil, err
			}
			_, err = stmt.ExecContext(ctx, docID, r.rev, r.id)
			if err != nil {
				return nil, fmt.Errorf("exec failed: %w", err)
			}
			if result.Purged == nil {
				result.Purged = map[string][]string{}
			}
			result.Purged[docID] = append(result.Purged[docID], rev)
		}
	}

	return result, tx.Commit()
}
