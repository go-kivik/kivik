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

package pg

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/x/options"
)

func (c *client) AllDBs(ctx context.Context, opts driver.Options) ([]string, error) {
	o := options.New(opts)
	descending, err := o.Descending()
	if err != nil {
		return nil, err
	}
	order := "ASC"
	if descending {
		order = "DESC"
	}

	args := []any{tablePrefix}

	endWhere := "true"
	endkey, err := o.EndKey()
	if err != nil {
		return nil, err
	}
	if endkey != "" {
		var endKeyString string
		if err := json.Unmarshal([]byte(endkey), &endKeyString); err != nil {
			return nil, fmt.Errorf("endkey must be a JSON string")
		}
		endWhere = "tablename <= $2"
		args = append(args, tablePrefix+endKeyString)
	}

	query := fmt.Sprintf(`
		SELECT substr(tablename, length($1)+1)
		FROM pg_tables
		WHERE tablename LIKE $1 || '%%'
			AND %s
		ORDER BY tablename %s
	`, endWhere, order)

	rows, err := c.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[string])
}
