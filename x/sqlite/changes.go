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
	"io"

	"github.com/go-kivik/kivik/v4/driver"
)

type changes struct {
	rows *sql.Rows
}

var _ driver.Changes = &changes{}

func (c *changes) Next(change *driver.Change) error {
	if !c.rows.Next() {
		if err := c.rows.Err(); err != nil {
			return err
		}
		return io.EOF
	}
	var rev string
	if err := c.rows.Scan(&change.ID, &change.Seq, &change.Deleted, &rev); err != nil {
		return err
	}
	change.Changes = driver.ChangedRevs{rev}
	return nil
}

func (c *changes) Close() error {
	return c.rows.Close()
}

func (c *changes) LastSeq() string {
	return ""
}

func (c *changes) Pending() int64 {
	return 0
}

func (c *changes) ETag() string {
	return ""
}

func (d *db) Changes(ctx context.Context, _ driver.Options) (driver.Changes, error) {
	query := d.query(`
	SELECT
		id,
		seq,
		deleted,
		rev || '-' || rev_id AS rev
	FROM {{ .Docs }}
	ORDER BY seq
`)
	rows, err := d.db.QueryContext(ctx, query) //nolint:rowserrcheck // Err checked in Next
	if err != nil {
		return nil, err
	}

	return &changes{
		rows: rows,
	}, nil
}
