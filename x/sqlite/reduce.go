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
	"database/sql"
	"encoding/json"
	"io"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/x/sqlite/v4/reduce"
)

type reduceRowIter struct {
	results *sql.Rows
}

type reduceRow struct {
	ID    string
	Key   string
	Value *string
}

func (r *reduceRowIter) ReduceNext(row *reduce.Row) error {
	if !r.results.Next() {
		if err := r.results.Err(); err != nil {
			return err
		}
		return io.EOF
	}
	var key, value *[]byte
	err := r.results.Scan(
		&row.ID, &key, &value, discard{}, discard{}, discard{},
		discard{}, discard{}, discard{}, discard{}, discard{}, discard{}, discard{},
	)
	if err != nil {
		return err
	}
	if key != nil {
		if err = json.Unmarshal(*key, &row.Key); err != nil {
			return err
		}
	} else {
		row.Key = nil
	}
	if value != nil {
		if err = json.Unmarshal(*value, &row.Value); err != nil {
			return err
		}
	} else {
		row.Value = nil
	}
	return err
}

func (d *db) reduceRows(ri reduce.Reducer, javascript string, vopts *viewOptions) (driver.Rows, error) {
	return reduce.Reduce(ri, javascript, d.logger, vopts.reduceGroupLevel(), nil)
}

type reducedRows []driver.Row

var _ driver.Rows = (*reducedRows)(nil)

func (r *reducedRows) Close() error {
	*r = nil
	return nil
}

func (r *reducedRows) Next(row *driver.Row) error {
	if len(*r) == 0 {
		return io.EOF
	}
	*row = (*r)[0]
	*r = (*r)[1:]
	return nil
}

func (*reducedRows) Offset() int64     { return 0 }
func (*reducedRows) TotalRows() int64  { return 0 }
func (*reducedRows) UpdateSeq() string { return "" }
