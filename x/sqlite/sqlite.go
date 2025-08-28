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
	"fmt"
	"log"
	"net/http"
	"regexp"

	"modernc.org/sqlite"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
)

func init() {
	kivik.Register("sqlite", &drv{})

	if err := sqlite.RegisterCollationUtf8("COUCHDB_UCI", couchdbCmpString); err != nil {
		panic(err)
	}
}

type drv struct{}

var _ driver.Driver = (*drv)(nil)

// NewClient returns a new SQLite client. dsn should be the full path to your
// SQLite database file.
func (drv) NewClient(dsn string, options driver.Options) (driver.Client, error) {
	cn := &connector{dsn: dsn}
	options.Apply(cn)
	db, err := cn.Connect()
	if err != nil {
		return nil, err
	}

	c := &client{
		dsn:    dsn,
		db:     db,
		logger: log.Default(),
	}
	options.Apply(c)

	return c, nil
}

type client struct {
	dsn    string
	db     *sql.DB
	logger *log.Logger
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

var validDBNameRE = regexp.MustCompile(`^[a-z][a-z0-9_$()+/-]*$`)

func validateDBName(name string) error {
	switch name {
	case "_users", "_replicator", "_global_changes":
		return nil
	default:
		if !validDBNameRE.MatchString(name) {
			return &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid database name: %s", name)}
		}
	}
	return nil
}

func (c *client) DestroyDB(ctx context.Context, name string, _ driver.Options) error {
	if err := validateDBName(name); err != nil {
		return err
	}
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	d := c.newDB(name)
	rows, err := tx.QueryContext(ctx, d.query(`
		SELECT
			id,
			rev,
			rev_id,
			func_name
		FROM {{ .Design }}
		WHERE func_type = 'map'
	`))
	if err != nil {
		if errIsNoSuchTable(err) {
			return &internal.Error{Status: http.StatusNotFound, Message: "database not found"}
		}
		return err
	}

	defer rows.Close()
	for rows.Next() {
		var (
			id, view string
			rev      revision
		)
		if err := rows.Scan(&id, &rev.rev, &rev.id, &view); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, d.ddocQuery(id, view, rev.String(), `DROP TABLE {{ .Map }}`))
		if err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, query := range destroySchema {
		_, err := tx.ExecContext(ctx, d.query(query))
		if err != nil {
			if errIsNoSuchTable(err) {
				return &internal.Error{Status: http.StatusNotFound, Message: "database not found"}
			}
			return err
		}
	}
	return tx.Commit()
}

func (c *client) DB(name string, _ driver.Options) (driver.DB, error) {
	if err := validateDBName(name); err != nil {
		return nil, err
	}
	return c.newDB(name), nil
}
