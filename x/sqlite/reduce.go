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
	"database/sql"
	"encoding/json"
	"io"
	"slices"

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

func (r *reduceRowIter) Next() (*reduceRow, error) {
	if !r.results.Next() {
		if err := r.results.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}
	var row reduceRow
	err := r.results.Scan(
		&row.ID, &row.Key, &row.Value, discard{}, discard{}, discard{},
		discard{}, discard{}, discard{}, discard{}, discard{}, discard{}, discard{},
	)
	return &row, err
}

type reduceRows interface {
	Next() (*reduceRow, error)
}

func (d *db) reduceRows(ri reduceRows, reduceFuncJS *string, vopts *viewOptions) (*reducedRows, error) {
	reduceFn, err := d.reduceFunc(reduceFuncJS)
	if err != nil {
		return nil, err
	}
	intermediate := map[string][]interface{}{}

	for {
		row, err := ri.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		var key, value interface{}
		_ = json.Unmarshal([]byte(row.Key), &key)
		if row.Value != nil {
			_ = json.Unmarshal([]byte(*row.Value), &value)
		}
		rv, err := reduceFn([][2]interface{}{{row.ID, key}}, []interface{}{value}, false)
		if err != nil {
			return nil, err
		}
		// group is handled below
		if vopts.groupLevel > 0 {
			var unkey []interface{}
			_ = json.Unmarshal([]byte(row.Key), &unkey)
			if len(unkey) > int(vopts.groupLevel) {
				newKey, _ := json.Marshal(unkey[:vopts.groupLevel])
				row.Key = string(newKey)
			}
		}
		intermediate[row.Key] = append(intermediate[row.Key], rv)
	}

	// group_level is handled above
	if !vopts.group {
		var values []interface{}
		for _, v := range intermediate {
			values = append(values, v...)
		}
		if len(values) == 0 {
			return &reducedRows{}, nil
		}
		rv, err := reduceFn(nil, values, true)
		if err != nil {
			return nil, err
		}
		tmp, _ := json.Marshal(rv)
		return &reducedRows{
			{
				Key:   json.RawMessage(`null`),
				Value: bytes.NewReader(tmp),
			},
		}, nil
	}

	final := make(reducedRows, 0, len(intermediate))
	for key, values := range intermediate {
		var value json.RawMessage
		if len(values) > 1 {
			rv, err := reduceFn(nil, values, true)
			if err != nil {
				return nil, err
			}
			value, _ = json.Marshal(rv)
		} else {
			value, _ = json.Marshal(values[0])
		}
		final = append(final, driver.Row{
			Key:   json.RawMessage(key),
			Value: bytes.NewReader(value),
		})
	}

	if vopts.sorted {
		slices.SortFunc(final, func(a, b driver.Row) int {
			return couchdbCmpJSON(a.Key, b.Key)
		})
		if vopts.descending {
			slices.Reverse(final)
		}
	}

	return &final, nil
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

type reduceFunc func(keys [][2]interface{}, values []interface{}, rereduce bool) (interface{}, error)

func (d *db) reduceFunc(reduceFuncJS *string) (reduceFunc, error) {
	var js string
	if reduceFuncJS != nil {
		js = *reduceFuncJS
	}
	f, err := reduce.ParseFunc(js, d.logger)
	if err != nil {
		return nil, err
	}
	return func(keys [][2]interface{}, values []interface{}, rereduce bool) (interface{}, error) {
		out, err := f(keys, values, rereduce)
		if err != nil {
			return nil, err
		}
		if len(out) == 1 {
			return out[0], nil
		}
		return out, nil
	}, nil
}
