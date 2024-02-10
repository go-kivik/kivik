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

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"

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

func (d *db) Put(ctx context.Context, docID string, doc interface{}, options driver.Options) (string, error) {
	docRev, err := extractRev(doc)
	if err != nil {
		return "", err
	}
	opts := map[string]interface{}{
		"new_edits": true,
	}
	options.Apply(opts)
	optsRev, _ := opts["rev"].(string)
	if optsRev != "" && docRev != "" && optsRev != docRev {
		return "", &internal.Error{Status: http.StatusBadRequest, Message: "Document rev and option have different values"}
	}
	if docRev == "" && optsRev != "" {
		docRev = optsRev
	}

	docID, rev, jsonDoc, err := prepareDoc(docID, doc)
	if err != nil {
		return "", err
	}

	if newEdits, _ := opts["new_edits"].(bool); !newEdits {
		if docRev == "" {
			return "", &internal.Error{Status: http.StatusBadRequest, Message: "When `new_edits: false`, the document needs `_rev` or `_revisions` specified"}
		}
		rev, err := parseRev(docRev)
		if err != nil {
			return "", err
		}
		var newRev string
		err = d.db.QueryRowContext(ctx, `
			INSERT INTO `+d.name+` (id, rev_id, rev, doc)
			VALUES ($1, $2, $3, $4)
			RETURNING rev_id || '-' || rev
		`, docID, rev.id, rev.rev, jsonDoc).Scan(&newRev)
		var sqliteErr *sqlite.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
			// In the case of a conflict for new_edits=false, we assume that the
			// documents are identical, for the sake of idempotency, and return
			// the current rev, to match CouchDB behavior.
			return docRev, nil
		}
		return newRev, err
	}

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var curRev string
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(rev_id || '-' || rev),'')
		FROM `+d.name+`
		WHERE id = $1
	`, docID).Scan(&curRev)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	if curRev != docRev {
		return "", &internal.Error{Status: http.StatusConflict, Message: "conflict"}
	}
	var newRev string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO `+d.name+` (id, rev_id, rev, doc)
		SELECT $1, COALESCE(MAX(rev_id),0) + 1, $2, $3
		FROM `+d.name+`
		WHERE id = $1
		RETURNING rev_id || '-' || rev
	`, docID, rev, jsonDoc).Scan(&newRev)
	if err != nil {
		return "", err
	}
	return newRev, tx.Commit()
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
