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
	"encoding/json"
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
		deleted  bool
		localSeq int
	)

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var (
		optConflicts, _        = opts["conflicts"].(bool)
		optDeletedConflicts, _ = opts["deleted_conflicts"].(bool)
		optRevsInfo, _         = opts["revs_info"].(bool)
		optRevs, _             = opts["revs"].(bool) // TODO: opts.revs()
		optLocalSeq, _         = opts["local_seq"].(bool)
		optAttachments, _      = opts["attachments"].(bool)
		optAttsSince, _        = opts["atts_since"].([]string)
		optsRev, _             = opts["rev"].(string)
		latest, _              = opts["latest"].(bool)
	)

	if optsRev != "" {
		r, err = parseRev(optsRev)
		if err != nil {
			return nil, err
		}
	}
	switch {
	case optsRev != "" && !latest:
		err = tx.QueryRowContext(ctx, d.query(`
			SELECT seq, doc, deleted
			FROM {{ .Docs }}
			WHERE id = $1
				AND rev = $2
				AND rev_id = $3
			`), id, r.rev, r.id).Scan(&localSeq, &body, &deleted)
	case optsRev != "" && latest:
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
		`), id, r.rev, r.id).Scan(&localSeq, &r.rev, &r.id, &body, &deleted)
	default:
		err = tx.QueryRowContext(ctx, d.query(`
			SELECT seq, rev, rev_id, doc, deleted
			FROM {{ .Docs }}
			WHERE id = $1
			ORDER BY rev DESC, rev_id DESC
			LIMIT 1
		`), id).Scan(&localSeq, &r.rev, &r.id, &body, &deleted)
	}

	switch {
	case errors.Is(err, sql.ErrNoRows) ||
		deleted && optsRev == "":
		return nil, &internal.Error{Status: http.StatusNotFound, Message: "not found"}
	case err != nil:
		return nil, err
	}

	toMerge := fullDoc{
		ID:       id,
		Rev:      r.String(),
		Deleted:  deleted,
		Doc:      body,
		LocalSeq: localSeq,
	}
	if !optLocalSeq {
		toMerge.LocalSeq = 0
	}

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

		toMerge.Conflicts = revs
	}

	if optDeletedConflicts {
		revs, err := d.conflicts(ctx, tx, id, r, true)
		if err != nil {
			return nil, err
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

	atts, err := d.getAttachments(ctx, tx, id, r, optAttachments, optAttsSince)
	if err != nil {
		return nil, err
	}
	if mergeAtts := atts.inlineAttachments(); mergeAtts != nil {
		toMerge.Attachments = mergeAtts
	}

	return &driver.Document{
		Attachments: atts,
		Rev:         r.String(),
		Body:        toMerge.toReader(),
	}, tx.Commit()
}

type dbOrTx interface {
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
}

func (d *db) conflicts(ctx context.Context, tx dbOrTx, id string, r revision, deleted bool) ([]string, error) {
	var revs []string
	rows, err := tx.QueryContext(ctx, d.query(`
			SELECT rev.rev || '-' || rev.rev_id
			FROM {{ .Revs }} AS rev
			LEFT JOIN {{ .Revs }} AS child
				ON rev.id = child.id
				AND rev.rev = child.parent_rev
				AND rev.rev_id = child.parent_rev_id
			JOIN {{ .Docs }} AS docs ON docs.id = rev.id
				AND docs.rev = rev.rev
				AND docs.rev_id = rev.rev_id
			WHERE rev.id = $1
				AND NOT (rev.rev = $2 AND rev.rev_id = $3)
				AND child.id IS NULL
				AND docs.deleted = $4
			`), id, r.rev, r.id, deleted)
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

// getAttachments returns the attachments for the given docID and revision.
// It may return nil if there are no attachments.
func (d *db) getAttachments(ctx context.Context, tx *sql.Tx, id string, rev revision, includeAttachments bool, since []string) (*attachments, error) {
	for _, s := range since {
		if _, err := parseRev(s); err != nil {
			return nil, err
		}
	}
	args := []interface{}{id, rev.rev, rev.id, includeAttachments}
	for _, s := range since {
		args = append(args, s)
	}
	sinceQuery := "FALSE"
	if len(since) > 0 {
		sinceQuery = fmt.Sprintf("a.parent_rev || '-' || a.parent_rev_id IN (%s)", placeholders(len(args)-len(since)+1, len(since)))
	}

	query := fmt.Sprintf(d.query(`
			WITH RECURSIVE ancestors AS (
				SELECT id, rev, rev_id, parent_rev, parent_rev_id
				FROM {{ .Revs }} AS revs

				UNION ALL

				SELECT child.id, child.rev, child.rev_id, a.parent_rev, a.parent_rev_id
				FROM ancestors AS a
				JOIN {{ .Revs }} AS child ON a.id = child.id AND a.rev = child.parent_rev AND a.rev_id = child.parent_rev_id
			)
			SELECT
				att.filename,
				att.content_type,
				att.digest,
				att.length,
				att.rev_pos,
				MAX(IIF($4 OR %s, att.data, NULL)) AS data
			FROM {{ .Attachments }} AS att
			JOIN {{ .AttachmentsBridge }} AS bridge ON att.pk = bridge.pk
			LEFT JOIN ancestors AS a ON att.rev_pos = a.rev
			WHERE bridge.id = $1
				AND bridge.rev = $2
				AND bridge.rev_id = $3
			GROUP BY att.filename, att.content_type, att.digest, att.length, att.rev_pos
		`), sinceQuery)
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var atts attachments
	var digest md5sum
	for rows.Next() {
		var a driver.Attachment
		var data *[]byte
		if err := rows.Scan(&a.Filename, &a.ContentType, &digest, &a.Size, &a.RevPos, &data); err != nil {
			return nil, err
		}
		if data == nil {
			a.Stub = true
		} else {
			a.Content = io.NopCloser(bytes.NewReader(*data))
		}
		a.Digest = digest.Digest()
		atts = append(atts, &a)
	}
	if len(atts) == 0 {
		return nil, rows.Err()
	}
	return &atts, rows.Err()
}

type attachments []*driver.Attachment

var _ driver.Attachments = &attachments{}

func (a *attachments) Next(att *driver.Attachment) error {
	if len(*a) == 0 {
		return io.EOF
	}
	*att = *(*a)[0]
	*a = (*a)[1:]
	return nil
}

func (a *attachments) Close() error {
	*a = nil
	return nil
}

func (a *attachments) inlineAttachments() map[string]*attachment {
	if a == nil || len(*a) == 0 {
		return nil
	}
	atts := make(map[string]*attachment, len(*a))
	for _, att := range *a {
		digest, err := parseDigest(att.Digest)
		if err != nil {
			// This should never happen, as the digest should have been validated
			// when the attachment was created.
			panic(err)
		}
		newAtt := &attachment{
			ContentType: att.ContentType,
			Digest:      digest,
			Length:      att.Size,
			RevPos:      int(att.RevPos),
			Stub:        att.Stub,
		}
		if att.Content != nil {
			var data bytes.Buffer
			_, _ = io.Copy(&data, att.Content)
			newAtt.Data, _ = json.Marshal(data.Bytes())
		}
		atts[att.Filename] = newAtt
	}
	return atts
}
