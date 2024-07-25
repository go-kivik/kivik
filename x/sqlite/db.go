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
	"fmt"
	"log"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
)

type db struct {
	db     *sql.DB
	name   string
	logger *log.Logger
}

var (
	_ driver.DB     = (*db)(nil)
	_ driver.Finder = (*db)(nil)
)

func (c *client) newDB(name string) *db {
	return &db{
		db:     c.db,
		name:   name,
		logger: c.logger,
	}
}

func (d *db) Close() error {
	return d.db.Close()
}

// TODO: I think Ping belongs on *client, not *db
func (d *db) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

/* -- stub methods -- */

func (db) Stats(context.Context) (*driver.DBStats, error) {
	return nil, errors.New("not implemented")
}

func (db) Compact(context.Context) error {
	return errors.New("not implemented")
}

func (db) CompactView(context.Context, string) error {
	return errors.New("not implemented")
}

func (db) ViewCleanup(context.Context) error {
	return errors.New("not implemented")
}

func (db) BulkDocs(context.Context, []interface{}, driver.Options) ([]driver.BulkResult, error) {
	return nil, errors.New("not implemented")
}

func (db) Copy(context.Context, string, string, driver.Options) (string, error) {
	return "", errors.New("not implemented")
}

func (db) CreateIndex(context.Context, string, string, interface{}, driver.Options) error {
	return errors.New("not implemented")
}

func (db) GetIndexes(context.Context, driver.Options) ([]driver.Index, error) {
	return nil, errors.New("not implemented")
}

func (db) DeleteIndex(context.Context, string, string, driver.Options) error {
	return errors.New("not implemented")
}

func (db) Explain(context.Context, interface{}, driver.Options) (*driver.QueryPlan, error) {
	return nil, errors.New("not implemented")
}

// errDatabaseNotFound converts a sqlite "no such table"  error into a kivik
// database not found error
func (d *db) errDatabaseNotFound(err error) error {
	if err == nil {
		return nil
	}
	if errIsNoSuchTable(err) {
		return &internal.Error{
			Status:  http.StatusNotFound,
			Message: fmt.Sprintf("database not found: %s", d.name),
		}
	}
	return err
}
