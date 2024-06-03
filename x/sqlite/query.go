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

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/x/sqlite/v4/reduce"
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
	vopts, err := opts.viewOptions(ddoc)
	if err != nil {
		return nil, err
	}

	if isBuiltinView(ddoc) {
		return d.queryBuiltinView(ctx, vopts)
	}

	// Normalize the ddoc and view values
	ddoc = strings.TrimPrefix(ddoc, "_design/")
	view = strings.TrimPrefix(view, "_view/")

	results, err := d.performQuery(
		ctx,
		ddoc, view,
		vopts,
	)
	if err != nil {
		return nil, err
	}

	if vopts.update == updateModeLazy {
		go func() {
			if _, err := d.updateIndex(context.Background(), ddoc, view, updateModeTrue); err != nil {
				d.logger.Print("Failed to update index: " + err.Error())
			}
		}()
	}

	return results, nil
}

const (
	leavesCTE = `
	WITH leaves AS (
		SELECT
			rev.id,
			rev.rev,
			rev.rev_id,
			rev.key,
			doc.doc,
			doc.deleted
		FROM {{ .Revs }} AS rev
		LEFT JOIN {{ .Revs }} AS child ON child.id = rev.id AND rev.rev = child.parent_rev AND rev.rev_id = child.parent_rev_id
		JOIN {{ .Docs }} AS doc ON rev.id = doc.id AND rev.rev = doc.rev AND rev.rev_id = doc.rev_id
		WHERE child.id IS NULL
			AND NOT doc.deleted
	)
`
)

