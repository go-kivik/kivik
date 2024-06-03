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

	"github.com/google/uuid"

	"github.com/go-kivik/kivik/v4/driver"
	kerrors "github.com/go-kivik/kivik/v4/int/errors"
)

func (d *db) CreateDoc(ctx context.Context, doc interface{}, _ driver.Options) (string, string, error) {
	data, err := prepareDoc("", doc)
	if err != nil {
		return "", "", err
	}
	if data.ID == "" {
		data.ID = uuid.NewString()
	}

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return "", "", err
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRowContext(ctx, d.query(`
		SELECT EXISTS (
			SELECT 1
			FROM {{ .Revs }} AS rev
			JOIN {{ .Docs }} AS doc ON doc.id = rev.id AND doc.rev = rev.rev AND doc.rev_id = rev.rev_id
			LEFT JOIN {{ .Revs }} AS child ON child.parent_rev = rev.rev AND child.parent_rev_id = rev.rev_id
			WHERE rev.id = $1
				AND child.rev IS NULL
				AND doc.deleted = FALSE
		)
	`), data.ID).Scan(&exists)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", "", err
	}
	if exists {
		return "", "", &kerrors.Error{Status: http.StatusConflict, Message: "document update conflict"}
	}

	_, err = tx.ExecContext(ctx, d.query(`
		INSERT INTO {{ .Revs }} (id, rev, rev_id)
		VALUES ($1, 1, $2)
	`), data.ID, data.RevID())
	if err != nil {
		return "", "", err
	}

	_, err = tx.ExecContext(ctx, d.query(`
		INSERT INTO {{ .Docs }} (id, rev, rev_id, doc, md5sum, deleted)
		VALUES ($1, 1, $2, $3, $4, $5)
	`), data.ID, data.RevID(), data.Doc, data.MD5sum, data.Deleted)
	if err != nil {
		return "", "", err
	}

	if err := d.createDocAttachments(ctx, data, tx, revision{rev: 1, id: data.RevID()}, nil); err != nil {
		return "", "", err
	}

	return data.ID, "1-" + data.RevID(), tx.Commit()
}
