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
	"encoding/json"
	"io"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

func (d *db) updateReduce(ctx context.Context, tx *sql.Tx, ddoc, view string, rev revision, reduceFuncJS *string) error {
	reduceFn, err := d.reduceFunc(reduceFuncJS, d.logger)
	if err != nil {
		return err
	}
	if reduceFn == nil {
		return nil
	}

	if _, err := tx.ExecContext(ctx, d.ddocQuery(ddoc, view, rev.String(), `
		DELETE FROM {{ .Reduce }}
	`)); err != nil {
		return err
	}

	rows, err := tx.QueryContext(ctx, d.ddocQuery(ddoc, view, rev.String(), `
		SELECT id, key, value
		FROM {{ .Map }}
		ORDER BY id, key
	`))
	if err != nil {
		return err
	}
	defer rows.Close()

	var (
		keys   [][2]interface{}
		values []interface{}

		id, key, value *string
	)

	for rows.Next() {
		if err := rows.Scan(&id, &key, &value); err != nil {
			return err
		}
		keys = append(keys, [2]interface{}{id, key})
		if value == nil {
			values = append(values, nil)
		} else {
			values = append(values, *value)
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	rv := reduceFn(keys, values, false)
	var rvJSON *json.RawMessage
	if rv != nil {
		tmp, _ := json.Marshal(rv)
		rvJSON = (*json.RawMessage)(&tmp)
	}

	if _, err := tx.ExecContext(ctx, d.ddocQuery(ddoc, view, rev.String(), `
			INSERT INTO {{ .Reduce }} (min_key, max_key, value)
			VALUES ($1, $2, $3)
		`), nil, kivik.EndKeySuffix, rvJSON); err != nil {
		return err
	}

	return nil
}

type reducedRows []driver.Row

var _ driver.Rows = (*reducedRows)(nil)

func (r *reducedRows) Close() error {
	*r = nil
	return nil
}

func (r *reducedRows) Next(row *driver.Row) error {
	if len(*r) == 0 {
		return io.EOF
	}
	*row = (*r)[0]
	*r = (*r)[1:]
	return nil
}

func (*reducedRows) Offset() int64     { return 0 }
func (*reducedRows) TotalRows() int64  { return 0 }
func (*reducedRows) UpdateSeq() string { return "" }
