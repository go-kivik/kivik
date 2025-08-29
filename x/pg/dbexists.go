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
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/go-kivik/kivik/v4/driver"
)

func (c *client) DBExists(ctx context.Context, dbName string, _ driver.Options) (bool, error) {
	var exists bool
	err := c.pool.QueryRow(ctx, "SELECT true FROM pg_tables WHERE tablename = $1", tablePrefix+dbName).Scan(&exists)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return false, nil
	case err != nil:
		return false, err
	}
	return exists, nil
}
