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
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/x/options"
)

type queryRunner interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (c *client) globalChangesExists(ctx context.Context, qr queryRunner) (bool, error) {
	var exists bool
	if err := qr.QueryRowContext(ctx, `
		SELECT COUNT(*) > 0 FROM sqlite_master
		WHERE type = 'table' AND name = 'kivik$_global_changes'
	`).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (c *client) DBUpdates(ctx context.Context, opts driver.Options) (driver.DBUpdates, error) {
	exists, err := c.globalChangesExists(ctx, c.db)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &internal.Error{Status: http.StatusServiceUnavailable, Message: "Service Unavailable"}
	}

	o := options.New(opts)
	feed, err := o.Feed()
	if err != nil {
		return nil, err
	}

	changeOpts := map[string]any{"include_docs": true}
	if _, feedRaw, ok := o.Get("feed"); ok {
		changeOpts["feed"] = feedRaw
	}
	if _, sinceRaw, ok := o.Get("since"); ok {
		changeOpts["since"] = sinceRaw
	} else if feed == options.FeedLongpoll || feed == options.FeedContinuous {
		changeOpts["since"] = "now"
	}

	globalDB := c.newDB("_global_changes")
	ch, err := globalDB.Changes(ctx, kivik.Params(changeOpts))
	if err != nil {
		return nil, err
	}
	return &globalChangesDBUpdates{changes: ch}, nil
}

type globalChangesDBUpdates struct {
	changes driver.Changes
}

func (u *globalChangesDBUpdates) Next(update *driver.DBUpdate) error {
	var change driver.Change
	if err := u.changes.Next(&change); err != nil {
		return err
	}
	var doc struct {
		DBName string `json:"db_name"`
		Type   string `json:"type"`
	}
	if err := json.Unmarshal(change.Doc, &doc); err != nil {
		return err
	}
	update.DBName = doc.DBName
	update.Type = doc.Type
	update.Seq = change.Seq
	return nil
}

func (u *globalChangesDBUpdates) Close() error { return u.changes.Close() }

func (u *globalChangesDBUpdates) LastSeq() (string, error) {
	return u.changes.LastSeq(), nil
}

func (c *client) logGlobalChange(ctx context.Context, tx *sql.Tx, dbName, eventType string) error {
	if dbName == "_global_changes" {
		return nil
	}

	exists, err := c.globalChangesExists(ctx, tx)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	docID := uuid.NewString()
	data, err := prepareDoc(docID, map[string]any{"db_name": dbName, "type": eventType})
	if err != nil {
		return err
	}

	d := c.newDB("_global_changes")
	rev := revision{rev: 1, id: data.RevID()}

	if _, err := tx.ExecContext(ctx, d.query(`
		INSERT INTO {{ .Revs }} (id, rev, rev_id)
		VALUES ($1, 1, $2)
	`), data.ID, rev.id); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, d.query(`
		INSERT INTO {{ .Docs }} (id, rev, rev_id, doc, md5sum, deleted)
		VALUES ($1, 1, $2, $3, $4, $5)
	`), data.ID, rev.id, data.Doc, data.MD5sum, data.Deleted); err != nil {
		return err
	}

	return nil
}
