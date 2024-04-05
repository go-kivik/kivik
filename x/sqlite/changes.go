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
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal"
)

const (
	feedNormal   = "normal"
	feedLongpoll = "longpoll"
)

type changes struct {
	rows    *sql.Rows
	lastSeq string
	etag    string
}

var _ driver.Changes = &changes{}

func (c *changes) Next(change *driver.Change) error {
	if !c.rows.Next() {
		if err := c.rows.Err(); err != nil {
			return err
		}
		return io.EOF
	}
	var rev string
	if err := c.rows.Scan(&change.ID, &change.Seq, &change.Deleted, &rev); err != nil {
		return err
	}
	change.Changes = driver.ChangedRevs{rev}
	c.lastSeq = change.Seq
	return nil
}

func (c *changes) Close() error {
	return c.rows.Close()
}

func (c *changes) LastSeq() string {
	// Columns returns an error if the rows are closed, so we can use that to
	// determine if we've actually read the last sequence id.
	if _, err := c.rows.Columns(); err == nil {
		return ""
	}
	return c.lastSeq
}

func (c *changes) Pending() int64 {
	return 0
}

func (c *changes) ETag() string {
	return c.etag
}

func (d *db) Changes(ctx context.Context, options driver.Options) (driver.Changes, error) {
	opts := newOpts(options)
	var (
		rows *sql.Rows
		etag string
	)

	sinceNow, since, err := opts.since()
	if err != nil {
		return nil, err
	}
	if sinceNow && opts.feed() == feedLongpoll {
		return d.newLongpollChanges(ctx)
	}
	limit, err := opts.limit()
	if err != nil {
		return nil, err
	}

	if since != nil {
		lastSeq, err := d.lastSeq(ctx)
		if err != nil {
			return nil, err
		}
		if lastSeq <= *since {
			*since = lastSeq - 1
			l := uint64(1)
			limit = &l
		}
	}

	switch opts.feed() {
	case feedNormal:
		query := d.query(`
			WITH results AS (
				SELECT
					id,
					seq,
					deleted,
					rev,
					rev_id
				FROM {{ .Docs }}
				WHERE ($1 IS NULL OR seq > $1)
				ORDER BY seq
			)
			SELECT
				NULL AS id,
				NULL AS seq,
				NULL AS deleted,
				COUNT(*) || '.' || COALESCE(MIN(seq),0) || '.' || COALESCE(MAX(seq),0) AS rev
			FROM results

			UNION ALL

			SELECT
				id,
				seq,
				deleted,
				rev || '-' || rev_id AS rev
			FROM results
		`)
		if limit != nil {
			query += " LIMIT " + strconv.FormatUint(*limit+1, 10)
		}
		var err error
		rows, err = d.db.QueryContext(ctx, query, since) //nolint:rowserrcheck // Err checked in Next
		if err != nil {
			return nil, err
		}

		// The first row is used to calculate the ETag; it's done as part of the
		// same query, even though it's a bit ugly, to ensure it's all in the same
		// implicit transaction.
		if !rows.Next() {
			// should never happen
			return nil, errors.New("no rows returned")
		}
		var discard *string
		var summary string
		if err := rows.Scan(&discard, &discard, &discard, &summary); err != nil {
			return nil, err
		}

		h := md5.New()
		_, _ = h.Write([]byte(summary))
		etag = hex.EncodeToString(h.Sum(nil))
	case feedLongpoll:
		query := d.query(`
			SELECT
				id,
				seq,
				deleted,
				rev || '-' || rev_id AS rev
			FROM {{ .Docs }}
			WHERE ($1 IS NULL OR seq > $1)
			ORDER BY seq
		`)
		if limit != nil {
			query += " LIMIT " + strconv.FormatUint(*limit, 10)
		}
		var err error
		rows, err = d.db.QueryContext(ctx, query, since) //nolint:rowserrcheck // Err checked in Next
		if err != nil {
			return nil, err
		}
	default:
		return nil, &internal.Error{Status: http.StatusBadRequest, Message: "supported `feed` types: normal, longpoll"}
	}

	return &changes{
		rows: rows,
		etag: etag,
	}, nil
}

type longpollChanges struct {
	stmt    *sql.Stmt
	since   uint64
	lastSeq string
	ctx     context.Context
	cancel  context.CancelFunc
	changes <-chan longpollChange
}

type longpollChange struct {
	change *driver.Change
	err    error
}

var _ driver.Changes = (*longpollChanges)(nil)

func (d *db) newLongpollChanges(ctx context.Context) (*longpollChanges, error) {
	since, err := d.lastSeq(ctx)
	if err != nil {
		return nil, err
	}

	stmt, err := d.db.PrepareContext(ctx, d.query(`
		SELECT
			id,
			seq,
			deleted,
			rev || '-' || rev_id AS rev
		FROM {{ .Docs }}
		WHERE seq > $1
		ORDER BY seq
		LIMIT 1
	`))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	changes := make(chan longpollChange)
	c := &longpollChanges{
		stmt:    stmt,
		since:   since,
		ctx:     ctx,
		cancel:  cancel,
		changes: changes,
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

	var change driver.Change
	var rev string
	err := backoff.Retry(func() error {
		err := c.stmt.QueryRowContext(c.ctx, c.since).Scan(&change.ID, &change.Seq, &change.Deleted, &rev)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return err
		case err != nil:
			return backoff.Permanent(err)
		default:
			change.Changes = driver.ChangedRevs{rev}
			c.lastSeq = change.Seq

			changes <- longpollChange{change: &change}
			return nil
		}
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
