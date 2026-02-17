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
	"time"

	"github.com/cenkalti/backoff/v4"

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

	return u.rows.Scan(&update.Seq, &update.DBName, &update.Type)
}

func (u *dbUpdates) Close() error {
	return u.rows.Close()
}

func (c *client) DBUpdates(ctx context.Context, opts driver.Options) (driver.DBUpdates, error) {
	optMap := options.New(opts)

	feed, err := optMap.Feed()
	if err != nil {
		return nil, err
	}

	if feed == "longpoll" || feed == "continuous" {
		return c.newLongpollDBUpdates(ctx, optMap, feed == "continuous")
	}

	sinceNow, sinceSeq, _, err := optMap.Since()
	if err != nil {
		return nil, err
	}

	var since uint64
	if sinceNow {
		row := c.db.QueryRowContext(ctx, c.query(`
			SELECT COALESCE(MAX(seq), 0)
			FROM {{ .DBUpdatesLog }}
		`))
		if err := row.Scan(&since); err != nil {
			return nil, err
		}
	} else {
		since = sinceSeq
	}

	var queryArgs []interface{}
	query := c.query(`
		SELECT seq, db_name, type
		FROM {{ .DBUpdatesLog }}
	`)

	if sinceNow || sinceSeq > 0 {
		query += ` WHERE seq > ?`
		queryArgs = append(queryArgs, since)
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

type longpollDBUpdates struct {
	stmt        *sql.Stmt
	since       uint64
	continuous  bool
	ctx         context.Context
	done        bool
	currentRows *sql.Rows
	backoff     *backoff.ExponentialBackOff
}

var _ driver.DBUpdates = (*longpollDBUpdates)(nil)

func (c *client) newLongpollDBUpdates(ctx context.Context, optMap options.Map, continuous bool) (*longpollDBUpdates, error) {
	sinceNow, sinceSeq, _, err := optMap.Since()
	if err != nil {
		return nil, err
	}

	var since uint64
	if sinceNow {
		row := c.db.QueryRowContext(ctx, c.query(`
			SELECT COALESCE(MAX(seq), 0)
			FROM {{ .DBUpdatesLog }}
		`))
		if err := row.Scan(&since); err != nil {
			return nil, err
		}
	} else {
		since = sinceSeq
	}

	stmt, err := c.db.PrepareContext(ctx, c.query(`
		SELECT seq, db_name, type
		FROM {{ .DBUpdatesLog }}
		WHERE seq > ?
		ORDER BY seq
	`))
	if err != nil {
		return nil, err
	}

	idleTimeout, err := optMap.Timeout()
	if err != nil {
		return nil, err
	}

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 100 * time.Millisecond
	bo.MaxInterval = 10 * time.Second
	bo.MaxElapsedTime = idleTimeout

	return &longpollDBUpdates{
		stmt:       stmt,
		since:      since,
		continuous: continuous,
		ctx:        ctx,
		backoff:    bo,
	}, nil
}

func (u *longpollDBUpdates) Next(update *driver.DBUpdate) error {
	if u.done {
		return io.EOF
	}

	for {
		if u.currentRows != nil {
			if u.currentRows.Next() {
				if err := u.currentRows.Scan(&update.Seq, &update.DBName, &update.Type); err != nil {
					return err
				}
				seq, _ := strconv.ParseUint(update.Seq, 10, 64)
				u.since = seq
				return nil
			}

			if err := u.currentRows.Err(); err != nil {
				return err
			}
			if err := u.currentRows.Close(); err != nil {
				return err
			}
			u.currentRows = nil

			if !u.continuous {
				u.done = true
				return io.EOF
			}

			u.backoff.Reset()
		}

		rows, err := u.stmt.QueryContext(u.ctx, u.since)
		if err != nil {
			return err
		}

		if rows.Next() {
			if err := rows.Scan(&update.Seq, &update.DBName, &update.Type); err != nil {
				_ = rows.Close() //nolint:sqlclosecheck
				return err
			}
			seq, _ := strconv.ParseUint(update.Seq, 10, 64)
			u.since = seq
			u.currentRows = rows
			return nil
		}

		if err := rows.Err(); err != nil {
			_ = rows.Close() //nolint:sqlclosecheck
			return err
		}
		if err := rows.Close(); err != nil {
			return err
		}

		next := u.backoff.NextBackOff()
		if next == backoff.Stop {
			if !u.continuous {
				return io.EOF
			}
			u.backoff.Reset()
			next = u.backoff.NextBackOff()
		}

		select {
		case <-time.After(next):
		case <-u.ctx.Done():
			return u.ctx.Err()
		}
	}
}

func (u *longpollDBUpdates) Close() error {
	if u.currentRows != nil {
		if err := u.currentRows.Close(); err != nil {
			_ = u.stmt.Close()
			return err
		}
	}
	return u.stmt.Close()
}
