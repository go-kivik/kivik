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

	"github.com/go-kivik/kivik/v4/driver"
)

func (c *client) DBExists(ctx context.Context, name string, _ driver.Options) (bool, error) {
	var exists bool
	err := c.db.QueryRowContext(ctx, `
		SELECT
			TRUE
		FROM
			sqlite_schema
		WHERE
			type = 'table' AND
			name = ?
		`, tablePrefix+name).Scan(&exists)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	return exists, nil
}
