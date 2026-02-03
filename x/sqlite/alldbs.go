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
	"strings"

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
			AND name LIKE 'kivik$%' ESCAPE '\'
			AND name NOT LIKE '%$attachments' ESCAPE '\'
			AND name NOT LIKE '%$revs' ESCAPE '\'
			AND name NOT LIKE '%$design' ESCAPE '\'
			AND name NOT LIKE '%$attachments_bridge' ESCAPE '\'
		`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var dbs []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		dbs = append(dbs, strings.TrimPrefix(name, tablePrefix))
	}
	return dbs, rows.Err()
}
