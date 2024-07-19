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
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/x/sqlite/v4/js"
)

const (
	feedNormal   = "normal"
	feedLongpoll = "longpoll"
)

type normalChanges struct {
	rows        *sql.Rows
	pending     int64
	lastSeq     string
	etag        string
	includeDocs bool
	filter      func(doc any, req any) (bool, error)
}

var _ driver.Changes = &normalChanges{}

func (d *db) newNormalChanges(ctx context.Context, opts optsMap, since, lastSeq *uint64, sinceNow bool, feed string) (*normalChanges, error) {
	limit, err := opts.changesLimit()
	if err != nil {
		return nil, err
	}

	if since != nil {
		last, err := d.lastSeq(ctx)
		if err != nil {
			return nil, err
		}
		lastSeq = &last
		if last <= *since {
			*since = last - 1
			limit = uint64(1)
		}
	}

	c := &normalChanges{}

	if sinceNow {
		if lastSeq == nil {
			last, err := d.lastSeq(ctx)
			if err != nil {
				return nil, err
			}
			lastSeq = &last
		}
		since = lastSeq
		limit = 0
		c.lastSeq = strconv.FormatUint(*lastSeq, 10)
	}

	descending, err := opts.descending()
	if err != nil {
		return nil, err
	}

	c.includeDocs, err = opts.includeDocs()
	if err != nil {
		return nil, err
	}
	attachments, err := opts.attachments()
	if err != nil {
		return nil, err
	}

	filterDdoc, filterName, err := opts.changesFilter()
	if err != nil {
		return nil, err
	}

	args := []any{since, attachments, c.includeDocs, filterDdoc, filterName}
	where, err := opts.changesWhere(&args)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(d.query(leavesCTE+`,
		results AS (
			SELECT
				id,
				seq,
				deleted,
				rev,
				rev_id,
				IIF($3 OR $5 != '', doc, NULL) AS doc
			FROM {{ .Docs }}
			WHERE ($1 IS NULL OR seq > $1)
			ORDER BY seq
		)
		SELECT
			COUNT(*) AS id,
			(
				-- Will return NULL if the ddoc doesn't exist, empty string
				-- if it exists but the filter func doesn't exist, or the
				-- filter function if it does exist.
				SELECT
					COALESCE(design.func_body, '')
				FROM leaves
				LEFT JOIN {{ .Design }} AS design ON design.id = leaves.id AND design.rev = leaves.rev AND design.rev_id = leaves.rev_id AND design.func_type = 'filter' AND design.func_name = $5
				WHERE leaves.id = $4
				ORDER BY leaves.rev DESC
				LIMIT 1
			) AS filter_func,
			NULL AS deleted,
			COALESCE(MAX(seq),0) AS summary,
			NULL AS doc,
			NULL AS attachment_count,
			NULL AS filename,
			NULL AS content_type,
			NULL AS length,
			NULL AS digest,
			NULL AS rev_pos,
			NULL AS data
		FROM results

		UNION ALL

		SELECT
			CASE WHEN row_number = 1 THEN id END AS id,
			CASE WHEN row_number = 1 THEN seq END AS seq,
			CASE WHEN row_number = 1 THEN deleted END AS deleted,
			CASE WHEN row_number = 1 THEN rev END AS rev,
			CASE WHEN row_number = 1 THEN doc END AS doc,
			attachment_count,
			filename,
			content_type,
			length,
			digest,
			rev_pos,
			data
		FROM (
			SELECT
				results.id,
				results.seq,
				results.deleted,
				results.rev || '-' || results.rev_id AS rev,
				results.doc,
				SUM(CASE WHEN bridge.pk IS NOT NULL THEN 1 ELSE 0 END) OVER (PARTITION BY results.id, results.rev, results.rev_id) AS attachment_count,
				ROW_NUMBER() OVER (PARTITION BY results.id, results.rev, results.rev_id) AS row_number,
				att.filename,
				att.content_type,
				att.length,
				att.digest,
				att.rev_pos,
				IIF($2, data, NULL) AS data
			FROM results
			LEFT JOIN {{ .AttachmentsBridge }} AS bridge ON bridge.id = results.id AND bridge.rev = results.rev AND bridge.rev_id = results.rev_id AND $3
			LEFT JOIN {{ .Attachments }} AS att ON att.pk = bridge.pk
			%[2]s -- WHERE
			ORDER BY seq %[1]s
		)
	`), descendingToDirection(descending), where)

	if limit > 0 {
		query += " LIMIT " + strconv.FormatUint(limit+1, 10)
	}

	c.rows, err = d.db.QueryContext(ctx, query, args...) //nolint:rowserrcheck,sqlclosecheck // Err checked in Next
	if err != nil {
		return nil, err
	}

	// The first row is used to calculate the ETag; it's done as part of the
	// same query, even though it's a bit ugly, to ensure it's all in the same
	// implicit transaction.
	if !c.rows.Next() {
		// should never happen
		return nil, errors.New("no rows returned")
	}
	var (
		summary      string
		filterFuncJS *string
	)
	if err := c.rows.Scan(
		&c.pending, &filterFuncJS,
		discard{},
		&summary,
		discard{}, discard{}, discard{}, discard{}, discard{}, discard{}, discard{}, discard{},
	); err != nil {
		return nil, err
	}

	if filterName != "" {
		if filterFuncJS == nil {
			return nil, &internal.Error{Status: http.StatusNotFound, Message: fmt.Sprintf("design doc '%s' not found", filterDdoc)}
		}
		if *filterFuncJS == "" {
			return nil, &internal.Error{Status: http.StatusNotFound, Message: fmt.Sprintf("design doc '%s' missing filter function '%s'", filterDdoc, filterName)}
		}

		c.filter, err = js.Filter(*filterFuncJS)
		if err != nil {
			return nil, &internal.Error{Status: http.StatusInternalServerError, Err: err}
		}
	}

	if feed == feedNormal {
		h := md5.New()
		_, _ = h.Write([]byte(summary))
		c.etag = hex.EncodeToString(h.Sum(nil))
	}

	return c, nil
}

