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
	"database/sql"

	"github.com/go-kivik/kivik/x/sqlite/v4/js"
)

// validateDoc runs all validate_doc_update functions stored in design documents
// against the proposed document change.
func (d *db) validateDoc(ctx context.Context, tx *sql.Tx, newDoc, oldDoc, userCtx, secObj any) error {
	rows, err := tx.QueryContext(ctx, d.query(`
		SELECT design.func_body
		FROM {{ .Design }} AS design
		JOIN (
			SELECT
				rev.id,
				rev.rev,
				rev.rev_id
			FROM {{ .Revs }} AS rev
			LEFT JOIN {{ .Revs }} AS child
				ON child.id = rev.id
				AND rev.rev = child.parent_rev
				AND rev.rev_id = child.parent_rev_id
			JOIN {{ .Docs }} AS doc
				ON rev.id = doc.id
				AND rev.rev = doc.rev
				AND rev.rev_id = doc.rev_id
			WHERE child.id IS NULL
				AND NOT doc.deleted
		) AS leaves ON design.id = leaves.id AND design.rev = leaves.rev AND design.rev_id = leaves.rev_id
		WHERE design.func_type = 'validate'
	`))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var funcBody string
		if err := rows.Scan(&funcBody); err != nil {
			return err
		}
		fn, err := js.Validate(funcBody)
		if err != nil {
			return err
		}
		if err := fn(newDoc, oldDoc, userCtx, secObj); err != nil {
			return err
		}
	}
	return rows.Err()
}
