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
	"io"

	"github.com/go-kivik/kivik/v4/driver"
)

func (d *db) AllDocs(ctx context.Context, _ driver.Options) (driver.Rows, error) {
	results, err := d.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT
			id
		FROM %[1]q
	`, d.name))
	if err != nil {
		return nil, err
	}

	return &rows{rows: results}, nil
}

type rows struct {
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
	return r.rows.Scan(&row.ID)
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
