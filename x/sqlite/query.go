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
	"errors"
	"net/http"
	"strings"

	"modernc.org/sqlite"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal"
)

func (d *db) Query(ctx context.Context, ddoc, view string, _ driver.Options) (driver.Rows, error) {
	// Normalize the ddoc and view values
	ddoc = strings.TrimPrefix(ddoc, "_design/")
	view = strings.TrimPrefix(view, "_view/")

	tx, err := d.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rev, err := d.winningRev(ctx, tx, "_design/"+ddoc)
	if err != nil {
		return nil, err
	}

	var funcBody string
	err = tx.QueryRowContext(ctx, d.ddocQuery(ddoc, view, `
		SELECT func_body
		FROM {{ .Map }}
		WHERE id = $1
			AND rev = $2
			AND rev_id = $3
			AND func_type = 'map'
			AND func_name = $4
	`), "_design/"+ddoc, rev.rev, rev.id, view).Scan(&funcBody)
	if err != nil {
		sqliteErr := new(sqlite.Error)
		if errors.As(err, &sqliteErr) &&
			sqliteErr.Code() == codeSQLiteError &&
			strings.Contains(sqliteErr.Error(), "no such table") {
			return nil, &internal.Error{Status: http.StatusNotFound, Message: "missing named view"}
		}

		return nil, err
	}

	return nil, tx.Commit()
}
