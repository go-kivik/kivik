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

	"github.com/google/uuid"

	"github.com/go-kivik/kivik/v4/driver"
)

func (d *db) CreateDoc(ctx context.Context, doc interface{}, _ driver.Options) (string, string, error) {
	data, err := prepareDoc("", doc)
	if err != nil {
		return "", "", err
	}
	if data.ID == "" {
		data.ID = uuid.NewString()
	}

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return "", "", err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, d.query(`
		INSERT INTO {{ .Revs }} (id, rev, rev_id)
		VALUES ($1, 1, $2)
	`), data.ID, data.RevID())
	if err != nil {
		return "", "", err
	}

	_, err = tx.ExecContext(ctx, d.query(`
		INSERT INTO {{ .Docs }} (id, rev, rev_id, doc, md5sum, deleted)
		VALUES ($1, 1, $2, $3, $4, FALSE)
	`), data.ID, data.RevID(), data.Doc, data.MD5sum)
	if err != nil {
		return "", "", err
	}

	return data.ID, "1-" + data.RevID(), tx.Commit()
}
