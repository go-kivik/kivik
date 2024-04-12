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
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal"
)

func toRevDiffRequest(revMap interface{}) (map[string][]string, error) {
	data, err := json.Marshal(revMap)
	if err != nil {
		return nil, &internal.Error{Message: "invalid body", Status: http.StatusBadRequest}
	}
	var req map[string][]string
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, &internal.Error{Message: "invalid body", Status: http.StatusBadRequest}
	}
	return req, nil
}

func (d *db) RevsDiff(ctx context.Context, revMap interface{}) (driver.Rows, error) {
	req, err := toRevDiffRequest(revMap)
	if err != nil {
		return nil, err
	}

	ids := make([]interface{}, 0, len(req))
	for id := range req {
		ids = append(ids, id)
	}

	query := fmt.Sprintf(d.query(`
		SELECT
			id,
			rev,
			rev_count
		FROM (
			SELECT
				id,
				rev || '-' || rev_id AS rev,
				COUNT(*) OVER (PARTITION BY id) AS rev_count
			FROM {{ .Docs }}
			WHERE id IN (%s)
			ORDER BY id, rev, rev_id
		)
	`), placeholders(1, len(ids)))

	rows, err := d.db.QueryContext(ctx, query, ids...) //nolint:rowserrcheck // Err checked in Next
	if err != nil {
		return nil, err
	}

	return &revsDiffResponse{
		rows: rows,
		req:  req,
	}, nil
}

type revsDiffResponse struct {
	rows         *sql.Rows
	req          map[string][]string
	sortedDocIDs []string
}

var _ driver.Rows = (*revsDiffResponse)(nil)

func (r *revsDiffResponse) Next(row *driver.Row) error {
	var (
		id       string
		revs     = map[string]struct{}{}
		revCount int
	)

	for {
		if !r.rows.Next() {
			if err := r.rows.Err(); err != nil {
				return err
			}
			if len(r.req) > 0 {
				if len(r.sortedDocIDs) == 0 {
					// First time, we need to sort the remaining doc IDs.
					r.sortedDocIDs = make([]string, 0, len(r.req))
					for id := range r.req {
						r.sortedDocIDs = append(r.sortedDocIDs, id)
					}
					sort.Strings(r.sortedDocIDs)
				}

				row.ID = r.sortedDocIDs[0]
				revs := r.req[row.ID]
				sort.Strings(revs)
				row.Value = jsonToReader(driver.RevDiff{
					Missing: revs,
				})
				delete(r.req, row.ID)
				r.sortedDocIDs = r.sortedDocIDs[1:]
				return nil
			}
			return io.EOF
		}
		var (
			rowID *string
			rev   string
		)
		if err := r.rows.Scan(&id, &rev, &rowID, &revCount); err != nil {
			return err
		}
		if rowID != nil {
			id = *rowID
		}
		revs[id] = struct{}{}
		if len(revs) == revCount {
			break
		}
	}
	row.ID = id
	missing := make([]string, 0, len(r.req[id]))
	for _, rev := range r.req[id] {
		if _, ok := revs[rev]; !ok {
			missing = append(missing, rev)
		}
	}
	row.Value = jsonToReader(driver.RevDiff{
		Missing: missing,
	})
	return nil
}

type errorReadCloser struct{ error }

func (e errorReadCloser) Read([]byte) (int, error) { return 0, error(e) }
func (e errorReadCloser) Close() error             { return error(e) }

func jsonToReader(i interface{}) io.ReadCloser {
	value, err := json.Marshal(i)
	if err != nil {
		return errorReadCloser{err}
	}
	return io.NopCloser(bytes.NewReader(value))
}

func (r *revsDiffResponse) Close() error {
	return r.rows.Close()
}

func (*revsDiffResponse) Offset() int64     { return 0 }
func (*revsDiffResponse) TotalRows() int64  { return 0 }
func (*revsDiffResponse) UpdateSeq() string { return "" }
