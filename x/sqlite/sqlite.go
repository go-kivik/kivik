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
	"net/http"
	"regexp"
	"strings"

	"modernc.org/sqlite"
	_ "modernc.org/sqlite" // SQLite driver

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal"
)

type drv struct{}

var _ driver.Driver = (*drv)(nil)

// NewClient returns a new SQLite client. dsn should be the full path to your
// SQLite database file.
func (drv) NewClient(dns string, _ driver.Options) (driver.Client, error) {
	db, err := sql.Open("sqlite", dns)
	if err != nil {
		return nil, err
	}
	return &client{
		db: db,
	}, nil
}

type client struct {
	db *sql.DB
}

var _ driver.Client = (*client)(nil)

const (
	version = "0.0.1"
	vendor  = "Kivik"
)

func (client) Version(context.Context) (*driver.Version, error) {
	return &driver.Version{
		Version: version,
		Vendor:  vendor,
	}, nil
}

func (c *client) AllDBs(ctx context.Context, _ driver.Options) ([]string, error) {
	rows, err := c.db.QueryContext(ctx, `
		SELECT
			name
		FROM
			sqlite_schema
		WHERE
			type ='table' AND
			name NOT LIKE 'sqlite_%'
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
		`, name).Scan(&exists)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	return exists, nil
}

var validDBNameRE = regexp.MustCompile(`^[a-z][a-z0-9_$()+/-]*$`)

const (
	// https://www.sqlite.org/rescode.html
	codeSQLiteError = 1
)

func (c *client) CreateDB(ctx context.Context, name string, _ driver.Options) error {
	if !validDBNameRE.MatchString(name) {
		return &internal.Error{Status: http.StatusBadRequest, Message: "invalid database name"}
	}
	_, err := c.db.ExecContext(ctx, `CREATE TABLE "`+name+`" (id INTEGER)`)
	if err == nil {
		return nil
	}
	sqliteErr := new(sqlite.Error)
	if errors.As(err, &sqliteErr) &&
		sqliteErr.Code() == codeSQLiteError &&
		strings.Contains(sqliteErr.Error(), "already exists") {
		return &internal.Error{Status: http.StatusPreconditionFailed, Message: "database already exists"}
	}
	return err
}

func (client) DestroyDB(context.Context, string, driver.Options) error {
	return nil
}

func (client) DB(string, driver.Options) (driver.DB, error) {
	return nil, nil
}
