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
	"github.com/go-kivik/kivik/v4/internal"
)

func (d *db) DeleteAttachment(ctx context.Context, docID, filename string, options driver.Options) (string, error) {
	opts := newOpts(options)
	if opts.rev() == "" {
		return "", &internal.Error{Status: http.StatusConflict, Message: "conflict"}
	}
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	data := &docData{
		ID: docID,
	}

	curRev, hash, err := d.winningRev(ctx, tx, docID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", &internal.Error{Status: http.StatusNotFound, Message: "document not found"}
	case err != nil:
		return "", err
	default:
		data.MD5sum = hash
	}

	if rev := opts.rev(); rev != "" && rev != curRev.String() {
		return "", &internal.Error{Status: http.StatusConflict, Message: "conflict"}
	}

	// Read list of current attachments, then remove the requested one

	rows, err := tx.QueryContext(ctx, d.query(`
		SELECT att.filename
		FROM {{ .Attachments }} AS att
		JOIN {{ .AttachmentsBridge }} AS bridge ON att.pk = bridge.pk
		WHERE bridge.id = $1
			AND bridge.rev = $2
			AND bridge.rev_id = $3
	`), docID, curRev.rev, curRev.id)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var attachments []string
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return "", err
		}
		attachments = append(attachments, filename)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	if !attachmentsContains(attachments, filename) {
		return "", &internal.Error{Status: http.StatusNotFound, Message: "attachment not found"}
	}

	r, err := d.createRev(ctx, tx, data, curRev)
	if err != nil {
		return "", err
	}

	return r.String(), tx.Commit()
}

func attachmentsContains(attachments []string, filename string) bool {
	for _, att := range attachments {
		if att == filename {
			return true
		}
	}
	return false
}
