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
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/x/mango"
)

func endKeyOp(descending, inclusive bool) string {
	switch {
	case descending && inclusive:
		return ">="
	case descending && !inclusive:
		return ">"
	case !descending && inclusive:
		return "<="
	case !descending && !inclusive:
		return "<"
	}
	panic("unreachable")
}

func startKeyOp(descending bool) string {
	if descending {
		return "<="
	}
	return ">="
}

const (
	viewAllDocs    = "_all_docs"
	viewLocalDocs  = "_local_docs"
	viewDesignDocs = "_design_docs"
)

func isBuiltinView(view string) bool {
	switch view {
	case viewAllDocs, viewLocalDocs, viewDesignDocs:
		return true
	}
	return false
}

func (d *db) AllDocs(ctx context.Context, options driver.Options) (driver.Rows, error) {
	return d.Query(ctx, viewAllDocs, "", options)
}

func (d *db) LocalDocs(ctx context.Context, options driver.Options) (driver.Rows, error) {
	return d.Query(ctx, viewLocalDocs, "", options)
}

func (d *db) DesignDocs(ctx context.Context, options driver.Options) (driver.Rows, error) {
	return d.Query(ctx, viewDesignDocs, "", options)
}

func (d *db) queryBuiltinView(
	ctx context.Context,
	vopts *viewOptions,
) (driver.Rows, error) {
	args := []interface{}{vopts.includeDocs, vopts.conflicts, vopts.updateSeq, vopts.attachments, vopts.bookmark}

	where := append([]string{""}, vopts.buildWhere(&args)...)

	query := fmt.Sprintf(d.query(leavesCTE+`,
		main AS (
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
				data,
				doc_number
			FROM (
				SELECT
					view.id,
					view.key,
					'{"value":{"rev":"' || view.rev || '-' || view.rev_id || '"}}' AS value,
					view.rev || '-' || view.rev_id AS rev,
					view.doc,
					view.conflicts,
					SUM(CASE WHEN bridge.pk IS NOT NULL THEN 1 ELSE 0 END) OVER (PARTITION BY view.id, view.rev, view.rev_id) AS attachment_count,
					ROW_NUMBER() OVER (PARTITION BY view.id, view.rev, view.rev_id) AS row_number,
					att.filename AS filename,
					att.content_type AS content_type,
					att.length AS length,
					att.digest AS digest,
					att.rev_pos AS rev_pos,
					IIF($4, att.data, NULL) AS data,
					ROW_NUMBER() OVER (%[1]s) AS doc_number
				FROM (
					SELECT
						view.id     AS id,
						view.key    AS key,
						view.rev    AS rev,
						view.rev_id AS rev_id,
						view.doc    AS doc,
						IIF($2, GROUP_CONCAT(conflicts.rev || '-' || conflicts.rev_id, ','), NULL) AS conflicts
					FROM (
						SELECT
							id                    AS id,
							rev                   AS rev,
							rev_id                AS rev_id,
							key                   AS key,
							IIF($1, doc, NULL)    AS doc,
							ROW_NUMBER() OVER (PARTITION BY id ORDER BY rev DESC, rev_id DESC) AS rank
						FROM leaves
					) AS view
					LEFT JOIN leaves AS conflicts ON conflicts.id = view.id AND NOT (view.rev = conflicts.rev AND view.rev_id = conflicts.rev_id)
					WHERE view.rank = 1
						%[2]s -- WHERE
					GROUP BY view.id, view.rev, view.rev_id
					%[1]s -- ORDER BY
					LIMIT %[3]d OFFSET %[4]d
				) AS view
				LEFT JOIN {{ .AttachmentsBridge }} AS bridge ON view.id = bridge.id AND view.rev = bridge.rev AND view.rev_id = bridge.rev_id AND $1
				LEFT JOIN {{ .Attachments }} AS att ON bridge.pk = att.pk
				%[1]s -- ORDER BY
			)
		),
		bookmark AS (
			SELECT doc_number
			FROM main
			WHERE id = $5
		)

		SELECT
			TRUE                  AS up_to_date,
			FALSE                 AS reducible,
			""                    AS reduce_func,
			IIF($3, MAX(seq), "") AS update_seq,
			NULL,
			NULL,
			NULL AS attachment_count,
			NULL AS filename,
			NULL AS content_type,
			NULL AS length,
			NULL AS digest,
			NULL AS rev_pos,
			NULL AS data
		FROM {{ .Docs }}

		UNION ALL

		SELECT
			id,
			key,
			value,
			rev,
			doc,
			conflicts,
			attachment_count,
			filename,
			content_type,
			length,
			digest,
			rev_pos,
			data
		FROM main
		%[5]s -- bookmark filtering
	`), vopts.buildOrderBy(), strings.Join(where, " AND "), vopts.limit, vopts.skip, vopts.bookmarkWhere())
	results, err := d.db.QueryContext(ctx, query, args...) //nolint:rowserrcheck // Err checked in Next
	if err != nil {
		return nil, err
	}

	meta, err := readFirstRow(results, vopts)
	if err != nil {
		return nil, err
	}

	return &rows{
		ctx:       ctx,
		db:        d,
		rows:      results,
		updateSeq: meta.updateSeq,
		selector:  vopts.selector,
		findLimit: vopts.findLimit,
		findSkip:  vopts.findSkip,
		fields:    vopts.fields,
	}, nil
}

