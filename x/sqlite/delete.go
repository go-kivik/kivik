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
	"fmt"
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

	var found bool
	err = tx.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT child.id IS NULL
		FROM %[2]q AS rev
		LEFT JOIN %[2]q AS child ON rev.id = child.id AND rev.rev = child.parent_rev AND rev.rev_id = child.parent_rev_id
		JOIN %[1]q AS doc ON rev.id = doc.id AND rev.rev = doc.rev AND rev.rev_id = doc.rev_id
		WHERE rev.id = $1
			AND rev.rev = $2
			AND rev.rev_id = $3
	`, d.name, d.name+"_revs"), data.ID, delRev.rev, delRev.id).Scan(&found)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", &internal.Error{Status: http.StatusNotFound, Message: "not found"}
	case err != nil:
		return "", err
	}
	if !found {
		return "", &internal.Error{Status: http.StatusConflict, Message: "conflict"}
	}
	var r revision
	err = tx.QueryRowContext(ctx, fmt.Sprintf(`
		INSERT INTO %[1]q (id, rev, rev_id, parent_rev, parent_rev_id)
		SELECT $1, COALESCE(MAX(rev),0) + 1, $2, $3, $4
		FROM %[1]q
		WHERE id = $1
		RETURNING rev, rev_id
	`, d.name+"_revs"), data.ID, data.RevID, delRev.rev, delRev.id).Scan(&r.rev, &r.id)
	if err != nil {
		return "", err
	}
	_, err = tx.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %[1]q (id, rev, rev_id, doc, deleted)
		VALUES ($1, $2, $3, $4, TRUE)
	`, d.name), data.ID, r.rev, r.id, data.Doc)
	if err != nil {
		return "", err
	}
	return r.String(), tx.Commit()
}