func (d *db) performQuery(
	ctx context.Context,
	ddoc, view string,
	vopts *viewOptions,
) (driver.Rows, error) {
	if vopts.group {
		return d.performGroupQuery(ctx, ddoc, view, vopts)
	}
	for {
		rev, err := d.updateIndex(ctx, ddoc, view, vopts.update)
		if err != nil {
			return nil, err
		}

		args := []interface{}{
			vopts.includeDocs, vopts.conflicts, vopts.reduce, vopts.updateSeq,
			"_design/" + ddoc, rev.rev, rev.id, view, vopts.attachments,
		}

		where := append([]string{""}, vopts.buildWhere(&args)...)

		query := fmt.Sprintf(d.ddocQuery(ddoc, view, rev.String(), leavesCTE+`,
			 reduce AS (
				SELECT
					CASE WHEN MAX(id) IS NOT NULL THEN TRUE ELSE FALSE END AS reducible,
					COALESCE(func_body, "")                                AS reduce_func
				FROM {{ .Design }}
				WHERE id = $5
					AND rev = $6
					AND rev_id = $7
					AND func_type = 'reduce'
					AND func_name = $8
			)

			SELECT
				COALESCE(MAX(last_seq), 0) == (SELECT COALESCE(max(seq),0) FROM {{ .Docs }}) AS up_to_date,
				reduce.reducible,
				reduce.reduce_func,
				IIF($4, last_seq, "") AS update_seq,
				NULL,
				NULL,
				0    AS attachment_count,
				NULL AS filename,
				NULL AS content_type,
				NULL AS length,
				NULL AS digest,
				NULL AS rev_pos,
				NULL AS data
			FROM {{ .Design }} AS map
			JOIN reduce
			WHERE id = $5
				AND rev = $6
				AND rev_id = $7
				AND func_type = 'map'
				AND func_name = $8

			UNION ALL

			SELECT *
			FROM (
				SELECT
					id    AS id,
					key   AS key,
					value AS value,
					NULL  AS rev,
					NULL  AS doc,
					NULL  AS conflicts,
					0     AS attachment_count,
					NULL AS filename,
					NULL AS content_type,
					NULL AS length,
					NULL AS digest,
					NULL AS rev_pos,
					NULL AS data
				FROM {{ .Map }} AS view
				JOIN reduce
				WHERE reduce.reducible AND ($3 IS NULL OR $3 == TRUE)
				ORDER BY key
			)

			UNION ALL

			SELECT
				CASE WHEN row_number = 1 THEN id        END AS id,
				CASE WHEN row_number = 1 THEN key       END AS key,
				CASE WHEN row_number = 1 THEN value     END AS value,
				CASE WHEN row_number = 1 THEN rev       END AS rev,
				CASE WHEN row_number = 1 THEN doc       END AS doc,
				CASE WHEN row_number = 1 THEN conflicts END AS conflicts,
				COALESCE(attachment_count, 0) AS attachment_count,
				filename,
				content_type,
				length,
				digest,
				rev_pos,
				data
			FROM (
				SELECT
					view.id,
					view.key,
					view.value,
					IIF($1, view.rev || '-' || view.rev_id, "") AS rev,
					view.doc,
					view.conflicts,
					SUM(CASE WHEN bridge.pk IS NOT NULL THEN 1 ELSE 0 END) OVER (PARTITION BY view.id, view.rev, view.rev_id) AS attachment_count,
					ROW_NUMBER() OVER (PARTITION BY view.id, view.rev, view.rev_id, view.pk) AS row_number,
					att.filename AS filename,
					att.content_type AS content_type,
					att.length AS length,
					att.digest AS digest,
					att.rev_pos AS rev_pos,
					IIF($9, att.data, NULL) AS data
				FROM (
					SELECT
						view.pk,
						view.id,
						view.key,
						view.value,
						docs.rev,
						docs.rev_id,
						IIF($1, docs.doc, NULL) AS doc,
						IIF($2, GROUP_CONCAT(conflicts.rev || '-' || conflicts.rev_id, ','), NULL) AS conflicts
					FROM {{ .Map }} AS view
					JOIN reduce
					JOIN {{ .Docs }} AS docs ON view.id = docs.id AND view.rev = docs.rev AND view.rev_id = docs.rev_id
					LEFT JOIN leaves AS conflicts ON conflicts.id = view.id AND NOT (view.rev = conflicts.rev AND view.rev_id = conflicts.rev_id)
					WHERE $3 == FALSE OR NOT reduce.reducible
						%[2]s -- WHERE
					GROUP BY view.id, view.key, view.value, view.rev, view.rev_id
					%[1]s -- ORDER BY
					LIMIT %[3]d OFFSET %[4]d
				) AS view
				LEFT JOIN {{ .AttachmentsBridge }} AS bridge ON view.id = bridge.id AND view.rev = bridge.rev AND view.rev_id = bridge.rev_id AND $1
				LEFT JOIN {{ .Attachments }} AS att ON bridge.pk = att.pk
				%[1]s -- ORDER BY
			)
		`), vopts.buildOrderBy(), strings.Join(where, " AND "), vopts.limit, vopts.skip)
		results, err := d.db.QueryContext(ctx, query, args...) //nolint:rowserrcheck // Err checked in Next
		switch {
		case errIsNoSuchTable(err):
			return nil, &internal.Error{Status: http.StatusNotFound, Message: "missing named view"}
		case err != nil:
			return nil, err
		}

		meta, err := readFirstRow(results, vopts)
		if err != nil {
			return nil, err
		}

		if !meta.upToDate && vopts.update == updateModeTrue {
			_ = results.Close() //nolint:sqlclosecheck // Not up to date, so close the results and try again
			continue
		}

		if meta.reducible && (vopts.reduce == nil || *vopts.reduce) {
			return d.reduce(results, meta.reduceFuncJS, vopts.reduceGroupLevel())
		}

		// If the results are up to date, OR, we're in false/lazy update mode,
		// then these results are fine.
		return &rows{
			ctx:       ctx,
			db:        d,
			rows:      results,
			updateSeq: meta.updateSeq,
		}, nil
	}
}

