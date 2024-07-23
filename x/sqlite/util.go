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
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	internal "github.com/go-kivik/kivik/v4/int/errors"
)

func placeholders(start, count int) string {
	result := make([]string, count)
	for i := range result {
		result[i] = fmt.Sprintf("$%d", start+i)
	}
	return strings.Join(result, ", ")
}

// isLeafRev returns the leaf rev's hash and a nil error if the specified
// revision is a leaf revision. If the revision is not a leaf revision, it
// returns a conflict error. If the document is not found, it returns a not
// found error.
func (d *db) isLeafRev(ctx context.Context, tx *sql.Tx, docID string, rev int, revID string) (md5sum, error) {
	var isLeaf bool
	var hash md5sum
	err := tx.QueryRowContext(ctx, d.query(`
		SELECT
			doc.md5sum,
			COALESCE(revs.is_leaf, FALSE) AS is_leaf
		FROM (
			SELECT
				id,
				rev,
				rev_id,
				md5sum
			FROM {{ .Docs }}
			WHERE id = $1
			LIMIT 1
		) AS doc
		LEFT JOIN (
			SELECT
				parent.id,
				parent.rev,
				parent.rev_id,
				child.id IS NULL AS is_leaf
			FROM {{ .Revs }} AS parent
			LEFT JOIN {{ .Revs }} AS child ON parent.id = child.id AND parent.rev = child.parent_rev AND parent.rev_id = child.parent_rev_id
			WHERE parent.id = $1
				AND parent.rev = $2
				AND parent.rev_id = $3
		) AS revs ON doc.id=revs.id
	`), docID, rev, revID).Scan(&hash, &isLeaf)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return md5sum{}, &internal.Error{Status: http.StatusNotFound, Message: "document not found"}
	case err != nil:
		return md5sum{}, err
	}
	if !isLeaf {
		return md5sum{}, &internal.Error{Status: http.StatusConflict, Message: "document update conflict"}
	}
	return hash, nil
}

// winningRev returns the current winning revision for the specified document.
func (d *db) winningRev(ctx context.Context, tx queryer, docID string) (revision, error) {
	var curRev revision
	err := tx.QueryRowContext(ctx, d.query(`
		SELECT rev, rev_id
		FROM {{ .Docs }}
		WHERE id = $1
		ORDER BY rev DESC, rev_id DESC
		LIMIT 1
	`), docID).Scan(&curRev.rev, &curRev.id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return curRev, &internal.Error{Status: http.StatusNotFound, Message: "missing"}
	}
	return curRev, d.errDatabaseNotFound(err)
}

type stmtCache map[string]*sql.Stmt

func newStmtCache() stmtCache {
	return make(stmtCache)
}

func (c stmtCache) prepare(ctx context.Context, tx *sql.Tx, query string) (*sql.Stmt, error) {
	stmt, ok := c[query]
	if !ok {
		var err error
		stmt, err = tx.PrepareContext(ctx, query)
		if err != nil {
			return nil, err
		}
		c[query] = stmt
	}
	return stmt, nil
}

// createRev creates a new entry in the revs table, inserts the document data
// into the docs table, attachments into the attachments table, and returns the
// new revision.
func (d *db) createRev(ctx context.Context, tx *sql.Tx, data *docData, curRev revision) (revision, error) {
	var (
		r         revision
		curRevRev *int
		curRevID  *string
	)
	if curRev.rev != 0 {
		curRevRev = &curRev.rev
		curRevID = &curRev.id
	}
	r.id = data.RevID()
	err := tx.QueryRowContext(ctx, d.query(`
		INSERT INTO {{ .Revs }} (id, rev, rev_id, parent_rev, parent_rev_id)
		VALUES ($1, COALESCE($3, 0) + 1, $2, $3, $4)
		RETURNING rev
	`), data.ID, r.id, curRevRev, curRevID).Scan(&r.rev)
	if err != nil {
		return r, err
	}
	if len(data.Doc) == 0 {
		// No body can happen for example when calling PutAttachment, so we
		// create the new docs table entry by reading the previous one.
		_, err = tx.ExecContext(ctx, d.query(`
			INSERT INTO {{ .Docs }} (id, rev, rev_id, doc, md5sum, deleted)
			SELECT $1, $2, $3, doc, md5sum, deleted
			FROM {{ .Docs }}
			WHERE id = $1
				AND rev = $4
				AND rev_id = $5
			`), data.ID, r.rev, r.id, curRev.rev, curRev.id)
	} else {
		_, err = tx.ExecContext(ctx, d.query(`
			INSERT INTO {{ .Docs }} (id, rev, rev_id, doc, md5sum, deleted)
			VALUES ($1, $2, $3, $4, $5, $6)
		`), data.ID, r.rev, r.id, data.Doc, data.MD5sum, data.Deleted)
	}
	if err != nil {
		return r, err
	}

	// order the filenames to insert for consistency
	if err := d.createDocAttachments(ctx, data, tx, r, &curRev); err != nil {
		return r, err
	}
	err = d.updateDesignDoc(ctx, tx, r, data)
	return r, err
}

