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
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/dop251/goja"

	"github.com/go-kivik/kivik/v4"
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

	reduce, err := opts.reduce()
	if err != nil {
		return nil, err
	}
	group, err := opts.group()
	if err != nil {
		return nil, err
	}
	groupLevel, err := opts.groupLevel()
	if err != nil {
		return nil, err
	}

	results, err := d.performQuery(ctx, ddoc, view, update, reduce, group, groupLevel)
	if err != nil {
		return nil, err
	}

	if update == updateModeLazy {
		go func() {
			if _, err := d.updateIndex(context.Background(), ddoc, view, updateModeTrue); err != nil {
				d.logger.Print("Failed to update index: " + err.Error())
			}
		}()
	}

	return results, nil
}

func (d *db) performQuery(ctx context.Context, ddoc, view, update string, reduce *bool, group bool, groupLevel uint64) (driver.Rows, error) {
	if group {
		return d.performGroupQuery(ctx, ddoc, view, update, groupLevel)
	}
	var (
		results      *sql.Rows
		reducible    bool
		reduceFuncJS *string
	)
	for {
		rev, err := d.updateIndex(ctx, ddoc, view, update)
		if err != nil {
			return nil, err
		}

		query := d.ddocQuery(ddoc, view, rev.String(), `
			WITH reduce AS (
				SELECT
					CASE WHEN MAX(id) IS NOT NULL THEN TRUE ELSE FALSE END AS reducable,
					func_body                                              AS reduce_func
				FROM {{ .Design }}
				WHERE id = $1
					AND rev = $2
					AND rev_id = $3
					AND func_type = 'reduce'
					AND func_name = $4
			)

			SELECT
				COALESCE(MAX(last_seq), 0) == (SELECT COALESCE(max(seq),0) FROM {{ .Docs }}) AS up_to_date,
				reduce.reducable,
				reduce.reduce_func,
				NULL,
				NULL,
				NULL
			FROM {{ .Design }} AS map
			JOIN reduce
			WHERE id = $1
				AND rev = $2
				AND rev_id = $3
				AND func_type = 'map'
				AND func_name = $4

			UNION ALL

			SELECT *
			FROM (
				SELECT
					id    AS id,
					key   AS key,
					value AS value,
					NULL  AS rev,
					NULL  AS doc,
					NULL  AS conflicts
				FROM {{ .Map }}
				JOIN reduce
				WHERE reduce.reducable AND ($6 IS NULL OR $6 == TRUE)
				ORDER BY id, key
			)

			UNION ALL

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
				JOIN reduce
				WHERE $6 == FALSE OR NOT reduce.reducable
				ORDER BY key
			)
		`)

		results, err = d.db.QueryContext( //nolint:rowserrcheck // Err checked in Next
			ctx, query,
			"_design/"+ddoc, rev.rev, rev.id, view, kivik.EndKeySuffix, reduce,
		)
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
		if err := results.Scan(&upToDate, &reducible, &reduceFuncJS, discard{}, discard{}, discard{}); err != nil {
			return nil, err
		}
		if reduce != nil && *reduce && !reducible {
			return nil, &internal.Error{Status: http.StatusBadRequest, Message: "reduce is invalid for map-only views"}
		}
		if upToDate {
			break
		}
	}

	if reducible && (reduce == nil || *reduce) {
		return d.reduceRows(results, reduceFuncJS, false, 0)
	}

	return &rows{
		ctx:  ctx,
		db:   d,
		rows: results,
	}, nil
}

func (d *db) performGroupQuery(ctx context.Context, ddoc, view, update string, groupLevel uint64) (driver.Rows, error) {
	var (
		results      *sql.Rows
		reducible    bool
		reduceFuncJS *string
	)
	for {
		rev, err := d.updateIndex(ctx, ddoc, view, update)
		if err != nil {
			return nil, err
		}

		query := d.ddocQuery(ddoc, view, rev.String(), `
			WITH reduce AS (
				SELECT
					CASE WHEN MAX(id) IS NOT NULL THEN TRUE ELSE FALSE END AS reducable,
					func_body                                              AS reduce_func
				FROM {{ .Design }}
				WHERE id = $1
					AND rev = $2
					AND rev_id = $3
					AND func_type = 'reduce'
					AND func_name = $4
			)

			SELECT
				COALESCE(MAX(last_seq), 0) == (SELECT COALESCE(max(seq),0) FROM {{ .Docs }}) AS up_to_date,
				reduce.reducable,
				reduce.reduce_func,
				NULL,
				NULL,
				NULL
			FROM {{ .Design }} AS map
			JOIN reduce
			WHERE id = $1
				AND rev = $2
				AND rev_id = $3
				AND func_type = 'map'
				AND func_name = $4

			UNION ALL

			SELECT *
			FROM (
				SELECT
					id    AS id,
					COALESCE(key, "null") AS key,
					value AS value,
					NULL  AS rev,
					NULL  AS doc,
					NULL  AS conflicts
				FROM {{ .Map }}
				JOIN reduce
				WHERE reduce.reducable AND ($6 IS NULL OR $6 == TRUE)
				ORDER BY id, key
			)
		`)

		results, err = d.db.QueryContext(
			ctx, query,
			"_design/"+ddoc, rev.rev, rev.id, view, kivik.EndKeySuffix, true,
		)
		switch {
		case errIsNoSuchTable(err):
			return nil, &internal.Error{Status: http.StatusNotFound, Message: "missing named view"}
		case err != nil:
			return nil, err
		}
		defer results.Close()

		// The first row is used to verify the index is up to date
		if !results.Next() {
			// should never happen
			return nil, errors.New("no rows returned")
		}
		if update != updateModeTrue {
			break
		}
		var upToDate bool
		if err := results.Scan(&upToDate, &reducible, &reduceFuncJS, discard{}, discard{}, discard{}); err != nil {
			return nil, err
		}
		if !reducible {
			field := "group"
			if groupLevel > 0 {
				field = "group_level"
			}
			return nil, &internal.Error{Status: http.StatusBadRequest, Message: field + " is invalid for map-only views"}
		}
		if upToDate {
			break
		}
	}

	return d.reduceRows(results, reduceFuncJS, true, groupLevel)
}

