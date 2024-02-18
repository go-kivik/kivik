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
	newEdits, _ := opts["new_edits"].(bool)
	data, err := prepareDoc(docID, doc)
	if err != nil {
		return "", err
	}

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	if data.Revisions.Start != 0 {
		if newEdits {
			stmt, err := tx.PrepareContext(ctx, fmt.Sprintf(`
				SELECT EXISTS(
					SELECT 1
					FROM %[1]q
					WHERE id = $1
						AND rev = $2
						AND rev_id = $3
				)
			`, d.name+"_revs"))
			if err != nil {
				return "", err
			}
			defer stmt.Close()
			var exists bool
			revs := data.Revisions.revs()
			for _, r := range revs[:len(revs)-1] {
				err := stmt.QueryRowContext(ctx, data.ID, r.rev, r.id).Scan(&exists)
				if err != nil {
					return "", err
				}
				if !exists {
					return "", &internal.Error{Status: http.StatusConflict, Message: "conflict"}
				}
			}
		}
		docRev = data.Revisions.leaf().String()
	}
	if optsRev != "" && docRev != "" && optsRev != docRev {
		return "", &internal.Error{Status: http.StatusBadRequest, Message: "Document rev and option have different values"}
	}
	if docRev == "" && optsRev != "" {
		docRev = optsRev
	}

	if !newEdits {
		var rev revision
		if data.Revisions.Start != 0 {
			stmt, err := tx.PrepareContext(ctx, fmt.Sprintf(`
				INSERT INTO %[1]q (id, rev, rev_id, parent_rev, parent_rev_id)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT DO UPDATE SET parent_rev = $4, parent_rev_id = $5
			`, d.name+"_revs"))
			if err != nil {
				return "", err
			}
			defer stmt.Close()

			var (
				parentRev   *int
				parentRevID *string
			)
			for _, r := range data.Revisions.revs() {
				r := r
				_, err := stmt.ExecContext(ctx, data.ID, r.rev, r.id, parentRev, parentRevID)
				if err != nil {
					return "", err
				}
				parentRev = &r.rev
				parentRevID = &r.id
			}
			rev = data.Revisions.leaf()
		} else {
			if docRev == "" {
				return "", &internal.Error{Status: http.StatusBadRequest, Message: "When `new_edits: false`, the document needs `_rev` or `_revisions` specified"}
			}
			rev, err = parseRev(docRev)
			if err != nil {
				return "", err
			}
			_, err = tx.ExecContext(ctx, fmt.Sprintf(`
			INSERT INTO %[1]q (id, rev, rev_id)
			VALUES ($1, $2, $3)
			ON CONFLICT DO NOTHING
		`, d.name+"_revs"), docID, rev.rev, rev.id)
			if err != nil {
				return "", err
			}
		}
		var newRev string
		err = tx.QueryRowContext(ctx, fmt.Sprintf(`
			INSERT INTO %q (id, rev, rev_id, doc, deleted)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT DO NOTHING
			RETURNING rev || '-' || rev_id
		`, d.name), docID, rev.rev, rev.id, data.Doc, data.Deleted).Scan(&newRev)
		if errors.Is(err, sql.ErrNoRows) {
			// No rows means a conflict, so  we assume that the documents are
			// identical, for the sake of idempotency, and return the current
			// rev, to match CouchDB behavior.
			return docRev, nil
		}
		if err != nil {
			return "", err
		}
		return newRev, tx.Commit()
	}

	var curRev revision
	err = tx.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT rev, rev_id
		FROM %q
		WHERE id = $1
		ORDER BY rev DESC, rev_id DESC
		LIMIT 1
	`, d.name), docID).Scan(&curRev.rev, &curRev.id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
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
		VALUES ($1, $2, $3, $4, $5)
	`, d.name), data.ID, r.rev, r.id, data.Doc, data.Deleted)
	if err != nil {
		return "", err
	}

	if len(data.Attachments) == 0 {
		return r.String(), tx.Commit()
	}

	stmt, err := tx.PrepareContext(ctx, fmt.Sprintf(`
		INSERT INTO %[1]q (id, rev, rev_id, filename, content_type, length, digest, data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, d.name+"_attachments"))
	if err != nil {
		return "", err
	}
	defer stmt.Close()
	for filename, att := range data.Attachments {
		if err := att.calculate(filename); err != nil {
			return "", err
		}
		contentType := att.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		_, err := stmt.ExecContext(ctx, data.ID, r.rev, r.id, filename, contentType, att.Length, att.Digest, att.Content)
		if err != nil {
			return "", err
		}
	}

	return r.String(), tx.Commit()
}
