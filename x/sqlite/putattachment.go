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
	"io"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal"
)

func (d *db) PutAttachment(ctx context.Context, docID string, att *driver.Attachment, options driver.Options) (string, error) {
	opts := newOpts(options)

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
		data.Doc = []byte("{}")
	case err != nil:
		return "", err
	default:
		data.MD5sum = hash
	}

	if rev := opts.rev(); rev != "" && rev != curRev.String() {
		return "", &internal.Error{Status: http.StatusConflict, Message: "conflict"}
	}

	content, err := io.ReadAll(att.Content)
	if err != nil {
		return "", err
	}
	file := attachment{
		ContentType: att.ContentType,
		Content:     content,
	}
	if err := file.calculate(att.Filename); err != nil {
		return "", err
	}
	data.Attachments = map[string]attachment{
		att.Filename: file,
	}

	r, err := d.createRev(ctx, tx, data, curRev)
	if err != nil {
		return "", err
	}

	return r.String(), tx.Commit()
}
