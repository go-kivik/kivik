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
	"io"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal"
)

func (d *db) OpenRevs(ctx context.Context, docID string, revs []string, _ driver.Options) (driver.Rows, error) {
	if len(revs) == 1 && revs[0] == "all" {
		query := d.query(`
			SELECT
				leaf.rev || '-' || leaf.rev_id AS rev,
				docs.deleted,
				docs.doc
			FROM (
				SELECT
					parent.id,
					parent.rev,
					parent.rev_id
				FROM {{ .Revs }} AS parent
				LEFT JOIN {{ .Revs }} AS child ON parent.id = child.id AND parent.rev = child.parent_rev AND parent.rev_id = child.parent_rev_id
				WHERE parent.id = $1 AND child.id IS NULL
			) AS leaf
			JOIN {{ .Docs }} AS docs ON leaf.id = docs.id AND leaf.rev = docs.rev AND leaf.rev_id = docs.rev_id
			ORDER BY leaf.rev DESC
		`)
		rows, err := d.db.QueryContext(ctx, query, docID) //nolint:rowserrcheck // Err checked in Next
		if err != nil {
			return nil, err
		}

		// Call rows.Next() to see if we get any results at all. If zero results,
		// we need to return 404 instead of an iterator for the open_revs=all case.
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return nil, err
			}
			return nil, &internal.Error{Message: "missing", Status: http.StatusNotFound}
		}

		return &openRevsRows{
			id:   docID,
			ctx:  ctx,
			pre:  true,
			rows: rows,
		}, nil
	}
	for _, rev := range revs {
		if _, err := parseRev(rev); err != nil {
			return nil, &internal.Error{Message: "invalid rev format", Status: http.StatusBadRequest}
		}
	}

	return nil, errors.New("TODO: not implemented")
}

type openRevsRows struct {
	id  string
	ctx context.Context
	// pre is during instantiation to indicate that the first call to Next has
	// already been done, so Next() should skip the next call to Next()
	pre  bool
	rows *sql.Rows
}

var _ driver.Rows = (*openRevsRows)(nil)

func (r *openRevsRows) Next(row *driver.Row) error {
	if !r.pre && !r.rows.Next() {
		if err := r.rows.Err(); err != nil {
			return err
		}
		return io.EOF
	}
	r.pre = false
	var (
		deleted bool
		doc     []byte
	)
	if err := r.rows.Scan(&row.Rev, &deleted, &doc); err != nil {
		return err
	}
	toMerge := fullDoc{
		ID:      r.id,
		Rev:     row.Rev,
		Doc:     doc,
		Deleted: deleted,
	}
	row.ID = r.id
	row.Doc = toMerge.toReader()
	return nil
}

func (r *openRevsRows) Close() error {
	return r.rows.Close()
}

func (*openRevsRows) Offset() int64     { return 0 }
func (*openRevsRows) UpdateSeq() string { return "" }
func (*openRevsRows) TotalRows() int64  { return 0 }
