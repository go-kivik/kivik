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

const (
	viewAllDocs    = "_all_docs"
	viewLocalDocs  = "_local_docs"
	viewDesignDocs = "_design_docs"
)

func isBuiltinView(view string) bool {
	switch view {
	case viewAllDocs, viewLocalDocs, viewDesignDocs:
		return true
	}
	return false
}

func (d *db) AllDocs(ctx context.Context, options driver.Options) (driver.Rows, error) {
	return d.Query(ctx, viewAllDocs, "", options)
}

func (d *db) LocalDocs(ctx context.Context, options driver.Options) (driver.Rows, error) {
	return d.Query(ctx, viewLocalDocs, "", options)
}

func (d *db) DesignDocs(ctx context.Context, options driver.Options) (driver.Rows, error) {
	return d.Query(ctx, viewDesignDocs, "", options)
}

func (d *db) queryBuiltinView(
	ctx context.Context,
	vopts *viewOptions,
) (driver.Rows, error) {
	args := []interface{}{vopts.includeDocs, vopts.conflicts}

	where := append([]string{""}, vopts.buildWhere(&args)...)

	query := fmt.Sprintf(d.query(leavesCTE+`
		SELECT *
		FROM (
			SELECT
				rev.id                       AS id,
				rev.id                       AS key,
				'{"value":{"rev":"' || rev.rev || '-' || rev.rev_id || '"}}' AS value,
				rev.rev || '-' || rev.rev_id AS rev,
				rev.doc                      AS doc,
				IIF($2, GROUP_CONCAT(conflicts.rev || '-' || conflicts.rev_id, ','), NULL) AS conflicts
			FROM (
				SELECT
					id                    AS id,
					rev                   AS rev,
					rev_id                AS rev_id,
					IIF($1, doc, NULL)    AS doc,
					deleted               AS deleted, -- TODO:remove this?
					ROW_NUMBER() OVER (PARTITION BY id ORDER BY rev DESC, rev_id DESC) AS rank
				FROM leaves
			) AS rev
			LEFT JOIN leaves AS conflicts ON conflicts.id = rev.id AND NOT (rev.rev = conflicts.rev AND rev.rev_id = conflicts.rev_id)
			WHERE rev.rank = 1
				%[2]s
			GROUP BY rev.id, rev.rev, rev.rev_id
			ORDER BY key %[1]s
			LIMIT %[3]d OFFSET %[4]d
		)
	`), descendingToDirection(vopts.descending), strings.Join(where, " AND "), vopts.limit, vopts.skip)
	results, err := d.db.QueryContext(ctx, query, args...) //nolint:rowserrcheck // Err checked in Next
	if err != nil {
		return nil, err
	}

	return &rows{
		ctx:  ctx,
		db:   d,
		rows: results,
	}, nil
}

func descendingToDirection(descending bool) string {
	if descending {
		return "DESC"
	}
	return "ASC"
}

type rows struct {
	ctx  context.Context
	db   *db
	rows *sql.Rows
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
		id        *string
		key, doc  []byte
		value     *[]byte
		conflicts *string
		rev       string
	)
	if err := r.rows.Scan(&id, &key, &value, &rev, &doc, &conflicts); err != nil {
		return err
	}
	if id != nil {
		row.ID = *id
	}
	row.Key = key
	if len(key) == 0 {
		row.Key = []byte("null")
	}
	if value == nil {
		row.Value = strings.NewReader("null")
	} else {
		row.Value = bytes.NewReader(*value)
	}
	if doc != nil {
		toMerge := fullDoc{
			ID:  row.ID,
			Rev: rev,
			Doc: doc,
		}
		if conflicts != nil {
			toMerge.Conflicts = strings.Split(*conflicts, ",")
		}
		row.Doc = toMerge.toReader()
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
