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

// Stats returns database statistics.
func (d *db) Stats(ctx context.Context) (*driver.DBStats, error) {
	var docCount, deletedCount int64
	err := d.db.QueryRowContext(ctx, d.query(`
		WITH leaves AS (
			SELECT
				rev.id,
				doc.deleted,
				ROW_NUMBER() OVER (PARTITION BY rev.id ORDER BY rev.rev DESC, rev.rev_id DESC) AS rank
			FROM {{ .Revs }} AS rev
			LEFT JOIN {{ .Revs }} AS child
				ON child.id = rev.id
				AND rev.rev = child.parent_rev
				AND rev.rev_id = child.parent_rev_id
			JOIN {{ .Docs }} AS doc
				ON rev.id = doc.id
				AND rev.rev = doc.rev
				AND rev.rev_id = doc.rev_id
			WHERE child.id IS NULL
		)
		SELECT
			COALESCE(SUM(CASE WHEN NOT deleted THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN deleted THEN 1 ELSE 0 END), 0)
		FROM leaves
		WHERE rank = 1
	`)).Scan(&docCount, &deletedCount)
	if err != nil {
		return nil, d.errDatabaseNotFound(err)
	}

	lastSeq, err := d.lastSeq(ctx)
	if err != nil {
		return nil, err
	}

	return &driver.DBStats{
		Name:         d.name,
		DocCount:     docCount,
		DeletedCount: deletedCount,
		UpdateSeq:    fmt.Sprintf("%d", lastSeq),
	}, nil
}

func (d *db) Compact(ctx context.Context) error {
	_, err := d.db.ExecContext(ctx, "VACUUM")
	return err
}

func (db) CompactView(context.Context, string) error { return nil }

func (db) ViewCleanup(context.Context) error { return nil }

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
