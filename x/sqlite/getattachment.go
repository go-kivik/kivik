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
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/x/options"
)

func (d *db) GetAttachment(ctx context.Context, docID string, filename string, opts driver.Options) (*driver.Attachment, error) {
	o := options.New(opts)

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var requestedRev revision
	if rev := o.Rev(); rev != "" {
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

	attachment, err := d.getAttachment(ctx, tx, docID, filename, requestedRev)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, &internal.Error{Status: http.StatusNotFound, Message: "missing"}
	}
	if err != nil {
		return nil, err
	}

	return attachment, tx.Commit()
}

func (d *db) getAttachment(
	ctx context.Context,
	tx *sql.Tx,
	docID, filename string,
	rev revision,
) (*driver.Attachment, error) {
	var att driver.Attachment
	var data []byte
	err := tx.QueryRowContext(ctx, d.query(`
		SELECT
			att.filename,
			att.content_type,
			att.digest,
			att.length,
			att.rev_pos,
			att.data
		FROM {{ .Attachments }} AS att
		JOIN {{ .AttachmentsBridge }} AS bridge ON bridge.pk = att.pk
		WHERE
			bridge.id = $1
			AND att.filename = $2
			AND bridge.rev = $3
			AND bridge.rev_id = $4	
	`), docID, filename, rev.rev, rev.id).
		Scan(&att.Filename, &att.ContentType, &att.Digest, &att.Size, &att.RevPos, &data)

	att.Content = io.NopCloser(bytes.NewReader(data))
	return &att, err
}
