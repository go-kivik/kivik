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

func (d *db) Get(ctx context.Context, id string, options driver.Options) (*driver.Document, error) {
	opts := map[string]interface{}{}
	options.Apply(opts)

	var (
		r        revision
		body     []byte
		err      error
		deleted  bool
		localSeq int
	)

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	optsRev, _ := opts["rev"].(string)
	latest, _ := opts["latest"].(bool)
	if optsRev != "" {
		r, err = parseRev(optsRev)
		if err != nil {
			return nil, err
		}
	}
	switch {
	case optsRev != "" && !latest:
		err = tx.QueryRowContext(ctx, fmt.Sprintf(`
			SELECT seq, doc, deleted
			FROM %q
			WHERE id = $1
				AND rev = $2
				AND rev_id = $3
			`, d.name), id, r.rev, r.id).Scan(&localSeq, &body, &deleted)
	case optsRev != "" && latest:
		err = tx.QueryRowContext(ctx, fmt.Sprintf(`
			SELECT seq, rev, rev_id, doc, deleted FROM (
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
				SELECT seq, rev.rev, rev.rev_id, doc, deleted
				FROM Descendants AS rev
				JOIN %[2]q AS doc ON doc.id = rev.id AND doc.rev = rev.rev AND doc.rev_id = rev.rev_id
				LEFT JOIN %[1]q AS child ON child.parent_rev = rev.rev AND child.parent_rev_id = rev.rev_id
				WHERE child.rev IS NULL
					AND doc.deleted = FALSE
				ORDER BY rev.rev DESC, rev.rev_id DESC
			)
			UNION ALL
			-- This query fetches the winning non-deleted rev, in case the above
			-- query returns nothing, because the latest leaf rev is deleted.
			SELECT seq, rev, rev_id, doc, deleted FROM (
				SELECT leaf.id, leaf.rev, leaf.rev_id, leaf.parent_rev, leaf.parent_rev_id, doc.doc, doc.deleted, doc.seq
				FROM %[1]q AS leaf
				LEFT JOIN %[1]q AS child ON child.id = leaf.id AND child.parent_rev = leaf.rev AND child.parent_rev_id = leaf.rev_id
				JOIN %[2]q AS doc ON doc.id = leaf.id AND doc.rev = leaf.rev AND doc.rev_id = leaf.rev_id
				WHERE child.rev IS NULL
					AND doc.deleted = FALSE
				ORDER BY leaf.rev DESC, leaf.rev_id DESC
			)
			LIMIT 1
		`, d.name+"_revs", d.name), id, r.rev, r.id).Scan(&localSeq, &r.rev, &r.id, &body, &deleted)
	default:
		err = tx.QueryRowContext(ctx, fmt.Sprintf(`
			SELECT seq, rev, rev_id, doc, deleted
			FROM %q
			WHERE id = $1
			ORDER BY rev DESC, rev_id DESC
			LIMIT 1
		`, d.name), id).Scan(&localSeq, &r.rev, &r.id, &body, &deleted)
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

	var (
		optConflicts, _        = opts["conflicts"].(bool)
		optDeletedConflicts, _ = opts["deleted_conflicts"].(bool)
		optRevsInfo, _         = opts["revs_info"].(bool)
		optRevs, _             = opts["revs"].(bool)
		optLocalSeq, _         = opts["local_seq"].(bool)
	)

	if meta, _ := opts["meta"].(bool); meta {
		optConflicts = true
		optDeletedConflicts = true
		optRevsInfo = true
	}

	if optConflicts {
		revs, err := d.conflicts(ctx, tx, id, r, false)
		if err != nil {
			return nil, err
		}

		toMerge["_conflicts"] = revs
	}

	if optDeletedConflicts {
		revs, err := d.conflicts(ctx, tx, id, r, true)
		if err != nil {
			return nil, err
		}

		toMerge["_deleted_conflicts"] = revs
	}

	if optRevsInfo || optRevs {
		rows, err := tx.QueryContext(ctx, fmt.Sprintf(`
			SELECT revs.rev, revs.rev_id,
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
		type revStatus struct {
			rev    int
			id     string
			status string
		}
		var revs []revStatus
		for rows.Next() {
			var rs revStatus
			if err := rows.Scan(&rs.rev, &rs.id, &rs.status); err != nil {
				return nil, err
			}
			revs = append(revs, rs)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		if optRevsInfo {
			info := make([]map[string]string, 0, len(revs))
			for _, r := range revs {
				info = append(info, map[string]string{
					"rev":    fmt.Sprintf("%d-%s", r.rev, r.id),
					"status": r.status,
				})
			}
			toMerge["_revs_info"] = info
		} else {
			// for revs=true, we include a different format of this data
			revsInfo := revsInfo{
				Start: revs[0].rev,
				IDs:   make([]string, len(revs)),
			}
			for i, r := range revs {
				revsInfo.IDs[i] = r.id
			}
			toMerge["_revisions"] = revsInfo
		}
	}
	if optLocalSeq {
		toMerge["_local_seq"] = localSeq
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
