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

	curRev, err := d.currentRev(ctx, tx, docID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", &internal.Error{Status: http.StatusNotFound, Message: "document not found"}
	case err != nil:
		return "", err
	default:
		data.RevID = curRev.id
	}

	if rev := opts.rev(); rev != "" && rev != curRev.String() {
		return "", &internal.Error{Status: http.StatusConflict, Message: "conflict"}
	}

	data.RemovedAttachments = []string{filename}

	r, err := d.createRev(ctx, tx, data, curRev)
	if err != nil {
		return "", err
	}

	return r.String(), tx.Commit()
}
