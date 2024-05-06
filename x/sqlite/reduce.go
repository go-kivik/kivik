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
)

func (d *db) reduceRows(results *sql.Rows, reduceFuncJS *string, group bool, groupLevel uint64) (driver.Rows, error) {
	reduceFn, err := d.reduceFunc(reduceFuncJS, d.logger)
	if err != nil {
		return nil, err
	}
	var (
		intermediate = map[string][]interface{}{}

		id, key  string
		rowValue *string
	)

	for results.Next() {
		if err := results.Scan(&id, &key, &rowValue, discard{}, discard{}, discard{}); err != nil {
			return nil, err
		}
		var value interface{}
		if rowValue != nil {
			value = *rowValue
		}
		rv, err := reduceFn([][2]interface{}{{id, key}}, []interface{}{value}, false)
		if err != nil {
			return nil, err
		}
		// group is handled below
		if groupLevel > 0 {
			var unkey []interface{}
			_ = json.Unmarshal([]byte(key), &unkey)
			if len(unkey) > int(groupLevel) {
				newKey, _ := json.Marshal(unkey[:groupLevel])
				key = string(newKey)
			}
		}
		intermediate[key] = append(intermediate[key], rv)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	// group_level is handled above
	if !group {
		var values []interface{}
		for _, v := range intermediate {
			values = append(values, v...)
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

	slices.SortFunc(final, func(a, b driver.Row) int {
		return couchdbCmpJSON(a.Key, b.Key)
	})

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
