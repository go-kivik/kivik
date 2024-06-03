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
	"io"
	"net/http"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
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

	var curRev revision
	rev := opts.rev()
	if rev != "" {
		curRev, err = parseRev(rev)
		if err != nil {
			return "", err
		}
	}

	data.MD5sum, err = d.isLeafRev(ctx, tx, docID, curRev.rev, curRev.id)
	switch {
	case kivik.HTTPStatus(err) == http.StatusNotFound:
		if rev != "" {
			return "", &internal.Error{Status: http.StatusConflict, Message: "document update conflict"}
		}
		// This means the doc is being created, and is empty other than the attachment
		data.Doc = []byte("{}")
	case err != nil:
		return "", err
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
