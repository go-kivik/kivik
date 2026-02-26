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
	"errors"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/x/sqlite/v4/js"
)

// Update calls the named update function with the provided document.
func (d *db) Update(ctx context.Context, ddoc, funcName, docID string, doc any, opts driver.Options) (string, error) {
	var funcBody string
	err := d.db.QueryRowContext(ctx, d.query(`
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
		WHERE design.func_type = 'update'
			AND design.id = $1
			AND design.func_name = $2
	`), ddoc, funcName).Scan(&funcBody)
	if errors.Is(err, sql.ErrNoRows) {
		return "", &internal.Error{Status: http.StatusNotFound, Message: "missing update function " + funcName + " on " + ddoc}
	}
	if err != nil {
		return "", err
	}

	updateFunc, err := js.Update(funcBody)
	if err != nil {
		return "", err
	}

	existingDoc, _, err := d.getCoreDoc(ctx, d.db, docID, revision{}, false, false)
	if err != nil {
		return "", err
	}

	updatedDoc, _, err := updateFunc(existingDoc.toMap(), map[string]any{})
	if err != nil {
		return "", err
	}

	if updatedDoc == nil {
		return "", nil
	}

	return d.Put(ctx, docID, updatedDoc, opts)
}
