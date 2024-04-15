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

func (d *db) GetRev(ctx context.Context, id string, options driver.Options) (string, error) {
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
		return "", &internal.Error{Status: http.StatusNotFound, Message: "not found"}
	case err != nil:
		return "", err
	}

	toMerge := fullDoc{
		ID:      id,
		Rev:     r.String(),
		Deleted: deleted,
	}

	var (
		optConflicts, _        = opts["conflicts"].(bool)
		optDeletedConflicts, _ = opts["deleted_conflicts"].(bool)
		optRevsInfo, _         = opts["revs_info"].(bool)
		optRevs, _             = opts["revs"].(bool) // TODO: opts.revs()
		optLocalSeq, _         = opts["local_seq"].(bool)
		optAttachments, _      = opts["attachments"].(bool)
		optAttsSince, _        = opts["atts_since"].([]string)
	)

	if meta, _ := opts["meta"].(bool); meta {
		optConflicts = true
		optDeletedConflicts = true
		optRevsInfo = true
	}

	if optConflicts {
		revs, err := d.conflicts(ctx, tx, id, r, false)
		if err != nil {
			return "", err
		}

		toMerge.Conflicts = revs
	}

	if optDeletedConflicts {
		revs, err := d.conflicts(ctx, tx, id, r, true)
		if err != nil {
			return "", err
		}

		toMerge.DeletedConflicts = revs
	}

	if optRevsInfo || optRevs {
		rows, err := tx.QueryContext(ctx, d.query(`
			WITH RECURSIVE Ancestors AS (
				-- Base case: Select the starting node for ancestors
				SELECT id, rev, rev_id, parent_rev, parent_rev_id
				FROM {{ .Revs }} AS revs
				WHERE id = $1
					AND rev = $2
					AND rev_id = $3
				UNION ALL
				-- Recursive step: Select the parent of the current node
				SELECT r.id, r.rev, r.rev_id, r.parent_rev, r.parent_rev_id
				FROM {{ .Revs }} r
				JOIN Ancestors a ON a.parent_rev_id = r.rev_id AND a.parent_rev = r.rev AND a.id = r.id
			)
			SELECT revs.rev, revs.rev_id,
				CASE
					WHEN doc.id IS NULL THEN 'missing'
					WHEN doc.deleted THEN    'deleted'
					ELSE                     'available'
				END
			FROM Ancestors AS revs
			LEFT JOIN {{ .Docs }} AS doc ON doc.id = revs.id
				AND doc.rev = revs.rev
				AND doc.rev_id = revs.rev_id
			ORDER BY revs.rev DESC, revs.rev_id DESC
		`), id, r.rev, r.id)
		if err != nil {
			return "", err
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
				return "", err
			}
			revs = append(revs, rs)
		}
		if err := rows.Err(); err != nil {
			return "", err
		}
		if optRevsInfo {
			info := make([]map[string]string, 0, len(revs))
			for _, r := range revs {
				info = append(info, map[string]string{
					"rev":    fmt.Sprintf("%d-%s", r.rev, r.id),
					"status": r.status,
				})
			}
			toMerge.RevsInfo = info
		} else {
			// for revs=true, we include a different format of this data
			revsInfo := revsInfo{
				Start: revs[0].rev,
				IDs:   make([]string, len(revs)),
			}
			for i, r := range revs {
				revsInfo.IDs[i] = r.id
			}
			toMerge.Revisions = &revsInfo
		}
	}
	if optLocalSeq {
		toMerge.LocalSeq = localSeq
	}
	atts, err := d.getAttachments(ctx, tx, id, r, optAttachments, optAttsSince)
	if err != nil {
		return "", err
	}
	if mergeAtts := atts.inlineAttachments(); mergeAtts != nil {
		toMerge.Attachments = mergeAtts
	}

	toMerge.Doc = body

	return r.String(), tx.Commit()
}
