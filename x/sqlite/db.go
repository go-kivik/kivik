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

func (d *db) Put(ctx context.Context, docID string, doc interface{}, _ driver.Options) (string, error) {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var latest string
	err = tx.QueryRowContext(ctx, fmt.Sprintf("SELECT MAX(rev) FROM %q WHERE id = ?", d.name), docID).Scan(&latest)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	const query = `
	INSERT INTO %[1]q (id, rev, doc)
	VALUES (?, (COALESCE(?, (SELECT COALESCE(MAX(rev),1) FROM %[1]q WHERE id = ?), '1')||'-'||?), ?)
	RETURNING rev
	`
	jsonDoc, id, revID, rev, err := marshalDoc(doc)
	fmt.Println(id, revID, rev, err)
	if err != nil {
		return "", err
	}
	if id != "" && id != docID {
		return "", &internal.Error{Status: http.StatusBadRequest, Message: "Document ID must match docID in URL"}
	}
	var revIDPtr *string
	if revID != "" {
		revIDPtr = &revID
	}
	var newRev string
	err = d.db.QueryRowContext(ctx, fmt.Sprintf(query, d.name), id, revIDPtr, id, rev, jsonDoc).Scan(&newRev)
	return newRev, err
}

func (db) Get(context.Context, string, driver.Options) (*driver.Document, error) {
	return nil, nil
}

func (db) Delete(context.Context, string, driver.Options) (string, error) {
	return "", nil
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
