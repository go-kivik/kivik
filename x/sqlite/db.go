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
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
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

	if data.Revisions.Start != 0 {
		if newEdits {
			stmt, err := d.db.PrepareContext(ctx, fmt.Sprintf(`
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

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

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
	return r.String(), tx.Commit()
}

func (d *db) Get(ctx context.Context, id string, options driver.Options) (*driver.Document, error) {
	opts := map[string]interface{}{}
	options.Apply(opts)

	var (
		r       revision
		body    []byte
		err     error
		deleted bool
	)

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	optsRev, _ := opts["rev"].(string)
	if optsRev != "" {
		r, err = parseRev(optsRev)
		if err != nil {
			return nil, err
		}
		if latest, _ := opts["latest"].(bool); latest {
			err = tx.QueryRowContext(ctx, fmt.Sprintf(`
				WITH RECURSIVE Descendants AS (
					-- Base case: Select the starting node for descendants
					SELECT id, rev, rev_id, parent_rev, parent_rev_id
					FROM %[1]q AS revs
					WHERE id = $1
						AND rev = $2
						AND rev_id = $3
					UNION ALL
					-- Recursive step: Select the children of the current node
					SELECT r.id, r.rev, r.rev_id, r.parent_rev, r.parent_rev_id
					FROM %[1]q r
					JOIN Descendants d ON d.rev_id = r.parent_rev_id AND d.rev = r.parent_rev AND d.id = r.id
				)
				-- Combine ancestors and descendants, excluding the starting node twice
				SELECT rev.rev, rev.rev_id, doc, deleted
				FROM Descendants AS rev
				JOIN %[2]q AS doc ON doc.id = rev.id AND doc.rev = rev.rev AND doc.rev_id = rev.rev_id
				LEFT JOIN %[1]q AS child ON child.parent_rev = rev.rev AND child.parent_rev_id = rev.rev_id
				WHERE child.rev IS NULL
				ORDER BY rev.rev DESC, rev.rev_id DESC
			`, d.name+"_revs", d.name), id, r.rev, r.id).Scan(&r.rev, &r.id, &body, &deleted)
		} else {
			err = tx.QueryRowContext(ctx, fmt.Sprintf(`
				SELECT doc, deleted
				FROM %q
				WHERE id = $1
					AND rev = $2
					AND rev_id = $3
				`, d.name), id, r.rev, r.id).Scan(&body, &deleted)
		}
	} else {
		err = tx.QueryRowContext(ctx, fmt.Sprintf(`
			SELECT rev, rev_id, doc, deleted
			FROM %q
			WHERE id = $1
			ORDER BY rev DESC, rev_id DESC
			LIMIT 1
		`, d.name), id).Scan(&r.rev, &r.id, &body, &deleted)
	}

	switch {
	case errors.Is(err, sql.ErrNoRows) ||
		deleted && optsRev == "":
		return nil, &internal.Error{Status: http.StatusNotFound, Message: "not found"}
	case err != nil:
		return nil, err
	}

	toMerge := map[string]interface{}{
		"_id":  id,
		"_rev": r.String(),
	}

	if meta, _ := opts["meta"].(bool); meta {
		opts["conflicts"] = true
		opts["deleted_conflicts"] = true
		opts["revs_info"] = true
	}

	if conflicts, _ := opts["conflicts"].(bool); conflicts {
		revs, err := d.conflicts(ctx, tx, id, r, false)
		if err != nil {
			return nil, err
		}

		toMerge["_conflicts"] = revs
	}

	if deletedConflicts, _ := opts["deleted_conflicts"].(bool); deletedConflicts {
		revs, err := d.conflicts(ctx, tx, id, r, true)
		if err != nil {
			return nil, err
		}

		toMerge["_deleted_conflicts"] = revs
	}

	if revsInfo, _ := opts["revs_info"].(bool); revsInfo {
		rows, err := tx.QueryContext(ctx, fmt.Sprintf(`
			SELECT revs.rev || '-' || revs.rev_id,
				CASE
					WHEN doc.id IS NULL THEN 'missing'
					WHEN doc.deleted THEN    'deleted'
					ELSE                     'available'
				END
			FROM (
				WITH RECURSIVE
				Ancestors AS (
					-- Base case: Select the starting node for ancestors
					SELECT id, rev, rev_id, parent_rev, parent_rev_id
					FROM %[1]q AS revs
					WHERE id = $1
						AND rev = $2
						AND rev_id = $3
					UNION ALL
					-- Recursive step: Select the parent of the current node
					SELECT r.id, r.rev, r.rev_id, r.parent_rev, r.parent_rev_id
					FROM %[1]q r
					JOIN Ancestors a ON a.parent_rev_id = r.rev_id AND a.parent_rev = r.rev AND a.id = r.id
				),
				Descendants AS (
					-- Base case: Select the starting node for descendants
					SELECT id, rev, rev_id, parent_rev, parent_rev_id
					FROM %[1]q AS revs
					WHERE id = $1
						AND rev = $2
						AND rev_id = $3
					UNION ALL
					-- Recursive step: Select the children of the current node
					SELECT r.id, r.rev, r.rev_id, r.parent_rev, r.parent_rev_id
					FROM %[1]q r
					JOIN Descendants d ON d.rev_id = r.parent_rev_id AND d.rev = r.parent_rev AND d.id = r.id
				)
				-- Combine ancestors and descendants, excluding the starting node twice
				SELECT id, rev, rev_id FROM Ancestors
				UNION
				SELECT id, rev, rev_id FROM Descendants
			) AS revs
			LEFT JOIN %[2]q AS doc ON doc.id = revs.id
				AND doc.rev = revs.rev
				AND doc.rev_id = revs.rev_id
			ORDER BY revs.rev DESC, revs.rev_id DESC
		`, d.name+"_revs", d.name), id, r.rev, r.id)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var revs []map[string]string
		for rows.Next() {
			var rev, status string
			if err := rows.Scan(&rev, &status); err != nil {
				return nil, err
			}
			revs = append(revs, map[string]string{
				"rev":    rev,
				"status": status,
			})
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		toMerge["_revs_info"] = revs
	}

	if len(toMerge) > 0 {
		body, err = mergeIntoDoc(body, toMerge)
		if err != nil {
			return nil, err
		}
	}

	return &driver.Document{
		Rev:  r.String(),
		Body: io.NopCloser(bytes.NewReader(body)),
	}, tx.Commit()
}

func (d *db) conflicts(ctx context.Context, tx *sql.Tx, id string, r revision, deleted bool) ([]string, error) {
	var revs []string
	rows, err := tx.QueryContext(ctx, fmt.Sprintf(`
			SELECT rev.rev || '-' || rev.rev_id
			FROM %[1]q AS rev
			LEFT JOIN %[1]q AS child
				ON rev.id = child.id
				AND rev.rev = child.parent_rev
				AND rev.rev_id = child.parent_rev_id
			JOIN %[2]q AS docs ON docs.id = rev.id
				AND docs.rev = rev.rev
				AND docs.rev_id = rev.rev_id
			WHERE rev.id = $1
				AND NOT (rev.rev = $2 AND rev.rev_id = $3)
				AND child.id IS NULL
				AND docs.deleted = $4
			`, d.name+"_revs", d.name), id, r.rev, r.id, deleted)
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
	return revs, rows.Err()
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
