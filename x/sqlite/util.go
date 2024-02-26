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
	"fmt"
	"sort"
	"strings"
)

func placeholders(start, count int) string {
	result := make([]string, count)
	for i := range result {
		result[i] = fmt.Sprintf("$%d", start+i)
	}
	return strings.Join(result, ", ")
}

func (d *db) currentRev(ctx context.Context, tx *sql.Tx, docID string) (revision, error) {
	var curRev revision
	err := tx.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT rev, rev_id
		FROM %q
		WHERE id = $1
		ORDER BY rev DESC, rev_id DESC
		LIMIT 1
	`, d.name), docID).Scan(&curRev.rev, &curRev.id)
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
	err := tx.QueryRowContext(ctx, fmt.Sprintf(`
		INSERT INTO %[1]q (id, rev, rev_id, parent_rev, parent_rev_id)
		SELECT $1, COALESCE(MAX(rev),0) + 1, $2, $3, $4
		FROM %[1]q
		WHERE id = $1
		RETURNING rev, rev_id
	`, d.name+"_revs"), data.ID, data.RevID, curRevRev, curRevID).Scan(&r.rev, &r.id)
	if err != nil {
		return r, err
	}
	if len(data.Doc) == 0 {
		// No body can happen for example when calling PutAttachment, so we
		// create the new docs table entry by reading the previous one.
		_, err = tx.ExecContext(ctx, fmt.Sprintf(`
			INSERT INTO %[1]q (id, rev, rev_id, doc, deleted)
			SELECT $1, $2, $3, doc, deleted
			FROM %[1]q
			WHERE id = $1
				AND rev = $2-1
				AND rev_id = $3
			`, d.name), data.ID, r.rev, r.id)
	} else {
		_, err = tx.ExecContext(ctx, fmt.Sprintf(`
			INSERT INTO %[1]q (id, rev, rev_id, doc, deleted)
			VALUES ($1, $2, $3, $4, $5)
		`, d.name), data.ID, r.rev, r.id, data.Doc, data.Deleted)
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

	stmt, err := tx.PrepareContext(ctx, fmt.Sprintf(`
			INSERT INTO %[1]q (id, rev, rev_id, filename, content_type, length, digest, data)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, d.name+"_attachments"))
	if err != nil {
		return r, err
	}
	defer stmt.Close()
	for _, filename := range orderedFilenames {
		att := data.Attachments[filename]
		if att.Stub {
			continue
		}
		if err := att.calculate(filename); err != nil {
			return r, err
		}
		contentType := att.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		_, err := stmt.ExecContext(ctx, data.ID, r.rev, r.id, filename, contentType, att.Length, att.Digest, att.Content)
		if err != nil {
			return r, err
		}
	}

	if len(data.Doc) == 0 {
		// The only reason that the doc is nil is when we're creating a new
		// attachment, and we should not delete existing attachments in that case.
		return r, nil
	}

	// Delete any attachments not included in the new revision
	args := []interface{}{r.rev, r.id, data.ID}
	for _, filename := range orderedFilenames {
		args = append(args, filename)
	}
	query := fmt.Sprintf(`
			UPDATE %[1]q
			SET deleted_rev = $1, deleted_rev_id = $2
			WHERE id = $3
				AND filename NOT IN (`+placeholders(len(args)-len(orderedFilenames)+1, len(orderedFilenames))+`)
				AND deleted_rev IS NULL
				AND deleted_rev_id IS NULL
		`, d.name+"_attachments")
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return r, err
	}

	return r, nil
}
