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

// runValidation runs all validate_doc_update functions against the proposed
// change. The old document is only fetched if validate functions exist.
func (d *db) runValidation(ctx context.Context, tx *sql.Tx, docID string, newDoc map[string]any, curRev revision) error {
	funcs, err := d.getValidateFuncs(ctx, tx)
	if err != nil || len(funcs) == 0 {
		return err
	}

	var oldDoc any
	if !curRev.IsZero() {
		doc, _, err := d.getCoreDoc(ctx, tx, docID, curRev, false, false)
		if err != nil {
			return err
		}
		oldDoc = doc.toMap()
	}

	userCtx := map[string]any{}
	secObj := map[string]any{}
	for _, fn := range funcs {
		if err := fn(newDoc, oldDoc, userCtx, secObj); err != nil {
			return err
		}
	}
	return nil
}

// getValidateFuncs returns all compiled validate_doc_update functions from
// design document leaf revisions.
func (d *db) getValidateFuncs(ctx context.Context, tx *sql.Tx) ([]js.ValidateFunc, error) {
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
		return nil, err
	}
	defer rows.Close()

	var funcs []js.ValidateFunc
	for rows.Next() {
		var funcBody string
		if err := rows.Scan(&funcBody); err != nil {
			return nil, err
		}
		fn, err := js.Validate(funcBody)
		if err != nil {
			return nil, err
		}
		funcs = append(funcs, fn)
	}
	return funcs, rows.Err()
}
