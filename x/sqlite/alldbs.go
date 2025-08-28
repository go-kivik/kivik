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

	"github.com/go-kivik/kivik/v4/driver"
)

func (c *client) AllDBs(ctx context.Context, _ driver.Options) ([]string, error) {
	rows, err := c.db.QueryContext(ctx, `
		SELECT
			name
		FROM
			sqlite_schema
		WHERE
			type ='table'
			AND name NOT LIKE 'sqlite_%'
			AND name NOT LIKE '%_attachments'
			AND name NOT LIKE '%_revs'
			AND name NOT LIKE '%_design'
			AND name NOT LIKE '%_attachments_bridge'
		`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var dbs []string
	for rows.Next() {
		var db string
		if err := rows.Scan(&db); err != nil {
			return nil, err
		}
		dbs = append(dbs, db)
	}
	return dbs, rows.Err()
}