func (d *db) performGroupQuery(ctx context.Context, ddoc, view string, vopts *viewOptions) (driver.Rows, error) {
	var (
		results      *sql.Rows
		reducible    bool
		reduceFuncJS string
	)
	for {
		rev, err := d.updateIndex(ctx, ddoc, view, vopts.update)
		if err != nil {
			return nil, err
		}

		query := fmt.Sprintf(d.ddocQuery(ddoc, view, rev.String(), `
			WITH reduce AS (
				SELECT
					CASE WHEN MAX(id) IS NOT NULL THEN TRUE ELSE FALSE END AS reducible,
					COALESCE(func_body, "")                                AS reduce_func
				FROM {{ .Design }}
				WHERE id = $1
					AND rev = $2
					AND rev_id = $3
					AND func_type = 'reduce'
					AND func_name = $4
			)

			SELECT
				COALESCE(MAX(last_seq), 0) == (SELECT COALESCE(max(seq),0) FROM {{ .Docs }}) AS up_to_date,
				reduce.reducible,
				reduce.reduce_func,
				NULL,
				NULL,
				NULL,
				0    AS attachment_count,
				NULL AS filename,
				NULL AS content_type,
				NULL AS length,
				NULL AS digest,
				NULL AS rev_pos,
				NULL AS data
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
					NULL  AS conflicts,
					0    AS attachment_count,
					NULL AS filename,
					NULL AS content_type,
					NULL AS length,
					NULL AS digest,
					NULL AS rev_pos,
					NULL AS data
				FROM {{ .Map }} AS view
				JOIN reduce
				WHERE reduce.reducible AND ($6 IS NULL OR $6 == TRUE)
				%[1]s -- ORDER BY
			)
		`), vopts.buildOrderBy())

		results, err = d.db.QueryContext( //nolint:rowserrcheck // Err checked in iterator
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
		if vopts.update != updateModeTrue {
			break
		}
		var upToDate bool
		if err := results.Scan(
			&upToDate, &reducible, &reduceFuncJS, discard{}, discard{}, discard{},
			discard{}, discard{}, discard{}, discard{}, discard{}, discard{}, discard{},
		); err != nil {
			return nil, err
		}
		if !reducible {
			field := "group"
			if vopts.groupLevel > 0 {
				field = "group_level"
			}
			return nil, &internal.Error{Status: http.StatusBadRequest, Message: field + " is invalid for map-only views"}
		}
		if upToDate {
			break
		}
	}

	return d.reduce(results, reduceFuncJS, vopts.reduceGroupLevel())
}

func (d *db) reduce(results *sql.Rows, reduceFuncJS string, groupLevel int) (driver.Rows, error) {
	callback := func([]reduce.Row) {
		// fmt.Println(rows)
	}
	return reduce.Reduce(&reduceRowIter{results: results}, reduceFuncJS, d.logger, groupLevel, callback)
}

const batchSize = 100

// updateIndex queries for the current index status, and returns the current
// ddoc revid and last_seq. If mode is "true", it will also update the index.
func (d *db) updateIndex(ctx context.Context, ddoc, view, mode string) (revision, error) {
	var (
		ddocRev                 revision
		mapFuncJS               *string
		lastSeq                 int
		includeDesign, localSeq sql.NullBool
	)
	err := d.db.QueryRowContext(ctx, d.query(`
		SELECT
			docs.rev,
			docs.rev_id,
			design.func_body,
			design.include_design,
			design.local_seq,
			COALESCE(design.last_seq, 0) AS last_seq
		FROM {{ .Docs }} AS docs
		LEFT JOIN {{ .Design }} AS design ON docs.id = design.id AND docs.rev = design.rev AND docs.rev_id = design.rev_id AND design.func_type = 'map'
		WHERE docs.id = $1
		ORDER BY docs.rev DESC, docs.rev_id DESC
		LIMIT 1
	`), "_design/"+ddoc).Scan(&ddocRev.rev, &ddocRev.id, &mapFuncJS, &includeDesign, &localSeq, &lastSeq)
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
		WITH leaves AS (
			SELECT
				rev.id                    AS id,
				rev.rev                   AS rev,
				rev.rev_id                AS rev_id,
				doc.doc,
				doc.deleted
			FROM {{ .Revs }} AS rev
			LEFT JOIN {{ .Revs }} AS child ON child.id = rev.id AND rev.rev = child.parent_rev AND rev.rev_id = child.parent_rev_id
			JOIN {{ .Docs }} AS doc ON rev.id = doc.id AND rev.rev = doc.rev AND rev.rev_id = doc.rev_id
			WHERE child.id IS NULL
		)
		SELECT
			CASE WHEN row_number = 1 THEN seq     END AS seq,
			CASE WHEN row_number = 1 THEN id      END AS id,
			CASE WHEN row_number = 1 THEN rev     END AS rev,
			CASE WHEN row_number = 1 THEN doc     END AS doc,
			CASE WHEN row_number = 1 THEN deleted END AS deleted,
			COALESCE(attachment_count, 0)             AS attachment_count,
			filename,
			content_type,
			length,
			digest,
			rev_pos
		FROM (
			SELECT
				seq.seq                      AS seq,
				doc.id                       AS id,
				doc.rev || '-' || doc.rev_id AS rev,
				seq.doc                      AS doc,
				seq.deleted                  AS deleted,
				doc.attachment_count,
				doc.row_number,
				doc.filename,
				doc.content_type,
				doc.length,
				doc.digest,
				doc.rev_pos
			FROM {{ .Docs }} AS seq
			LEFT JOIN (
				SELECT
					rev.id,
					rev.rev,
					rev.rev_id,
					SUM(CASE WHEN bridge.pk IS NOT NULL THEN 1 ELSE 0 END) OVER (PARTITION BY rev.id, rev.rev, rev.rev_id) AS attachment_count,
					ROW_NUMBER() OVER (PARTITION BY rev.id, rev.rev, rev.rev_id) AS row_number,
					att.filename,
					att.content_type,
					att.length,
					att.digest,
					att.rev_pos
				FROM (
					SELECT
						id                    AS id,
						rev                   AS rev,
						rev_id                AS rev_id,
						IIF($1, doc, NULL)    AS doc,
						ROW_NUMBER() OVER (PARTITION BY id ORDER BY rev DESC, rev_id DESC) AS rank
					FROM leaves
				) AS rev
				LEFT JOIN {{ .AttachmentsBridge }} AS bridge ON rev.id = bridge.id AND rev.rev = bridge.rev AND rev.rev_id = bridge.rev_id
				LEFT JOIN {{ .Attachments }} AS att ON bridge.pk = att.pk
				WHERE rev.rank = 1
			) AS doc ON seq.id = doc.id AND seq.rev = doc.rev AND seq.rev_id = doc.rev_id
			WHERE seq.seq > $1
			ORDER BY seq.seq
		)
	`)
	docs, err := d.db.QueryContext(ctx, query, lastSeq)
	if err != nil {
		return revision{}, err
	}
	defer docs.Close()

	batch := newMapIndexBatch()

	vm := goja.New()

	emit := func(id string, rev revision) func(interface{}, interface{}) {
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
			batch.add(id, rev, k, v)
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

		// Skip design/local docs
		if full.ID == "" {
			continue
		}

		rev, err := full.rev()
		if err != nil {
			return revision{}, err
		}

		// TODO move this to the query
		if strings.HasPrefix(full.ID, "_local/") ||
			(!includeDesign.Bool && strings.HasPrefix(full.ID, "_design/")) {
			continue
		}

		if full.Deleted {
			batch.delete(full.ID, rev)
			continue
		}

		if localSeq.Bool {
			full.LocalSeq = seq
		}

		if err := vm.Set("emit", emit(full.ID, rev)); err != nil {
			return revision{}, err
		}
		if _, err := mapFunc(goja.Undefined(), vm.ToValue(full.toMap())); err != nil {
			var exception *goja.Exception
			if errors.As(err, &exception) {
				d.logger.Printf("map function threw exception for %s: %s", full.ID, exception.String())
				batch.delete(full.ID, rev)
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

type docRev struct {
	id    string
	rev   int
	revID string
}

type mapIndexBatch struct {
	insertCount int
	entries     map[docRev][]mapIndexEntry
	deleted     []docRev
}

type mapIndexEntry struct {
	Key   *string
	Value *string
}

func newMapIndexBatch() *mapIndexBatch {
	return &mapIndexBatch{
		entries: make(map[docRev][]mapIndexEntry, batchSize),
	}
}

func (b *mapIndexBatch) add(id string, rev revision, key, value *string) {
	mapKey := docRev{
		id:    id,
		rev:   rev.rev,
		revID: rev.id,
	}
	b.entries[mapKey] = append(b.entries[mapKey], mapIndexEntry{
		Key:   key,
		Value: value,
	})
	b.insertCount++
}

func (b *mapIndexBatch) delete(id string, rev revision) {
	mapKey := docRev{
		id:    id,
		rev:   rev.rev,
		revID: rev.id,
	}
	b.deleted = append(b.deleted, mapKey)
	b.insertCount -= len(b.entries[mapKey])
	delete(b.entries, mapKey)
}

func (b *mapIndexBatch) clear() {
	b.insertCount = 0
	b.deleted = b.deleted[:0]
	b.entries = make(map[docRev][]mapIndexEntry, batchSize)
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
		for mapKey := range batch.entries {
			ids = append(ids, mapKey.id)
		}
		for _, mapKey := range batch.deleted {
			ids = append(ids, mapKey.id)
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
		args := make([]interface{}, 0, batch.insertCount*5)
		values := make([]string, 0, batch.insertCount)
		for mapKey, entries := range batch.entries {
			for _, entry := range entries {
				values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", len(args)+1, len(args)+2, len(args)+3, len(args)+4, len(args)+5))
				args = append(args, mapKey.id, mapKey.rev, mapKey.revID, entry.Key, entry.Value)
			}
		}
		query := d.ddocQuery(ddoc, viewName, rev.String(), `
		INSERT INTO {{ .Map }} (id, rev, rev_id, key, value)
		VALUES
	`) + strings.Join(values, ",")
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}

	return tx.Commit()
}
