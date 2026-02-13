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
	"strconv"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/x/options"
)

type dbUpdates struct {
	rows *sql.Rows
}

func (u *dbUpdates) Next(update *driver.DBUpdate) error {
	if !u.rows.Next() {
		if err := u.rows.Err(); err != nil {
			return err
		}
		return io.EOF
	}

	var seq int64
	if err := u.rows.Scan(&seq, &update.DBName, &update.Type); err != nil {
		return err
	}
	update.Seq = strconv.FormatInt(seq, 10)
	return nil
}

func (u *dbUpdates) Close() error {
	return u.rows.Close()
}

func (c *client) DBUpdates(ctx context.Context, opts driver.Options) (driver.DBUpdates, error) {
	optMap := options.New(opts)

	var queryArgs []interface{}
	query := c.query(`
		SELECT seq, db_name, type
		FROM {{ .DBUpdatesLog }}
	`)

	if _, sinceVal, ok := optMap.Get("since"); ok {
		query += ` WHERE seq > ?`
		queryArgs = append(queryArgs, sinceVal)
	}

	query += ` ORDER BY seq`

	//nolint:rowserrcheck // Err checked in Next
	rows, err := c.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, err
	}

	return &dbUpdates{rows: rows}, nil
}

func (c *client) ensureDBUpdatesLog(ctx context.Context) error {
	_, err := c.db.ExecContext(ctx, c.query(`
		CREATE TABLE IF NOT EXISTS {{ .DBUpdatesLog }} (
			seq INTEGER PRIMARY KEY AUTOINCREMENT,
			db_name TEXT NOT NULL,
			type TEXT NOT NULL
		)
	`))
	return err
}

func (c *client) logDBUpdate(ctx context.Context, tx *sql.Tx, dbName, eventType string) error {
	_, err := tx.ExecContext(ctx, c.query(`
		INSERT INTO {{ .DBUpdatesLog }} (db_name, type)
		VALUES (?, ?)
	`), dbName, eventType)
	return err
}