func (c *normalChanges) Next(change *driver.Change) error {
	var (
		rev             string
		doc             *string
		atts            map[string]*attachment
		attachmentCount = 1
	)

	for {
		if !c.rows.Next() {
			if err := c.rows.Err(); err != nil {
				return err
			}
			return io.EOF
		}
		var (
			rowID, rowSeq, rowRev, rowDoc *string
			rowDeleted                    *bool
			filename, contentType         *string
			length                        *int64
			revPos                        *int
			digest                        *md5sum
			data                          *[]byte
		)
		if err := c.rows.Scan(
			&rowID, &rowSeq, &rowDeleted, &rowRev, &rowDoc,
			&attachmentCount, &filename, &contentType, &length, &digest, &revPos, &data,
		); err != nil {
			return err
		}
		if rowID != nil {
			change.ID = *rowID
			change.Seq = *rowSeq
			change.Deleted = *rowDeleted
			change.Changes = driver.ChangedRevs{*rowRev}
			c.lastSeq = change.Seq
			rev = *rowRev
			doc = rowDoc
		}
		if filename != nil {
			if atts == nil {
				atts = map[string]*attachment{}
			}
			att := &attachment{
				ContentType: *contentType,
				Digest:      *digest,
				Length:      *length,
				RevPos:      *revPos,
			}
			if data != nil {
				att.Data, _ = json.Marshal(*data)
			}

			atts[*filename] = att
		}
		if attachmentCount == len(atts) {
			break
		}
	}
	c.pending--

	if doc != nil {
		toMerge := fullDoc{
			ID:          change.ID,
			Rev:         rev,
			Deleted:     change.Deleted,
			Doc:         []byte(*doc),
			Attachments: atts,
		}
		if c.filter != nil {
			ok, err := c.filter(toMerge.toMap(), nil)
			if err != nil {
				return &internal.Error{Status: http.StatusInternalServerError, Err: err}
			}
			if !ok {
				return c.Next(change)
			}
		}
		if c.includeDocs {
			change.Doc = toMerge.toRaw()
		}
	}
	return nil
}

func (c *normalChanges) Close() error {
	return c.rows.Close()
}

func (c *normalChanges) LastSeq() string {
	// Columns returns an error if the rows are closed, so we can use that to
	// determine if we've actually read the last sequence id.
	if _, err := c.rows.Columns(); err == nil {
		return ""
	}
	return c.lastSeq
}

func (c *normalChanges) Pending() int64 {
	return c.pending
}

func (c *normalChanges) ETag() string {
	return c.etag
}

func (d *db) Changes(ctx context.Context, options driver.Options) (driver.Changes, error) {
	opts := newOpts(options)

	var lastSeq *uint64
	sinceNow, since, err := opts.since()
	if err != nil {
		return nil, err
	}
	feed, err := opts.feed()
	if err != nil {
		return nil, err
	}
	includeDocs, err := opts.includeDocs()
	if err != nil {
		return nil, err
	}
	if sinceNow && feed == feedLongpoll {
		attachments, err := opts.attachments()
		if err != nil {
			return nil, err
		}
		return d.newLongpollChanges(ctx, includeDocs, attachments)
	}

	return d.newNormalChanges(ctx, opts, since, lastSeq, sinceNow, feed)
}

type longpollChanges struct {
	stmt        *sql.Stmt
	since       uint64
	includeDocs bool
	attachments bool
	lastSeq     string
	ctx         context.Context
	cancel      context.CancelFunc
	changes     <-chan longpollChange
}

type longpollChange struct {
	change *driver.Change
	err    error
}

var _ driver.Changes = (*longpollChanges)(nil)

