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

type db struct {
	db   *sql.DB
	name string
}

var _ driver.DB = (*db)(nil)

func (db) AllDocs(context.Context, driver.Options) (driver.Rows, error) {
	return nil, nil
}

func (db) CreateDoc(context.Context, interface{}, driver.Options) (string, string, error) {
	return "", "", nil
}

func (d *db) Delete(ctx context.Context, docID string, options driver.Options) (string, error) {
	opts := map[string]interface{}{}
	options.Apply(opts)
	docRev, _ := opts["rev"].(string)

	data, err := prepareDoc(docID, map[string]interface{}{"_deleted": true})
	if err != nil {
		return "", err
	}

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var curRev revision
	err = tx.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT rev, rev_id
		FROM %q
		WHERE id = $1
		ORDER BY rev DESC, rev_id DESC
		LIMIT 1
	`, d.name), data.ID).Scan(&curRev.rev, &curRev.id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", &internal.Error{Status: http.StatusNotFound, Message: "not found"}
	case err != nil:
		return "", err
	}
	if curRev.String() != docRev {
		return "", &internal.Error{Status: http.StatusConflict, Message: "conflict"}
	}
	var (
		r         revision
		curRevRev *int
		curRevID  *string
	)
	if curRev.rev != 0 {
		curRevRev = &curRev.rev
		curRevID = &curRev.id
	}
	err = tx.QueryRowContext(ctx, fmt.Sprintf(`
		INSERT INTO %[1]q (id, rev, rev_id, parent_rev, parent_rev_id)
		SELECT $1, COALESCE(MAX(rev),0) + 1, $2, $3, $4
		FROM %[1]q
		WHERE id = $1
		RETURNING rev, rev_id
	`, d.name+"_revs"), data.ID, data.RevID, curRevRev, curRevID).Scan(&r.rev, &r.id)
	if err != nil {
		return "", err
	}
	_, err = tx.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %[1]q (id, rev, rev_id, doc, deleted)
		VALUES ($1, $2, $3, $4, TRUE)
	`, d.name), data.ID, r.rev, r.id, data.Doc)
	if err != nil {
		return "", err
	}
	return r.String(), tx.Commit()
}

func (db) Stats(context.Context) (*driver.DBStats, error) {
	return nil, nil
}

func (db) Compact(context.Context) error {
	return nil
}

func (db) CompactView(context.Context, string) error {
	return nil
}

func (db) ViewCleanup(context.Context) error {
	return nil
}

func (db) Changes(context.Context, driver.Options) (driver.Changes, error) {
	return nil, nil
}

func (db) PutAttachment(context.Context, string, *driver.Attachment, driver.Options) (string, error) {
	return "", nil
}

func (db) GetAttachment(context.Context, string, string, driver.Options) (*driver.Attachment, error) {
	return nil, nil
}

func (db) DeleteAttachment(context.Context, string, string, driver.Options) (string, error) {
	return "", nil
}

func (db) Query(context.Context, string, string, driver.Options) (driver.Rows, error) {
	return nil, nil
}

func (db) Close() error {
	return nil
}
