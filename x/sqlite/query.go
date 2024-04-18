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
	"github.com/go-kivik/kivik/x/sqlite/v4/internal"
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

	rev, err := d.winningRev(ctx, d.db, "_design/"+ddoc)
	if err != nil {
		return nil, err
	}

	query := d.ddocQuery(ddoc, view, rev.String(), `
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
	var (
		rev      revision
		funcBody *string
	)
	err := d.db.QueryRowContext(ctx, d.query(`
		SELECT docs.rev, docs.rev_id, design.func_body
		FROM {{ .Docs }} AS docs
		LEFT JOIN {{ .Design }} AS design ON docs.id = design.id AND docs.rev = design.rev AND docs.rev_id = design.rev_id
		WHERE docs.id = $1
		ORDER BY docs.rev DESC, docs.rev_id DESC
		LIMIT 1
	`), "_design/"+ddoc).Scan(&rev.rev, &rev.id, &funcBody)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &internal.Error{Status: http.StatusNotFound, Message: "missing"}
		}
		return err
	}

	if funcBody == nil {
		return &internal.Error{Status: http.StatusNotFound, Message: "missing named view"}
	}

	insert, err := d.db.PrepareContext(ctx, d.ddocQuery(ddoc, view, rev.String(), `
		INSERT INTO {{ .Map }} (id, key, value)
		VAlUES ($1, $2, $3)		
	`))
	if err != nil {
		return err
	}

	query := d.query(`
		SELECT
			rev.id                       AS id,
			rev.rev || '-' || rev.rev_id AS rev,
			doc.doc                      AS doc
		FROM {{ .Revs }} AS rev
		LEFT JOIN {{ .Revs }} AS child ON rev.id = child.id AND rev.rev = child.parent_rev AND rev.rev_id = child.parent_rev_id
		JOIN {{ .Docs }} AS doc ON rev.id = doc.id AND rev.rev = doc.rev AND rev.rev_id = doc.rev_id
		WHERE rev.id NOT LIKE '_local/%'
			AND child.id IS NULL
			AND NOT doc.deleted
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
	if _, err := vm.RunString("const map = " + *funcBody); err != nil {
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

	return docs.Err()
}
