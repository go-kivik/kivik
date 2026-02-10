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
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/x/options"
)

func (d *db) Delete(ctx context.Context, docID string, opts driver.Options) (string, error) {
	o := options.New(opts)
	optRev := o.Rev()
	if optRev == "" {
		// Special case: No rev for DELETE is always a conflict, since you can't
		// delete a doc without a rev.
		return "", &internal.Error{Status: http.StatusConflict, Message: "document update conflict"}
	}

	data, err := prepareDoc(docID, map[string]any{"_deleted": true})
	if err != nil {
		return "", err
	}

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	curRev, err := parseRev(o.Rev())
	if err != nil {
		return "", err
	}

	data.MD5sum, err = d.isLeafRev(ctx, tx, docID, curRev.rev, curRev.id)
	if err != nil {
		return "", d.errDatabaseNotFound(err)
	}
	data.Deleted = true
	data.Doc = []byte("{}")

	newDocMap := map[string]any{
		"_id":      data.ID,
		"_deleted": true,
	}

	if err := d.runValidation(ctx, tx, docID, newDocMap, curRev); err != nil {
		return "", err
	}

	r, err := d.createRev(ctx, tx, data, curRev)
	if err != nil {
		return "", err
	}

	return r.String(), tx.Commit()
}
