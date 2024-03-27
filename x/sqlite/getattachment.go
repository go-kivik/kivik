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
	"github.com/go-kivik/kivik/v4/internal"
)

func (d *db) GetAttachment(ctx context.Context, docID string, filename string, _ driver.Options) (*driver.Attachment, error) {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	curRev, err := d.currentRev(ctx, tx, docID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, &internal.Error{Status: http.StatusNotFound, Message: "Not Found: missing"}
	}

	attachment, err := d.attachmentExists(ctx, tx, docID, filename, curRev)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, &internal.Error{Status: http.StatusNotFound, Message: "Not Found: missing"}
	}
	if err != nil {
		return nil, err
	}

	return attachment, tx.Commit()
}

func (d *db) attachmentExists(
	ctx context.Context,
	tx *sql.Tx,
	docID, filename string,
	rev revision,
) (*driver.Attachment, error) {
	var att driver.Attachment
	var data []byte
	err := tx.QueryRowContext(ctx, d.query(`
		SELECT filename, content_type, length, rev, data
		FROM {{ .Attachments }}
		WHERE id = $1
			AND filename = $2
			AND rev = $3
			AND rev_id = $4
		`), docID, filename, rev.rev, rev.id).
		Scan(&att.Filename, &att.ContentType, &att.Size, &att.RevPos, &data)
	att.Content = io.NopCloser(bytes.NewReader(data))
	return &att, err
}