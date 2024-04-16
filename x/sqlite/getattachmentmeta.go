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

func (d *db) GetAttachmentMeta(ctx context.Context, docID, filename string, options driver.Options) (*driver.Attachment, error) {
	opts := newOpts(options)

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var requestedRev revision
	if rev := opts.rev(); rev != "" {
		requestedRev, err = parseRev(rev)
		if err != nil {
			return nil, &internal.Error{Message: err.Error(), Status: http.StatusBadRequest}
		}
	} else {
		requestedRev, err = d.winningRev(ctx, tx, docID)
		if err != nil {
			return nil, err
		}
	}

	attachment, err := d.getAttachmentMetadata(ctx, tx, docID, filename, requestedRev)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, &internal.Error{Status: http.StatusNotFound, Message: "missing"}
	}
	if err != nil {
		return nil, err
	}

	return attachment, tx.Commit()
}

func (d *db) getAttachmentMetadata(
	ctx context.Context,
	tx *sql.Tx,
	docID, filename string,
	rev revision,
) (*driver.Attachment, error) {
	var att driver.Attachment
	var hash md5sum
	err := tx.QueryRowContext(ctx, d.query(`
		SELECT
			att.filename,
			att.content_type,
			att.digest,
			att.length
		FROM {{ .Attachments }} AS att
		JOIN {{ .AttachmentsBridge }} AS bridge ON bridge.pk = att.pk
		WHERE
			bridge.id = $1
			AND att.filename = $2
			AND bridge.rev = $3
			AND bridge.rev_id = $4	
	`), docID, filename, rev.rev, rev.id).
		Scan(&att.Filename, &att.ContentType, &hash, &att.Size)
	att.Digest = hash.Digest()
	return &att, err
}
