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
	"io"
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

	var results *sql.Rows
	for {
		rev, err := d.updateIndex(ctx, ddoc, view, update)
		if err != nil {
			return nil, err
		}

		query := d.ddocQuery(ddoc, view, rev.String(), `
			SELECT
				COALESCE(MAX(last_seq), 0) == (SELECT COALESCE(max(seq),0) FROM {{ .Docs }}) AS up_to_date,
				NULL,
				NULL,
				NULL,
				NULL,
				NULL
			FROM {{ .Design }}
			WHERE id = $1
				AND rev = $2
				AND rev_id = $3
				AND func_type = 'map'
				AND func_name = $4

			UNION

			SELECT *
			FROM (
				SELECT
					id,
					key,
					value,
					"" AS rev,
					NULL AS doc,
					"" AS conflicts
				FROM {{ .Map }}
				ORDER BY key
			)
		`)
		results, err = d.db.QueryContext(ctx, query, "_design/"+ddoc, rev.rev, rev.id, view) //nolint:rowserrcheck // Err checked in Next
		switch {
		case errIsNoSuchTable(err):
			return nil, &internal.Error{Status: http.StatusNotFound, Message: "missing named view"}
		case err != nil:
			return nil, err
		}

		// The first row is used to verify the index is up to date
		if !results.Next() {
			// should never happen
			return nil, errors.New("no rows returned")
		}
		if update != updateModeTrue {
			break
		}
		var upToDate bool
		if err := results.Scan(&upToDate, discard{}, discard{}, discard{}, discard{}, discard{}); err != nil {
			return nil, err
		}
		if upToDate {
			break
		}
	}

	if update == updateModeLazy {
		go func() {
			if _, err := d.updateIndex(context.Background(), ddoc, view, updateModeTrue); err != nil {
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

const batchSize = 100

// updateIndex queries for the current index status, and returns the current
// ddoc revid and last_seq. If mode is "true", it will also update the index.
func (d *db) updateIndex(ctx context.Context, ddoc, view, mode string) (revision, error) {
	var (
		ddocRev  revision
		funcBody *string
		lastSeq  int
	)
	err := d.db.QueryRowContext(ctx, d.query(`
		SELECT
			docs.rev,
			docs.rev_id,
			design.func_body,
			COALESCE(design.last_seq, 0) AS last_seq
		FROM {{ .Docs }} AS docs
		LEFT JOIN {{ .Design }} AS design ON docs.id = design.id AND docs.rev = design.rev AND docs.rev_id = design.rev_id
		WHERE docs.id = $1
		ORDER BY docs.rev DESC, docs.rev_id DESC
		LIMIT 1
	`), "_design/"+ddoc).Scan(&ddocRev.rev, &ddocRev.id, &funcBody, &lastSeq)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return revision{}, &internal.Error{Status: http.StatusNotFound, Message: "missing"}
	case err != nil:
		return revision{}, err
	}

	if mode != "true" {
		return ddocRev, nil
	}

	if funcBody == nil {
		return revision{}, &internal.Error{Status: http.StatusNotFound, Message: "missing named view"}
	}

	query := d.query(`
		SELECT
			doc.seq                      AS seq,
			rev.id                       AS id,
			rev.rev || '-' || rev.rev_id AS rev,
			doc.doc                      AS doc,
			doc.deleted                  AS deleted
		FROM {{ .Revs }} AS rev
		LEFT JOIN {{ .Revs }} AS child ON rev.id = child.id AND rev.rev = child.parent_rev AND rev.rev_id = child.parent_rev_id
		JOIN {{ .Docs }} AS doc ON rev.id = doc.id AND rev.rev = doc.rev AND rev.rev_id = doc.rev_id
		WHERE rev.id NOT LIKE '_local/%'
			AND child.id IS NULL
			AND doc.seq > $1
		ORDER BY doc.seq
	`)
	docs, err := d.db.QueryContext(ctx, query, lastSeq)
	if err != nil {
		return revision{}, err
	}
	defer docs.Close()

	batch := newMapIndexBatch()

	vm := goja.New()

	emit := func(id string) func(interface{}, interface{}) {
		return func(key, value interface{}) {
			defer func() {
				if r := recover(); r != nil {
					panic(vm.ToValue(r))
				}
			}()
			k, err := fromJSValue(key)
			if err != nil {
				panic(err)
			}
			v, err := fromJSValue(value)
			if err != nil {
				panic(err)
			}
			batch.add(id, k, v)
		}
	}

	if _, err := vm.RunString("const map = " + *funcBody); err != nil {
		return revision{}, err
	}

	mf, ok := goja.AssertFunction(vm.Get("map"))
	if !ok {
		return revision{}, fmt.Errorf("expected map to be a function, got %T", vm.Get("map"))
	}

	var seq int
	for {
		full := &fullDoc{}
		err := iter(docs, &seq, full)
		if err == io.EOF {
			break
		}
		if err != nil {
			return revision{}, err
		}

		if full.Deleted {
			batch.delete(full.ID)
			continue
		}

		if err := vm.Set("emit", emit(full.ID)); err != nil {
			return revision{}, err
		}
		if _, err := mf(goja.Undefined(), vm.ToValue(full.toMap())); err != nil {
			var exception *goja.Exception
			if errors.As(err, &exception) {
				d.logger.Printf("map function threw exception for %s: %s", full.ID, exception.String())
				batch.delete(full.ID)
			} else {
				return revision{}, err
			}
		}
		if batch.insertCount >= batchSize {
			if err := d.writeMapIndexBatch(ctx, seq, ddocRev, ddoc, view, batch); err != nil {
				return revision{}, err
			}
			batch.clear()
		}
	}

	if err := d.writeMapIndexBatch(ctx, seq, ddocRev, ddoc, view, batch); err != nil {
		return revision{}, err
	}

	return ddocRev, docs.Err()
}

func iter(docs *sql.Rows, seq *int, full *fullDoc) error {
	if !docs.Next() {
		if err := docs.Err(); err != nil {
			return err
		}
		return io.EOF
	}
	if err := docs.Scan(seq, &full.ID, &full.Rev, &full.Doc, &full.Deleted); err != nil {
		return err
	}
	return nil
}

type mapIndexBatch struct {
	insertCount int
	entries     map[string][]mapIndexEntry
	deleted     []string
}

type mapIndexEntry struct {
	Key   *string
	Value *string
}

func newMapIndexBatch() *mapIndexBatch {
	return &mapIndexBatch{
		entries: make(map[string][]mapIndexEntry, batchSize),
	}
}

func (b *mapIndexBatch) add(id string, key, value *string) {
	b.entries[id] = append(b.entries[id], mapIndexEntry{
		Key:   key,
		Value: value,
	})
	b.insertCount++
}

func (b *mapIndexBatch) delete(id string) {
	b.deleted = append(b.deleted, id)
	b.insertCount -= len(b.entries[id])
	delete(b.entries, id)
}

func (b *mapIndexBatch) clear() {
	b.insertCount = 0
	b.deleted = b.deleted[:0]
	b.entries = make(map[string][]mapIndexEntry, batchSize)
}

func (d *db) writeMapIndexBatch(ctx context.Context, seq int, rev revision, ddoc, viewName string, batch *mapIndexBatch) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, d.query(`
		UPDATE {{ .Design }}
		SET last_seq=$1
		WHERE id = $2
			AND rev = $3
			AND rev_id = $4
			AND func_type = 'map'
			AND func_name = $5
	`), seq, "_design/"+ddoc, rev.rev, rev.id, viewName); err != nil {
		return err
	}

	// Clear any stale entries
	ids := make([]interface{}, 0, len(batch.entries)+len(batch.deleted))
	for id := range batch.entries {
		ids = append(ids, id)
	}
	for _, id := range batch.deleted {
		ids = append(ids, id)
	}
	query := fmt.Sprintf(d.ddocQuery(ddoc, viewName, rev.String(), `
		DELETE FROM {{ .Map }}
		WHERE id IN (%s)		
	`), placeholders(1, len(ids)))
	if _, err := tx.ExecContext(ctx, query, ids...); err != nil {
		return err
	}

	if batch.insertCount == 0 {
		return tx.Commit()
	}

	args := make([]interface{}, 0, batch.insertCount*3)
	values := make([]string, 0, batch.insertCount)
	for id, entries := range batch.entries {
		for _, entry := range entries {
			values = append(values, fmt.Sprintf("($%d, $%d, $%d)", len(args)+1, len(args)+2, len(args)+3))
			args = append(args, id, entry.Key, entry.Value)
		}
	}
	query = d.ddocQuery(ddoc, viewName, rev.String(), `
		INSERT INTO {{ .Map }} (id, key, value)
		VALUES
	`) + strings.Join(values, ",")
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return tx.Commit()
}
