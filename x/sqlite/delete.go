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

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal"
)

func (d *db) Delete(ctx context.Context, docID string, options driver.Options) (string, error) {
	opts := map[string]interface{}{}
	options.Apply(opts)
	optRev, ok := opts["rev"].(string)
	if !ok {
		// Special case: No rev for DELETE is always a conflict, since you can't
		// delete a doc without a rev.
		return "", &internal.Error{Status: http.StatusConflict, Message: "conflict"}
	}
	delRev, err := parseRev(optRev)
	if err != nil {
		return "", err
	}

	data, err := prepareDoc(docID, map[string]interface{}{"_deleted": true})
	if err != nil {
		return "", err
	}

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	found, err := d.docRevExists(ctx, tx, docID, delRev)
	if err != nil {
		return "", err
	}
	if !found {
		return "", &internal.Error{Status: http.StatusConflict, Message: "conflict"}
	}
	var r revision
	err = tx.QueryRowContext(ctx, d.query(`
		INSERT INTO {{ .Revs }} (id, rev, rev_id, parent_rev, parent_rev_id)
		SELECT $1, COALESCE(MAX(rev),0) + 1, $2, $3, $4
		FROM {{ .Revs }}
		WHERE id = $1
		RETURNING rev, rev_id
	`), data.ID, data.RevID, delRev.rev, delRev.id).Scan(&r.rev, &r.id)
	if err != nil {
		return "", err
	}
	_, err = tx.ExecContext(ctx, d.query(`
		INSERT INTO {{ .Docs }} (id, rev, rev_id, doc, deleted)
		VALUES ($1, $2, $3, $4, TRUE)
	`), data.ID, r.rev, r.id, data.Doc)
	if err != nil {
		return "", err
	}

	return r.String(), tx.Commit()
}
