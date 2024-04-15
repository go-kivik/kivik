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

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal"
)

func (d *db) GetRev(ctx context.Context, id string, options driver.Options) (string, error) {
	opts := map[string]interface{}{}
	options.Apply(opts)

	var (
		r       revision
		deleted bool
	)

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	optsRev, _ := opts["rev"].(string)
	latest, _ := opts["latest"].(bool)
	if optsRev != "" {
		r, err = parseRev(optsRev)
		if err != nil {
			return "", err
		}
	}
	switch {
	case optsRev != "" && !latest:
		err = tx.QueryRowContext(ctx, d.query(`
			SELECT deleted
			FROM {{ .Docs }}
			WHERE id = $1
				AND rev = $2
				AND rev_id = $3
			`), id, r.rev, r.id).Scan(&deleted)
	case optsRev != "" && latest:
		err = tx.QueryRowContext(ctx, d.query(`
			SELECT rev, rev_id
			FROM (
				WITH RECURSIVE Descendants AS (
					-- Base case: Select the starting node for descendants
					SELECT id, rev, rev_id, parent_rev, parent_rev_id
					FROM {{ .Revs }} AS revs
					WHERE id = $1
						AND rev = $2
						AND rev_id = $3
					UNION ALL
					-- Recursive step: Select the children of the current node
					SELECT r.id, r.rev, r.rev_id, r.parent_rev, r.parent_rev_id
					FROM {{ .Revs }} AS r
					JOIN Descendants d ON d.rev_id = r.parent_rev_id AND d.rev = r.parent_rev AND d.id = r.id
				)
				-- Combine ancestors and descendants, excluding the starting node twice
				SELECT rev.rev, rev.rev_id
				FROM Descendants AS rev
				JOIN {{ .Docs }} AS doc ON doc.id = rev.id AND doc.rev = rev.rev AND doc.rev_id = rev.rev_id
				LEFT JOIN {{ .Revs }} AS child ON child.parent_rev = rev.rev AND child.parent_rev_id = rev.rev_id
				WHERE child.rev IS NULL
					AND doc.deleted = FALSE
				ORDER BY rev.rev DESC, rev.rev_id DESC
			)
			UNION ALL
			-- This query fetches the winning non-deleted rev, in case the above
			-- query returns nothing, because the latest leaf rev is deleted.
			SELECT rev, rev_id
			FROM (
				SELECT leaf.id, leaf.rev, leaf.rev_id, leaf.parent_rev, leaf.parent_rev_id
				FROM {{ .Revs }} AS leaf
				LEFT JOIN {{ .Revs }} AS child ON child.id = leaf.id AND child.parent_rev = leaf.rev AND child.parent_rev_id = leaf.rev_id
				JOIN {{ .Docs }} AS doc ON doc.id = leaf.id AND doc.rev = leaf.rev AND doc.rev_id = leaf.rev_id
				WHERE child.rev IS NULL
					AND doc.deleted = FALSE
				ORDER BY leaf.rev DESC, leaf.rev_id DESC
			)
			LIMIT 1
		`), id, r.rev, r.id).Scan(&r.rev, &r.id)
	default:
		err = tx.QueryRowContext(ctx, d.query(`
			SELECT rev, rev_id, deleted
			FROM {{ .Docs }}
			WHERE id = $1
			ORDER BY rev DESC, rev_id DESC
			LIMIT 1
		`), id).Scan(&r.rev, &r.id, &deleted)
	}

	switch {
	case errors.Is(err, sql.ErrNoRows) ||
		deleted && optsRev == "":
		return "", &internal.Error{Status: http.StatusNotFound, Message: "not found"}
	case err != nil:
		return "", err
	}

	return r.String(), tx.Commit()
}

func (d *db) getCoreDoc(ctx context.Context, tx *sql.Tx, id string, rev revision, latest bool) (*fullDoc, revision, error) {
	var (
		localSeq int
		body     []byte
		deleted  bool
		r        revision
		err      error
	)
	switch {
	case rev.IsZero():
		err = tx.QueryRowContext(ctx, d.query(`
			SELECT seq, rev, rev_id, doc, deleted
			FROM {{ .Docs }}
			WHERE id = $1
			ORDER BY rev DESC, rev_id DESC
			LIMIT 1
		`), id).Scan(&localSeq, &r.rev, &r.id, &body, &deleted)
	case !latest:
		err = tx.QueryRowContext(ctx, d.query(`
			SELECT seq, doc, deleted
			FROM {{ .Docs }}
			WHERE id = $1
				AND rev = $2
				AND rev_id = $3
		`), id, rev.rev, rev.id).Scan(&localSeq, &body, &deleted)
		r = rev
	case latest:
		err = tx.QueryRowContext(ctx, d.query(`
			SELECT seq, rev, rev_id, doc, deleted
			FROM (
				WITH RECURSIVE Descendants AS (
					-- Base case: Select the starting node for descendants
					SELECT id, rev, rev_id, parent_rev, parent_rev_id
					FROM {{ .Revs }} AS revs
					WHERE id = $1
						AND rev = $2
						AND rev_id = $3
					UNION ALL
					-- Recursive step: Select the children of the current node
					SELECT r.id, r.rev, r.rev_id, r.parent_rev, r.parent_rev_id
					FROM {{ .Revs }} AS r
					JOIN Descendants d ON d.rev_id = r.parent_rev_id AND d.rev = r.parent_rev AND d.id = r.id
				)
				-- Combine ancestors and descendants, excluding the starting node twice
				SELECT seq, rev.rev, rev.rev_id, doc, deleted
				FROM Descendants AS rev
				JOIN {{ .Docs }} AS doc ON doc.id = rev.id AND doc.rev = rev.rev AND doc.rev_id = rev.rev_id
				LEFT JOIN {{ .Revs }} AS child ON child.parent_rev = rev.rev AND child.parent_rev_id = rev.rev_id
				WHERE child.rev IS NULL
					AND doc.deleted = FALSE
				ORDER BY rev.rev DESC, rev.rev_id DESC
			)
			UNION ALL
			-- This query fetches the winning non-deleted rev, in case the above
			-- query returns nothing, because the latest leaf rev is deleted.
			SELECT seq, rev, rev_id, doc, deleted
			FROM (
				SELECT leaf.id, leaf.rev, leaf.rev_id, leaf.parent_rev, leaf.parent_rev_id, doc.doc, doc.deleted, doc.seq
				FROM {{ .Revs }} AS leaf
				LEFT JOIN {{ .Revs }} AS child ON child.id = leaf.id AND child.parent_rev = leaf.rev AND child.parent_rev_id = leaf.rev_id
				JOIN {{ .Docs }} AS doc ON doc.id = leaf.id AND doc.rev = leaf.rev AND doc.rev_id = leaf.rev_id
				WHERE child.rev IS NULL
					AND doc.deleted = FALSE
				ORDER BY leaf.rev DESC, leaf.rev_id DESC
			)
			LIMIT 1
		`), id, rev.rev, rev.id).Scan(&localSeq, &r.rev, &r.id, &body, &deleted)
	}
	switch {
	case errors.Is(err, sql.ErrNoRows) || deleted && rev.IsZero():
		return nil, revision{}, &internal.Error{Status: http.StatusNotFound, Message: "not found"}
	case err != nil:
		return nil, revision{}, err
	}

	return &fullDoc{
		ID:       id,
		Rev:      r.String(),
		Deleted:  deleted,
		Doc:      body,
		LocalSeq: localSeq,
	}, r, nil
}