func (d *db) newLongpollChanges(ctx context.Context, includeDocs, attachments bool) (*longpollChanges, error) {
	since, err := d.lastSeq(ctx)
	if err != nil {
		return nil, err
	}

	stmt, err := d.db.PrepareContext(ctx, d.query(`
		SELECT
			CASE WHEN row_number = 1 THEN id END AS id,
			CASE WHEN row_number = 1 THEN seq END AS seq,
			CASE WHEN row_number = 1 THEN deleted END AS deleted,
			CASE WHEN row_number = 1 THEN rev END AS rev,
			CASE WHEN row_number = 1 THEN doc END AS doc,
			filename,
			content_type,
			length,
			digest,
			rev_pos,
			data
		FROM (
			SELECT
				doc.id,
				doc.seq,
				doc.deleted,
				doc.rev || '-' || doc.rev_id AS rev,
				doc,
				att.filename,
				att.content_type,
				att.length,
				att.digest,
				att.rev_pos,
				IIF($2, att.data, NULL) AS data,
				ROW_NUMBER() OVER () AS row_number
			FROM (
				SELECT
					id,
					seq,
					deleted,
					rev,
					rev_id,
					IIF($3, doc, NULL) AS doc
				FROM {{ .Docs }}
				WHERE seq > $1
				ORDER BY seq
				LIMIT 1
			) AS doc
			LEFT JOIN {{ .AttachmentsBridge }} AS bridge ON bridge.id = doc.id AND bridge.rev = doc.rev AND bridge.rev_id = doc.rev_id AND doc IS NOT NULL
			LEFT JOIN {{ .Attachments }} AS att ON att.pk = bridge.pk
		)
	`))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	changes := make(chan longpollChange)
	c := &longpollChanges{
		stmt:        stmt,
		since:       since,
		attachments: attachments,
		includeDocs: includeDocs,
		ctx:         ctx,
		cancel:      cancel,
		changes:     changes,
	}

	go c.watch(changes)

	return c, nil
}

// watch runs in a loop until either the context is cancelled, or a change is
// detected.
func (c *longpollChanges) watch(changes chan<- longpollChange) {
	defer close(changes)

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 50 * time.Millisecond
	bo.MaxInterval = 3 * time.Minute
	bo.MaxElapsedTime = 0

	err := backoff.Retry(func() error {
		rows, err := c.stmt.QueryContext(c.ctx, c.since, c.attachments, c.includeDocs)
		if err != nil {
			return backoff.Permanent(err)
		}
		defer rows.Close()

		var (
			rowID, rowSeq, rowRev, rowDoc *string
			rowDeleted                    *bool
			change                        driver.Change
			rev                           string
			doc                           *string
			atts                          map[string]*attachment
			filename                      *string
			contentType                   *string
			length                        *int64
			digest                        *md5sum
			revPos                        *int
			data                          *[]byte
		)
		for rows.Next() {
			if err := rows.Scan(
				&rowID, &rowSeq, &rowDeleted, &rowRev, &rowDoc,
				&filename, &contentType, &length, &digest, &revPos, &data,
			); err != nil {
				return backoff.Permanent(err)
			}
			if rowID != nil {
				change.ID = *rowID
				change.Seq = *rowSeq
				change.Deleted = *rowDeleted
				rev = *rowRev
				doc = rowDoc
			}

			if filename != nil {
				if atts == nil {
					atts = map[string]*attachment{}
				}
				att := &attachment{
					ContentType: *contentType,
					Digest:      *digest,
					Length:      *length,
					RevPos:      *revPos,
				}
				if data != nil {
					att.Data, _ = json.Marshal(*data)
				}
				atts[*filename] = att
			}
		}
		if err := rows.Err(); err != nil {
			return backoff.Permanent(err)
		}

		if change.ID == "" {
			return errors.New("retry")
		}

		change.Changes = driver.ChangedRevs{rev}
		c.lastSeq = change.Seq

		if doc != nil {
			toMerge := fullDoc{
				ID:          change.ID,
				Rev:         rev,
				Deleted:     change.Deleted,
				Doc:         []byte(*doc),
				Attachments: atts,
			}
			change.Doc = toMerge.toRaw()
		}

		changes <- longpollChange{change: &change}
		return nil
	}, bo)
	if err != nil {
		changes <- longpollChange{err: err}
	}
}

func (c *longpollChanges) Next(change *driver.Change) error {
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case ch, ok := <-c.changes:
		if !ok {
			return io.EOF
		}
		if ch.err != nil {
			return ch.err
		}
		*change = *ch.change
		return nil
	}
}

func (c *longpollChanges) Close() error {
	return nil
}

func (c *longpollChanges) LastSeq() string {
	return c.lastSeq
}

func (*longpollChanges) Pending() int64 { return 0 }
func (*longpollChanges) ETag() string   { return "" }