func (d *db) createDocAttachments(ctx context.Context, data *docData, tx *sql.Tx, r revision, curRev *revision) error {
	if len(data.Attachments) == 0 {
		return nil
	}
	orderedFilenames := make([]string, 0, len(data.Attachments))
	for filename := range data.Attachments {
		orderedFilenames = append(orderedFilenames, filename)
	}
	sort.Strings(orderedFilenames)

	stmts := newStmtCache()

	for _, filename := range orderedFilenames {
		att := data.Attachments[filename]

		var pk int
		if att.Stub {
			if curRev == nil {
				return &internal.Error{Status: http.StatusPreconditionFailed, Message: fmt.Sprintf("invalid attachment stub in %s for %s", data.ID, filename)}
			}
			stubStmt, err := stmts.prepare(ctx, tx, d.query(`
				INSERT INTO {{ .AttachmentsBridge }} (pk, id, rev, rev_id)
				SELECT att.pk, $1, $2, $3
				FROM {{ .Attachments }} AS att
				JOIN {{ .AttachmentsBridge }} AS bridge ON att.pk = bridge.pk
				WHERE bridge.id = $1
					AND bridge.rev = $4
					AND bridge.rev_id = $5
					AND att.filename = $6
				RETURNING pk
			`))
			if err != nil {
				return err
			}
			err = stubStmt.QueryRowContext(ctx, data.ID, r.rev, r.id, curRev.rev, curRev.id, filename).Scan(&pk)
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return &internal.Error{Status: http.StatusPreconditionFailed, Message: fmt.Sprintf("invalid attachment stub in %s for %s", data.ID, filename)}
			case err != nil:
				return err
			}
		} else {
			if err := att.calculate(filename); err != nil {
				return err
			}
			contentType := att.ContentType
			if contentType == "" {
				contentType = "application/octet-stream"
			}

			attStmt, err := stmts.prepare(ctx, tx, d.query(`
				INSERT INTO {{ .Attachments }} (rev_pos, filename, content_type, length, digest, data)
				VALUES ($1, $2, $3, $4, $5, $6)
				RETURNING pk
			`))
			if err != nil {
				return err
			}

			err = attStmt.QueryRowContext(ctx, r.rev, filename, contentType, att.Length, att.Digest, att.Content).Scan(&pk)
			if err != nil {
				return err
			}

			bridgeStmt, err := stmts.prepare(ctx, tx, d.query(`
				INSERT INTO {{ .AttachmentsBridge }} (pk, id, rev, rev_id)
				VALUES ($1, $2, $3, $4)
			`))
			if err != nil {
				return err
			}
			_, err = bridgeStmt.ExecContext(ctx, pk, data.ID, r.rev, r.id)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *db) lastSeq(ctx context.Context) (uint64, error) {
	var lastSeq uint64
	err := d.db.QueryRowContext(ctx, d.query(`
			SELECT COALESCE(MAX(seq), 0) FROM {{ .Docs }}
		`)).Scan(&lastSeq)
	return lastSeq, d.errDatabaseNotFound(err)
}

// discard implements the [database/sql.Scanner] interface, but discards the
// value. Useful when your query returns rows you don't always need.
type discard struct{}

func (discard) Scan(interface{}) error {
	return nil
}

// md5sumString returns the hex-encoded MD5 sum of s.
func md5sumString(s string) string {
	h := md5.New()
	_, _ = h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
