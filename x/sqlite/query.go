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
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dop251/goja"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal"
)

func fromJSValue(v interface{}) (*string, error) {
	if v == nil {
		return nil, nil
	}
	vv, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	s := string(vv)
	return &s, nil
}

func (d *db) Query(ctx context.Context, ddoc, view string, options driver.Options) (driver.Rows, error) {
	opts := newOpts(options)
	update, err := opts.update()
	if err != nil {
		return nil, err
	}
	// Normalize the ddoc and view values
	ddoc = strings.TrimPrefix(ddoc, "_design/")
	view = strings.TrimPrefix(view, "_view/")

	if update == updateModeTrue {
		if err := d.updateIndex(ctx, ddoc, view); err != nil {
			return nil, err
		}
	}

	query := d.ddocQuery(ddoc, view, `
		SELECT
			id,
			key,
			value,
			"" AS rev,
			NULL AS doc,
			"" AS conflicts
		FROM {{ .Map }}
		ORDER BY key
	`)
	results, err := d.db.QueryContext(ctx, query) //nolint:rowserrcheck // Err checked in Next
	if err != nil {
		return nil, err
	}

	if update == updateModeLazy {
		go func() {
			if err := d.updateIndex(context.Background(), ddoc, view); err != nil {
				d.logger.Print("Failed to update index: " + err.Error())
			}
		}()
	}

	return &rows{
		ctx:  ctx,
		db:   d,
		rows: results,
	}, nil
}

func (d *db) updateIndex(ctx context.Context, ddoc, view string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rev, err := d.winningRev(ctx, tx, "_design/"+ddoc)
	if err != nil {
		return err
	}

	query := d.query(`
		SELECT func_body
		FROM {{ .Design }}
		WHERE id = $1
			AND rev = $2
			AND rev_id = $3
			AND func_type = 'map'
			AND func_name = $4
	`)

	var funcBody string
	err = tx.QueryRowContext(ctx, query, "_design/"+ddoc, rev.rev, rev.id, view).Scan(&funcBody)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &internal.Error{Status: http.StatusNotFound, Message: "missing named view"}
		}

		return err
	}

	insert, err := tx.PrepareContext(ctx, d.ddocQuery(ddoc, view, `
		INSERT INTO {{ .Map }} (id, key, value)
		VAlUES ($1, $2, $3)		
	`))
	if err != nil {
		return err
	}

	query = d.query(`
		WITH RankedRevisions AS (
			SELECT
				id,
				rev,
				rev_id,
				doc
			FROM {{ .Leaves }} AS rev
			WHERE NOT deleted
		)
		SELECT
			rev.id                       AS id,
			rev.rev || '-' || rev.rev_id AS rev,
			rev.doc                      AS doc
		FROM RankedRevisions AS rev
		WHERE rev.id NOT LIKE '_local/%'
		GROUP BY rev.id, rev.rev, rev.rev_id
	`)
	docs, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer docs.Close()

	emit := func(id string) func(interface{}, interface{}) {
		return func(key, value interface{}) {
			k, err := fromJSValue(key)
			if err != nil {
				panic(err)
			}
			v, err := fromJSValue(value)
			if err != nil {
				panic(err)
			}
			if _, err := insert.ExecContext(ctx, id, k, v); err != nil {
				panic(err)
			}
		}
	}

	vm := goja.New()
	if _, err := vm.RunString("const map = " + funcBody); err != nil {
		return err
	}

	mf, ok := goja.AssertFunction(vm.Get("map"))
	if !ok {
		return fmt.Errorf("expected map to be a function, got %T", vm.Get("map"))
	}

	for docs.Next() {
		var (
			id, rev string
			doc     json.RawMessage
		)
		if err := docs.Scan(&id, &rev, &doc); err != nil {
			return err
		}
		full := &fullDoc{
			ID:  id,
			Rev: rev,
			Doc: doc,
		}
		if err := vm.Set("emit", emit(id)); err != nil {
			return err
		}
		if _, err := mf(goja.Undefined(), vm.ToValue(full.toMap())); err != nil {
			return err
		}
	}
	if err := docs.Err(); err != nil {
		return err
	}

	return tx.Commit()
}