type viewMetadata struct {
	upToDate     bool
	reducible    bool
	reduceFuncJS string
	updateSeq    string
	lastSeq      int
}

// readFirstRow reads the first row from the resultset, which contains. In the
// case of an error, the result set is closed and an error is returned.
func readFirstRow(results *sql.Rows, vopts *viewOptions) (*viewMetadata, error) {
	if !results.Next() {
		// should never happen
		_ = results.Close() //nolint:sqlclosecheck // Aborting
		return nil, errors.New("no rows returned")
	}
	var meta viewMetadata
	var lastSeq *int
	if err := results.Scan(
		&meta.upToDate, &meta.reducible, &meta.reduceFuncJS, &meta.updateSeq, &lastSeq, discard{},
		discard{}, discard{}, discard{}, discard{}, discard{}, discard{}, discard{},
	); err != nil {
		_ = results.Close() //nolint:sqlclosecheck // Aborting
		return nil, err
	}
	if vopts.reduce != nil && *vopts.reduce && !meta.reducible {
		_ = results.Close() //nolint:sqlclosecheck // Aborting
		opt := "reduce"
		switch {
		case vopts.groupLevel > 0:
			opt = "group_level"
		case vopts.group:
			opt = "group"
		}
		return nil, &internal.Error{Status: http.StatusBadRequest, Message: opt + " is invalid for map-only views"}
	}
	if lastSeq != nil {
		meta.lastSeq = *lastSeq
	}
	return &meta, nil
}

func descendingToDirection(descending bool) string {
	if descending {
		return "DESC"
	}
	return "ASC"
}

type rows struct {
	ctx                 context.Context
	db                  *db
	rows                *sql.Rows
	updateSeq           string
	selector            *mango.Selector
	findLimit, findSkip int64
	index               int64
	fields              []string
	bookmark            string
}

var _ driver.Rows = (*rows)(nil)

func (r *rows) Next(row *driver.Row) error {
	var (
		attachmentCount int
		full            *fullDoc
	)
	for {
		if !r.rows.Next() {
			if err := r.rows.Err(); err != nil {
				return err
			}
			return io.EOF
		}
		var (
			key, doc                                     []byte
			value, data                                  *[]byte
			id, conflicts, rowRev, filename, contentType *string
			length                                       *int64
			revPos                                       *int
			digest                                       *md5sum
		)
		if err := r.rows.Scan(
			&id, &key, &value, &rowRev, &doc, &conflicts,
			&attachmentCount,
			&filename, &contentType, &length, &digest, &revPos, &data,
		); err != nil {
			return err
		}
		if rowRev != nil {
			// If rowRev is populated, it means we're on the first row for the
			// document. Otherwise, we're on an attachment-only row.
			if id != nil {
				row.ID = *id
			}
			row.Key = key
			if len(key) == 0 {
				row.Key = []byte("null")
			}
			if value == nil {
				row.Value = strings.NewReader("null")
			} else {
				row.Value = bytes.NewReader(*value)
			}
			if doc != nil {
				full = &fullDoc{
					ID:  row.ID,
					Rev: *rowRev,
					Doc: doc,
				}
				if conflicts != nil {
					full.Conflicts = strings.Split(*conflicts, ",")
				}
			}
		}
		if filename != nil {
			if full.Attachments == nil {
				full.Attachments = make(map[string]*attachment)
			}
			var jsonData json.RawMessage
			if data != nil {
				var err error
				jsonData, err = json.Marshal(*data)
				if err != nil {
					return err
				}
			}
			full.Attachments[*filename] = &attachment{
				ContentType: *contentType,
				Length:      *length,
				Digest:      *digest,
				RevPos:      *revPos,
				Data:        jsonData,
			}
		}
		if attachmentCount == 0 || attachmentCount == len(full.Attachments) {
			break
		}
	}
	if full != nil {
		if r.selector != nil {
			// This means we're responding to a _find query, which requires
			// filtering the results, and a different format.
			if !r.selector.Match(full.toMap()) {
				return r.Next(row)
			}
			r.index++
			if r.index <= r.findSkip {
				return r.Next(row)
			}
			if r.findLimit > 0 && r.index > r.findLimit+r.findSkip {
				return io.EOF
			}
			// These values are omitted from the _find response
			r.bookmark = row.ID
			row.ID = ""
			row.Key = nil
			row.Value = nil
		}
		row.Doc = full.toReader(r.fields...)
	}
	return nil
}

func (r *rows) Close() error {
	return r.rows.Close()
}

func (r *rows) UpdateSeq() string {
	return r.updateSeq
}

func (*rows) Offset() int64 {
	return 0
}

func (*rows) TotalRows() int64 {
	return 0
}

func (r *rows) Bookmark() string {
	return r.bookmark
}
