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
	"sort"
	"strings"

	"github.com/go-kivik/kivik/v4/internal"
)

func placeholders(start, count int) string {
	result := make([]string, count)
	for i := range result {
		result[i] = fmt.Sprintf("$%d", start+i)
	}
	return strings.Join(result, ", ")
}

// isLeafRev returns a nil error if the specified revision is a leaf revision.
// If the revision is not a leaf revision, it returns a conflict error.
// If the document is not found, it returns a not found error.
func (d *db) isLeafRev(ctx context.Context, tx *sql.Tx, docID string, rev int, revID string) error {
	var exists bool
	err := tx.QueryRowContext(ctx, d.query(`
		SELECT
			parent.parent_rev IS NOT NULL
		FROM {{ .Revs }} AS r
		LEFT JOIN {{ .Revs }} AS parent
			ON r.id = parent.id
			AND r.parent_rev = parent.parent_rev
			AND r.parent_rev_id = parent.parent_rev_id
		WHERE r.id = $1
			AND r.rev = $2
			AND r.rev_id = $3
	`), docID, rev, revID).Scan(&exists)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return &internal.Error{Status: http.StatusNotFound, Message: "document not found"}
	case err != nil:
		return err
	}
	if !exists {
		return &internal.Error{Status: http.StatusConflict, Message: "Document update conflict"}
	}
	return nil
}

// winningRev returns the current winning revision for the specified document.
func (d *db) winningRev(ctx context.Context, tx *sql.Tx, docID string) (revision, error) {
	var curRev revision
	err := tx.QueryRowContext(ctx, d.query(`
		SELECT rev, rev_id
		FROM {{ .Docs }}
		WHERE id = $1
		ORDER BY rev DESC, rev_id DESC
		LIMIT 1
	`), docID).Scan(&curRev.rev, &curRev.id)
	return curRev, err
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
	err := tx.QueryRowContext(ctx, d.query(`
		INSERT INTO {{ .Revs }} (id, rev, rev_id, parent_rev, parent_rev_id)
		SELECT $1, COALESCE(MAX(rev),0) + 1, $2, $3, $4
		FROM {{ .Revs }}
		WHERE id = $1
		RETURNING rev, rev_id
	`), data.ID, data.RevID, curRevRev, curRevID).Scan(&r.rev, &r.id)
	if err != nil {
		return r, err
	}
	if len(data.Doc) == 0 {
		// No body can happen for example when calling PutAttachment, so we
		// create the new docs table entry by reading the previous one.
		_, err = tx.ExecContext(ctx, d.query(`
			INSERT INTO {{ .Docs }} (id, rev, rev_id, doc, deleted)
			SELECT $1, $2, $3, doc, deleted
			FROM {{ .Docs }}
			WHERE id = $1
				AND rev = $2-1
				AND rev_id = $3
			`), data.ID, r.rev, r.id)
	} else {
		_, err = tx.ExecContext(ctx, d.query(`
			INSERT INTO {{ .Docs }} (id, rev, rev_id, doc, deleted)
			VALUES ($1, $2, $3, $4, $5)
		`), data.ID, r.rev, r.id, data.Doc, data.Deleted)
	}
	if err != nil {
		return r, err
	}

	if len(data.Attachments) == 0 {
		return r, nil
	}

	// order the filenames to insert for consistency
	orderedFilenames := make([]string, 0, len(data.Attachments))
	for filename := range data.Attachments {
		orderedFilenames = append(orderedFilenames, filename)
	}
	sort.Strings(orderedFilenames)

	attStmt, err := tx.PrepareContext(ctx, d.query(`
			INSERT INTO {{ .Attachments }} (rev_pos, filename, content_type, length, digest, data)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING pk
		`))
	if err != nil {
		return r, err
	}
	defer attStmt.Close()

	stubStmt, err := tx.PrepareContext(ctx, d.query(`
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
		return r, err
	}
	defer stubStmt.Close()

	bridgeStmt, err := tx.PrepareContext(ctx, d.query(`
			INSERT INTO {{ .AttachmentsBridge }} (pk, id, rev, rev_id)
			VALUES ($1, $2, $3, $4)
		`))
	if err != nil {
		return r, err
	}
	defer bridgeStmt.Close()

	for _, filename := range orderedFilenames {
		att := data.Attachments[filename]

		var pk int
		if att.Stub {
			err := stubStmt.QueryRowContext(ctx, data.ID, r.rev, r.id, curRev.rev, curRev.id, filename).Scan(&pk)
			if err != nil {
				return r, err
			}
		} else {
			if err := att.calculate(filename); err != nil {
				return r, err
			}
			contentType := att.ContentType
			if contentType == "" {
				contentType = "application/octet-stream"
			}
			err := attStmt.QueryRowContext(ctx, r.rev, filename, contentType, att.Length, att.Digest, att.Content).Scan(&pk)
			if err != nil {
				return r, err
			}

			_, err = bridgeStmt.ExecContext(ctx, pk, data.ID, r.rev, r.id)
			if err != nil {
				return r, err
			}
		}
	}

	return r, nil
}

// docRevExists returns an error if the requested document does not exist. It
// returns false if the document does exist, but the specified revision is not
// the latest. It returns true, nil if both the doc and revision are valid.
func (d *db) docRevExists(ctx context.Context, tx *sql.Tx, docID string, rev revision) (bool, error) {
	var found bool
	err := tx.QueryRowContext(ctx, d.query(`
		SELECT child.id IS NULL
		FROM {{ .Revs }} AS rev
		LEFT JOIN {{ .Revs }} AS child ON rev.id = child.id AND rev.rev = child.parent_rev AND rev.rev_id = child.parent_rev_id
		JOIN {{ .Docs }} AS doc ON rev.id = doc.id AND rev.rev = doc.rev AND rev.rev_id = doc.rev_id
		WHERE rev.id = $1
			AND rev.rev = $2
			AND rev.rev_id = $3
	`), docID, rev.rev, rev.id).Scan(&found)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return false, &internal.Error{Status: http.StatusNotFound, Message: "document not found"}
	case err != nil:
		return false, err
	}
	return found, nil
}
