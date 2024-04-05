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

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal"
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

	since, err := opts.since()
	if err != nil {
		return nil, err
	}
	limit, err := opts.limit()
	if err != nil {
		return nil, err
	}

	if since != nil {
		var lastSeq uint64
		err := d.db.QueryRowContext(ctx, d.query(`
			SELECT COALESCE(MAX(seq), 0) FROM {{ .Docs }}
		`), *since).Scan(&lastSeq)
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
	case "normal":
		query := d.query(`
			WITH results AS (
				SELECT
					id,
					seq,
					deleted,
					rev,
					rev_id
				FROM test
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
	case "longpoll":
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
