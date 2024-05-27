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
	if err := r.results.Scan(
		&row.ID, &row.Key, &row.Value, discard{}, discard{}, discard{},
		discard{}, discard{}, discard{}, discard{}, discard{}, discard{}, discard{},
	); err != nil {
		return nil, err
	}
	return &row, nil
}

type reduceRows interface {
	Next() (*reduceRow, error)
}

func (d *db) reduceRows(ri reduceRows, reduceFuncJS *string, vopts *viewOptions) (*reducedRows, error) {
	reduceFn, err := d.reduceFunc(reduceFuncJS, d.logger)
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

// toFloatValues converts values to a slice of float64 slices, if possible.
// This is used when a map function returns an array of numbers to be aggregated
// by the _stats function
func toFloatValues(values []interface{}, rereduce bool) ([][]float64, bool) {
	if rereduce {
		return nil, false
	}
	_, isSlice := values[0].([]interface{})
	if !isSlice {
		return nil, false
	}
	floatValues := make([][]float64, 0, len(values))
	for _, v := range values {
		fv := v.([]interface{})
		float := make([]float64, 0, len(fv))
		for _, f := range fv {
			floatValue, ok := f.(float64)
			if !ok {
				return nil, false
			}
			float = append(float, floatValue)
		}
		floatValues = append(floatValues, float)
	}
	return floatValues, true
}

func toStatsValues(values []interface{}, rereduce bool) ([][]stats, bool) {
	if !rereduce {
		return nil, false
	}
	_, isSlice := values[0].([]stats)
	if !isSlice {
		return nil, false
	}
	statsValues := make([][]stats, 0, len(values))
	for _, v := range values {
		statsValues = append(statsValues, v.([]stats))
	}
	return statsValues, true
}

func reduceStats(_ [][2]interface{}, values []interface{}, rereduce bool) (interface{}, error) {
	if floatValues, ok := toFloatValues(values, rereduce); ok {
		return reduceStatsFloatArray(floatValues), nil
	}
	statsValues, ok := toStatsValues(values, rereduce)
	if ok {
		return rereduceStatsFloatArray(statsValues), nil
	}

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

func reduceStatsFloatArray(values [][]float64) []stats {
	results := make([]stats, len(values[0]))
	minmax := make([][]float64, len(values[0]))

	for _, numbers := range values {
		for i, v := range numbers {
			results[i].Sum += v
			results[i].SumSqr += v * v
			minmax[i] = append(minmax[i], v)
		}
	}
	for i, mm := range minmax {
		results[i].Count += float64(len(values))
		results[i].Min = slices.Min(mm)
		results[i].Max = slices.Max(mm)
	}

	return results
}

func rereduceStatsFloatArray(values [][]stats) []stats {
	result := make([]stats, len(values[0]))
	mins := make([][]float64, len(values[0]))
	maxs := make([][]float64, len(values[0]))
	for _, value := range values {
		for j, stat := range value {
			result[j].Sum += stat.Sum
			result[j].Count += stat.Count
			result[j].SumSqr += stat.SumSqr
			mins[j] = append(mins[j], stat.Min)
			maxs[j] = append(maxs[j], stat.Max)
		}
	}
	for i, mm := range mins {
		result[i].Min = slices.Min(mm)
		result[i].Max = slices.Max(maxs[i])
	}
	return result
}
