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
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/x/options"
)

func (c *client) AllDBs(ctx context.Context, opts driver.Options) ([]string, error) {
	o, err := options.New(opts).PaginationOptions(true)
	if err != nil {
		return nil, err
	}

	args := []any{tablePrefix}
	where := append([]string{"TRUE"}, o.BuildWhere(&args)...)

	query := fmt.Sprintf(`
		WITH view AS (
			SELECT substr(tablename, length($1)+1) AS key
			FROM pg_tables
			WHERE tablename LIKE $1 || '%%'
		)
		SELECT key
		FROM view
		WHERE %s
		%s
	`, strings.Join(where, " AND "), o.BuildOrderBy())

	rows, err := c.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[string])
}
