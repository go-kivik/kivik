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
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/dop251/goja"
	"github.com/mitchellh/mapstructure"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/x/sqlite/v4/internal"
)

func (d *db) reduceRows(results *sql.Rows, reduceFuncJS *string, group bool, groupLevel uint64) (driver.Rows, error) {
	reduceFn, err := d.reduceFunc(reduceFuncJS, d.logger)
	if err != nil {
		return nil, err
	}
	var (
		intermediate = map[string][]interface{}{}

		id, rowKey string
		rowValue   *string
	)

	for results.Next() {
		if err := results.Scan(&id, &rowKey, &rowValue, discard{}, discard{}, discard{}); err != nil {
			return nil, err
		}
		var key, value interface{}
		_ = json.Unmarshal([]byte(rowKey), &key)
		if rowValue != nil {
			_ = json.Unmarshal([]byte(*rowValue), &value)
		}
		rv, err := reduceFn([][2]interface{}{{id, key}}, []interface{}{value}, false)
		if err != nil {
			return nil, err
		}
		// group is handled below
		if groupLevel > 0 {
			var unkey []interface{}
			_ = json.Unmarshal([]byte(rowKey), &unkey)
			if len(unkey) > int(groupLevel) {
				newKey, _ := json.Marshal(unkey[:groupLevel])
				rowKey = string(newKey)
			}
		}
		intermediate[rowKey] = append(intermediate[rowKey], rv)
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

type reduceFunc func(keys [][2]interface{}, values []interface{}, rereduce bool) (interface{}, error)

func (d *db) reduceFunc(reduceFuncJS *string, logger *log.Logger) (reduceFunc, error) {
	if reduceFuncJS == nil {
		return nil, nil
	}
	switch *reduceFuncJS {
	case "_count":
		return reduceCount, nil
	case "_sum":
		return reduceSum, nil
	case "_stats":
		return reduceStats, nil
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

func reduceCount(_ [][2]interface{}, values []interface{}, rereduce bool) (interface{}, error) {
	if !rereduce {
		return len(values), nil
	}
	var total uint64
	for _, value := range values {
		v, _ := toUint64(value, "")
		total += v
	}
	return total, nil
}

func reduceSum(_ [][2]interface{}, values []interface{}, _ bool) (interface{}, error) {
	var total uint64
	for _, value := range values {
		v, _ := toUint64(value, "")
		total += v
	}
	return total, nil
}

type stats struct {
	Sum    float64 `json:"sum"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Count  float64 `json:"count"`
	SumSqr float64 `json:"sumsqr"`
}

func reduceStats(_ [][2]interface{}, values []interface{}, rereduce bool) (interface{}, error) {
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
	mins := make([]float64, 0, len(values))
	maxs := make([]float64, 0, len(values))
	for _, v := range values {
		switch t := v.(type) {
		case float64:
			nvals = append(nvals, t)
			result.Sum += t
			result.SumSqr += t * t
			continue
		case nil:
			// jump to end of switch, to return an error
		default:
			var (
				mapStats stats
				metadata mapstructure.Metadata
			)

			if err := mapstructure.DecodeMetadata(v, &mapStats, &metadata); err == nil {
				if len(metadata.Unset) > 0 {
					raw, _ := json.Marshal(v)
					slices.Sort(metadata.Unset)
					return nil, &internal.Error{
						Status:  http.StatusInternalServerError,
						Message: fmt.Sprintf("user _stats input missing required field %s (%s)", strings.ToLower(metadata.Unset[0]), string(raw)),
					}
				}
				// The map function emitted pre-aggregated stats
				result.Sum += mapStats.Sum
				result.Count += mapStats.Count
				result.SumSqr += mapStats.SumSqr
				result.Count-- // don't double-count the map stats
				mins = append(mins, mapStats.Min)
				maxs = append(maxs, mapStats.Max)
				continue
			}
		}

		valBytes, _ := json.Marshal(v)
		return nil, &internal.Error{
			Status:  http.StatusInternalServerError,
			Message: fmt.Sprintf("the _stats function requires that map values be numbers or arrays of numbers, not '%s'", string(valBytes)),
		}
	}
	result.Min = slices.Min(slices.Concat(nvals, mins))
	result.Max = slices.Max(slices.Concat(nvals, maxs))
	return result, nil
}
