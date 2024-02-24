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

	"github.com/go-kivik/kivik/v4/driver"
)

func (d *db) PutAttachment(ctx context.Context, docID string, att *driver.Attachment, _ driver.Options) (string, error) {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	data, err := prepareDoc(docID, map[string]string{})
	if err != nil {
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

	r, err := d.createRev(ctx, tx, data, revision{})
	if err != nil {
		return "", err
	}

	return r.String(), tx.Commit()
}
