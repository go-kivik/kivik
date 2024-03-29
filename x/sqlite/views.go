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
	"fmt"
	"io"
	"strings"

	"github.com/go-kivik/kivik/v4/driver"
)

func endKeyOp(descending, inclusive bool) string {
	switch {
	case descending && inclusive:
		return ">="
	case descending && !inclusive:
		return ">"
	case !descending && inclusive:
		return "<="
	case !descending && !inclusive:
		return "<"
	}
	panic("unreachable")
}

func startKeyOp(descending bool) string {
	if descending {
		return "<="
	}
	return ">="
}

func (d *db) AllDocs(ctx context.Context, options driver.Options) (driver.Rows, error) {
	opts := newOpts(options)

	var (
		optConflicts, _   = opts["conflicts"].(bool)
		optDescending, _  = opts["descending"].(bool)
		optIncludeDocs, _ = opts["include_docs"].(bool)
	)

	direction := "ASC"
	if optDescending {
		direction = "DESC"
	}

	args := []interface{}{optIncludeDocs}

	where := []string{"rev.rank = 1"}
	if endkey := opts.endKey(); endkey != "" {
		where = append(where, fmt.Sprintf("rev.id %s $%d", endKeyOp(optDescending, opts.inclusiveEnd()), len(args)+1))
		args = append(args, endkey)
	}
	if startkey := opts.startKey(); startkey != "" {
		where = append(where, fmt.Sprintf("rev.id %s $%d", startKeyOp(optDescending), len(args)+1))
		args = append(args, startkey)
	}

	query := fmt.Sprintf(d.query(`
		WITH RankedRevisions AS (
			SELECT
				id                    AS id,
				rev                   AS rev,
				rev_id                AS rev_id,
				IIF($1, doc, NULL)    AS doc,
				deleted               AS deleted,
				ROW_NUMBER() OVER (PARTITION BY id ORDER BY rev DESC, rev_id DESC) AS rank
			FROM {{ .Leaves }} AS rev
			WHERE NOT deleted
		)
		SELECT
			rev.id                       AS id,
			rev.rev || '-' || rev.rev_id AS rev,
			rev.doc                      AS doc,
			GROUP_CONCAT(conflicts.rev || '-' || conflicts.rev_id, ',') AS conflicts
		FROM RankedRevisions AS rev
		LEFT JOIN (
			SELECT rev.id, rev.rev, rev.rev_id
			FROM {{ .Revs }} AS rev
			LEFT JOIN {{ .Revs }} AS child
				ON child.id = rev.id
				AND rev.rev = child.parent_rev
				AND rev.rev_id = child.parent_rev_id
			WHERE child.id IS NULL
		) AS conflicts ON conflicts.id = rev.id AND NOT (rev.rev = conflicts.rev AND rev.rev_id = conflicts.rev_id)
		WHERE %[2]s
		GROUP BY rev.id, rev.rev, rev.rev_id
		ORDER BY id %[1]s
	`), direction, strings.Join(where, " AND "))
	results, err := d.db.QueryContext(ctx, query, args...) //nolint:rowserrcheck // Err checked in Next
	if err != nil {
		return nil, err
	}

	return &rows{
		ctx:       ctx,
		db:        d,
		rows:      results,
		conflicts: optConflicts,
	}, nil
}

type rows struct {
	ctx       context.Context
	db        *db
	rows      *sql.Rows
	conflicts bool
}

var _ driver.Rows = (*rows)(nil)

func (r *rows) Next(row *driver.Row) error {
	if !r.rows.Next() {
		if err := r.rows.Err(); err != nil {
			return err
		}
		return io.EOF
	}
	var (
		doc       []byte
		conflicts *string
	)
	if err := r.rows.Scan(&row.ID, &row.Rev, &doc, &conflicts); err != nil {
		return err
	}
	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(map[string]interface{}{"value": map[string]string{"rev": row.Rev}})
	row.Value = &buf
	if doc != nil {
		toMerge := fullDoc{
			ID:  row.ID,
			Rev: row.Rev,
			Doc: doc,
		}
		if r.conflicts {
			toMerge.Conflicts = strings.Split(*conflicts, ",")
		}
		row.Doc = mergeIntoDoc(toMerge)
	}
	return nil
}

func (r *rows) Close() error {
	return r.rows.Close()
}

func (r *rows) UpdateSeq() string {
	return ""
}

func (r *rows) Offset() int64 {
	return 0
}

func (r *rows) TotalRows() int64 {
	return 0
}
