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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

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
		err = d.db.QueryRowContext(ctx, fmt.Sprintf(`
			INSERT INTO %q (id, rev, rev_id, doc)
			VALUES ($1, $2, $3, $4)
			RETURNING rev || '-' || rev_id
		`, d.name), docID, rev.rev, rev.id, jsonDoc).Scan(&newRev)
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
	err = tx.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT COALESCE(MAX(rev || '-' || rev_id),'')
		FROM %q
		WHERE id = $1
	`, d.name), docID).Scan(&curRev)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	if curRev != docRev {
		return "", &internal.Error{Status: http.StatusConflict, Message: "conflict"}
	}
	var newRev string
	err = tx.QueryRowContext(ctx, fmt.Sprintf(`
		INSERT INTO %[1]q (id, rev, rev_id, doc)
		SELECT $1, COALESCE(MAX(rev),0) + 1, $2, $3
		FROM %[1]q
		WHERE id = $1
		RETURNING rev || '-' || rev_id
	`, d.name), docID, rev, jsonDoc).Scan(&newRev)
	if err != nil {
		return "", err
	}
	return newRev, tx.Commit()
}

func (d *db) Get(ctx context.Context, id string, options driver.Options) (*driver.Document, error) {
	opts := map[string]interface{}{}
	options.Apply(opts)

	var rev, body string
	var err error

	if optsRev, _ := opts["rev"].(string); optsRev != "" {
		var r revision
		r, err = parseRev(optsRev)
		if err != nil {
			return nil, err
		}
		err = d.db.QueryRowContext(ctx, fmt.Sprintf(`
			SELECT doc
			FROM %q
			WHERE id = $1
				AND rev = $2
				AND rev_id = $3
			`, d.name), id, r.rev, r.id).Scan(&body)
		rev = optsRev
	} else {
		err = d.db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT rev || '-' || rev_id, doc
		FROM %q
		WHERE id = $1
			AND deleted = FALSE
		ORDER BY rev DESC, rev_id DESC
		LIMIT 1
	`, d.name), id).Scan(&rev, &body)
	}

	if errors.Is(err, sql.ErrNoRows) {
		return nil, &internal.Error{Status: http.StatusNotFound, Message: "not found"}
	}
	if err != nil {
		return nil, err
	}

	if conflicts, _ := opts["conflicts"].(bool); conflicts {
		var revs []string
		rows, err := d.db.QueryContext(ctx, fmt.Sprintf(`
			SELECT rev || '-' || rev_id
			FROM %q
			WHERE id = $1
				AND rev || '-' || rev_id != $2
				AND DELETED = FALSE
			`, d.name), id, rev)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var r string
			if err := rows.Scan(&r); err != nil {
				return nil, err
			}
			revs = append(revs, r)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		var doc map[string]interface{}
		if err := json.Unmarshal([]byte(body), &doc); err != nil {
			return nil, err
		}
		doc["_conflicts"] = revs
		jonDoc, err := json.Marshal(doc)
		if err != nil {
			return nil, err
		}
		body = string(jonDoc)
	}
	return &driver.Document{
		Rev:  rev,
		Body: io.NopCloser(strings.NewReader(body)),
	}, nil
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
