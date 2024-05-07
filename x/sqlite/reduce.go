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

	"github.com/dop251/goja"

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

type preAggregateStats struct {
	Sum    *float64 `json:"sum"`
	Min    *float64 `json:"min"`
	Max    *float64 `json:"max"`
	Count  *float64 `json:"count"`
	SumSqr *float64 `json:"sumsqr"`
	raw    json.RawMessage
}

func (s *preAggregateStats) UnmarshalJSON(data []byte) error {
	alias := struct {
		preAggregateStats
		UnmarshalJSON struct{} `json:"-"`
	}{
		preAggregateStats: *s,
	}
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*s = alias.preAggregateStats
	s.raw = data
	return nil
}

func (s preAggregateStats) Validate() error {
	fieldError := func(field string) error {
		return &internal.Error{
			Status:  http.StatusInternalServerError,
			Message: fmt.Sprintf("user _stats input missing required field %s (%s)", field, string(s.raw)),
		}
	}
	if s.Sum == nil {
		return fieldError("sum")
	}
	if s.Count == nil {
		return fieldError("count")
	}
	if s.Min == nil {
		return fieldError("min")
	}
	if s.Max == nil {
		return fieldError("max")
	}
	if s.SumSqr == nil {
		return fieldError("sumsqr")
	}
	return nil
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
		value, ok := toFloat64(v)
		if !ok {
			if strValue, ok := v.(string); ok {
				var mapStats preAggregateStats

				if err := json.Unmarshal([]byte(strValue), &mapStats); err == nil {
					if err := mapStats.Validate(); err != nil {
						return nil, err
					}
					// The map function emitted pre-aggregated stats
					result.Sum += *mapStats.Sum
					result.Count += *mapStats.Count
					result.SumSqr += *mapStats.SumSqr
					result.Count-- // don't double-count the map stats
					mins = append(mins, *mapStats.Min)
					maxs = append(maxs, *mapStats.Max)
					continue
				}
			}
			val, _ := v.(string)
			if v == nil {
				val = "null"
			}
			return nil, &internal.Error{
				Status:  http.StatusInternalServerError,
				Message: fmt.Sprintf("the _stats function requires that map values be numbers or arrays of numbers, not '%s'", val),
			}
		}
		nvals = append(nvals, value)
		result.Sum += value
		result.SumSqr += value * value
	}
	result.Min = slices.Min(slices.Concat(nvals, mins))
	result.Max = slices.Max(slices.Concat(nvals, maxs))
	return result, nil
}
