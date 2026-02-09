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

package reduce

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/mitchellh/mapstructure"

	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/x/sqlite/v4/js"
)

// Count is the built-in reduce function, [_count].
//
// [_count]: https://docs.couchdb.org/en/stable/ddocs/ddocs.html#count
func Count(_ [][2]any, values []any, rereduce bool) ([]any, error) {
	if !rereduce {
		return []any{float64(len(values))}, nil
	}
	var total float64
	for _, value := range values {
		if value != nil {
			total += value.(float64)
		}
	}
	return []any{total}, nil
}

// Sum is the built-in reduce function, [_sum].
//
// [_sum]: https://docs.couchdb.org/en/stable/ddocs/ddocs.html#sum
func Sum(_ [][2]any, values []any, _ bool) ([]any, error) {
	var total float64
	for _, value := range values {
		if value != nil {
			total += value.(float64)
		}
	}
	return []any{total}, nil
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
func toFloatValues(values []any, rereduce bool) ([][]float64, bool) {
	if rereduce {
		return nil, false
	}
	_, isSlice := values[0].([]any)
	if !isSlice {
		return nil, false
	}
	floatValues := make([][]float64, 0, len(values))
	for _, v := range values {
		fv := v.([]any)
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

func toStatsValues(values []any, rereduce bool) ([][]stats, bool) {
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

func flattenStats(values []any) []stats {
	statsValues := make([]stats, 0, len(values))
	for _, v := range values {
		switch t := v.(type) {
		case stats:
			statsValues = append(statsValues, t)
		case []any:
			for _, vv := range t {
				if stat, ok := vv.(stats); ok {
					statsValues = append(statsValues, stat)
				}
			}
		}
	}
	return statsValues
}

// Stats is the built-in reduce function, [_stats].
//
// [_stats]: https://docs.couchdb.org/en/stable/ddocs/ddocs.html#stats
func Stats(_ [][2]any, values []any, rereduce bool) ([]any, error) {
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
		for _, value := range flattenStats(values) {
			mins = append(mins, value.Min)
			maxs = append(maxs, value.Max)
			result.Sum += value.Sum
			result.Count += value.Count
			result.SumSqr += value.SumSqr
		}
		result.Min = slices.Min(mins)
		result.Max = slices.Max(maxs)
		return []any{result}, nil
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
	return []any{result}, nil
}

func reduceStatsFloatArray(values [][]float64) []any {
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

	return []any{results}
}

func rereduceStatsFloatArray(values [][]stats) []any {
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

	return []any{result}
}

// ParseFunc parses the passed javascript string, and returns a Go function that
// will execute it.  If the input is empty, nil is returned. If the input is a
// string that corresponds to one of the built-in function names (i.e. '_sum',
// '_count', etc), the native Go implementation is returned instead. The logger
// is used to log any unhandled exceptions thrown by the JavaScript function.
func ParseFunc(javascript string, logger *log.Logger) (Func, error) {
	switch javascript {
	case "":
		return nil, nil
	case "_count":
		return Count, nil
	case "_sum":
		return Sum, nil
	case "_stats":
		return Stats, nil
	default:
		reduceFunc, err := js.Reduce(javascript)
		if err != nil {
			return nil, err
		}
		return func(keys [][2]any, values []any, rereduce bool) ([]any, error) {
			ret, err := reduceFunc(keys, values, rereduce)
			// According to CouchDB reference implementation, when a user-defined
			// reduce function throws an exception, the error is logged and the
			// return value is set to null.
			if err != nil {
				logger.Printf("reduce function threw exception: %s", err)
				return []any{nil}, nil
			}
			return ret, nil
		}, nil
	}
}