const batchSize = 100

// updateIndex queries for the current index status, and returns the current
// ddoc revid and last_seq. If mode is "true", it will also update the index.
func (d *db) updateIndex(ctx context.Context, ddoc, view, mode string) (revision, error) {
	var (
		ddocRev   revision
		mapFuncJS *string
		lastSeq   int
	)
	err := d.db.QueryRowContext(ctx, d.query(`
		SELECT
			docs.rev,
			docs.rev_id,
			design.func_body,
			COALESCE(design.last_seq, 0) AS last_seq
		FROM {{ .Docs }} AS docs
		LEFT JOIN {{ .Design }} AS design ON docs.id = design.id AND docs.rev = design.rev AND docs.rev_id = design.rev_id AND design.func_type = 'map'
		WHERE docs.id = $1
		ORDER BY docs.rev DESC, docs.rev_id DESC
		LIMIT 1
	`), "_design/"+ddoc).Scan(&ddocRev.rev, &ddocRev.id, &mapFuncJS, &lastSeq)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return revision{}, &internal.Error{Status: http.StatusNotFound, Message: "missing"}
	case err != nil:
		return revision{}, err
	}

	if mode != "true" {
		return ddocRev, nil
	}

	if mapFuncJS == nil {
		return revision{}, &internal.Error{Status: http.StatusNotFound, Message: "missing named view"}
	}

	query := d.query(`
		SELECT
			CASE WHEN row_number = 1 THEN seq     END AS seq,
			CASE WHEN row_number = 1 THEN id      END AS id,
			CASE WHEN row_number = 1 THEN rev     END AS rev,
			CASE WHEN row_number = 1 THEN doc     END AS doc,
			CASE WHEN row_number = 1 THEN deleted END AS deleted,
			attachment_count,
			filename,
			content_type,
			length,
			digest,
			rev_pos
		FROM (
			SELECT
				doc.seq                      AS seq,
				rev.id                       AS id,
				rev.rev || '-' || rev.rev_id AS rev,
				doc.doc                      AS doc,
				doc.deleted                  AS deleted,
				SUM(CASE WHEN bridge.pk IS NOT NULL THEN 1 ELSE 0 END) OVER (PARTITION BY rev.id, rev.rev, rev.rev_id) AS attachment_count,
				ROW_NUMBER() OVER (PARTITION BY rev.id, rev.rev, rev.rev_id) AS row_number,
				att.filename,
				att.content_type,
				att.length,
				att.digest,
				att.rev_pos
			FROM {{ .Revs }} AS rev
			LEFT JOIN {{ .Revs }} AS child ON rev.id = child.id AND rev.rev = child.parent_rev AND rev.rev_id = child.parent_rev_id
			JOIN {{ .Docs }} AS doc ON rev.id = doc.id AND rev.rev = doc.rev AND rev.rev_id = doc.rev_id
			LEFT JOIN {{ .AttachmentsBridge }} AS bridge ON doc.id = bridge.id AND doc.rev = bridge.rev AND doc.rev_id = bridge.rev_id
			LEFT JOIN {{ .Attachments }} AS att ON bridge.pk = att.pk
			WHERE rev.id NOT LIKE '_local/%'
				AND child.id IS NULL
				AND doc.seq > $1
			ORDER BY doc.seq
		)
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

	if _, err := vm.RunString("const map = " + *mapFuncJS); err != nil {
		return revision{}, err
	}

	mapFunc, ok := goja.AssertFunction(vm.Get("map"))
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
		if _, err := mapFunc(goja.Undefined(), vm.ToValue(full.toMap())); err != nil {
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
	var (
		attachmentsCount int
		rowSeq           *int
		rowID, rowRev    *string
		rowDoc           *[]byte
		rowDeleted       *bool
	)
	for {
		if !docs.Next() {
			if err := docs.Err(); err != nil {
				return err
			}
			return io.EOF
		}
		var (
			filename, contentType *string
			length                *int64
			revPos                *int
			digest                *md5sum
		)
		if err := docs.Scan(
			&rowSeq, &rowID, &rowRev, &rowDoc, &rowDeleted,
			&attachmentsCount,
			&filename, &contentType, &length, &digest, &revPos,
		); err != nil {
			return err
		}
		if rowSeq != nil {
			*seq = *rowSeq
			*full = fullDoc{
				ID:      *rowID,
				Rev:     *rowRev,
				Doc:     *rowDoc,
				Deleted: *rowDeleted,
			}
		}
		if filename != nil {
			if full.Attachments == nil {
				full.Attachments = make(map[string]*attachment)
			}
			full.Attachments[*filename] = &attachment{
				ContentType: *contentType,
				Length:      *length,
				Digest:      *digest,
				RevPos:      *revPos,
			}
		}
		if attachmentsCount == len(full.Attachments) {
			break
		}
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
	if len(batch.entries) > 0 || len(batch.deleted) > 0 {
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
	}

	if batch.insertCount > 0 {
		args := make([]interface{}, 0, batch.insertCount*3)
		values := make([]string, 0, batch.insertCount)
		for id, entries := range batch.entries {
			for _, entry := range entries {
				values = append(values, fmt.Sprintf("($%d, $%d, $%d)", len(args)+1, len(args)+2, len(args)+3))
				args = append(args, id, entry.Key, entry.Value)
			}
		}
		query := d.ddocQuery(ddoc, viewName, rev.String(), `
		INSERT INTO {{ .Map }} (id, key, value)
		VALUES
	`) + strings.Join(values, ",")
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}

	return tx.Commit()
}

type reduceFunc func(keys [][2]interface{}, values []interface{}, rereduce bool) (interface{}, error)

func (d *db) reduceFunc(reduceFuncJS *string, logger *log.Logger) (reduceFunc, error) {
	if reduceFuncJS == nil {
		return nil, nil
	}
	switch *reduceFuncJS {
	case "_count":
		return func(_ [][2]interface{}, values []interface{}, rereduce bool) (interface{}, error) {
			if !rereduce {
				return len(values), nil
			}
			var total uint64
			for _, value := range values {
				v, _ := toUint64(value, "")
				total += v
			}
			return total, nil
		}, nil
	case "_sum":
		return func(_ [][2]interface{}, values []interface{}, _ bool) (interface{}, error) {
			var total uint64
			for _, value := range values {
				v, _ := toUint64(value, "")
				total += v
			}
			return total, nil
		}, nil
	case "_stats":
		return func(_ [][2]interface{}, values []interface{}, rereduce bool) (interface{}, error) {
			type stats struct {
				Sum    float64 `json:"sum"`
				Min    float64 `json:"min"`
				Max    float64 `json:"max"`
				Count  float64 `json:"count"`
				SumSqr float64 `json:"sumsqr"`
			}
			var result stats
			if rereduce {
				mins := make([]float64, 0, len(values))
				maxs := make([]float64, 0, len(values))
				for _, v := range values {
					value := v.(stats)
					mins = append(mins, value.Min)
					maxs = append(maxs, value.Max)
					result.Sum += value.Sum
					result.Count += value.Count
					result.SumSqr += value.SumSqr
				}
				result.Min = slices.Min(mins)
				result.Max = slices.Max(maxs)
				return result, nil
			}
			result.Count = float64(len(values))
			nvals := make([]float64, 0, len(values))
			for _, v := range values {
				value, ok := toFloat64(v)
				if !ok {
					return nil, &internal.Error{
						Status:  http.StatusInternalServerError,
						Message: fmt.Sprintf("the _stats function requires that map values be numbers or arrays of numbers, not '%s'", v),
					}
				}
				nvals = append(nvals, value)
				result.Sum += value
				result.SumSqr += value * value
			}
			result.Min = slices.Min(nvals)
			result.Max = slices.Max(nvals)
			return result, nil
		}, nil
	default:
		vm := goja.New()

		if _, err := vm.RunString("const reduce = " + *reduceFuncJS); err != nil {
			return nil, err
		}
		reduceFunc, ok := goja.AssertFunction(vm.Get("reduce"))
		if !ok {
			return nil, fmt.Errorf("expected reduce to be a function, got %T", vm.Get("map"))
		}

		return func(keys [][2]interface{}, values []interface{}, rereduce bool) (interface{}, error) {
			reduceValue, err := reduceFunc(goja.Undefined(), vm.ToValue(keys), vm.ToValue(values), vm.ToValue(rereduce))
			// According to CouchDB reference implementation, when a user-defined
			// reduce function throws an exception, the error is logged and the
			// return value is set to null.
			if err != nil {
				logger.Printf("reduce function threw exception: %s", err.Error())
				return nil, nil
			}

			return reduceValue.Export(), nil
		}, nil
	}
}
