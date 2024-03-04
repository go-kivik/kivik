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

func (d *db) GetAttachment(ctx context.Context, docID string, filename string, _ driver.Options) (*driver.Attachment, error) {
	attachment, err := d.attachmentExists(ctx, docID, filename)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, &internal.Error{Status: http.StatusNotFound, Message: "Not Found: missing"}
	}
	if err != nil {
		return nil, err
	}

	return attachment, nil
}

func (d *db) attachmentExists(ctx context.Context, docID string, filename string) (*driver.Attachment, error) {
	var att driver.Attachment
	err := d.db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT filename, content_type, length, rev
		FROM %s
		WHERE id = $1 AND filename = $2
		`, d.name+"_attachments"), docID, filename).Scan(&att.Filename, &att.ContentType, &att.Size, &att.RevPos)
	return &att, err
}
