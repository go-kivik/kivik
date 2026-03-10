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
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"math"
	"math/bits"
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
func Count(_ context.Context, _ [][2]any, values []any, rereduce bool) ([]any, error) {
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
func Sum(_ context.Context, _ [][2]any, values []any, _ bool) ([]any, error) {
	var totals []float64
	for _, value := range values {
		switch v := value.(type) {
		case float64:
			if totals == nil {
				totals = []float64{v}
			} else {
				totals[0] += v
			}
		case []any:
			for i, elem := range v {
				f, ok := elem.(float64)
				if !ok {
					continue
				}
				for len(totals) <= i {
					totals = append(totals, 0)
				}
				totals[i] += f
			}
		case map[string]any:
			if totals != nil {
				return nil, &internal.Error{
					Status:  http.StatusInternalServerError,
					Message: "the _sum function requires that objects not be mixed with other data structures",
				}
			}
			result, err := sumObjects(values)
			if err != nil {
				return nil, err
			}
			return []any{result}, nil
		default:
			valBytes, _ := json.Marshal(v)
			return nil, &internal.Error{
				Status:  http.StatusInternalServerError,
				Message: fmt.Sprintf("the _sum function requires that map values be numbers, arrays of numbers, or objects, not '%s'", string(valBytes)),
			}
		}
	}
	if totals == nil {
		totals = []float64{0}
	}
	result := make([]any, len(totals))
	for i, v := range totals {
		result[i] = v
	}
	return result, nil
}

func sumObjects(values []any) (map[string]any, error) {
	result := map[string]any{}
	for _, value := range values {
		obj, ok := value.(map[string]any)
		if !ok {
			return nil, &internal.Error{
				Status:  http.StatusInternalServerError,
				Message: "the _sum function requires that objects not be mixed with other data structures",
			}
		}
		for k, v := range obj {
			switch val := v.(type) {
			case float64:
				existing, _ := result[k].(float64)
				result[k] = existing + val
			case map[string]any:
				existing, ok := result[k].(map[string]any)
				if !ok {
					result[k] = val
				} else {
					merged, err := sumObjects([]any{existing, val})
					if err != nil {
						return nil, err
					}
					result[k] = merged
				}
			}
		}
	}
	return result, nil
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
func toFloatValues(values []any, rereduce bool) ([][]float64, bool, error) {
	if rereduce {
		return nil, false, nil
	}
	_, isSlice := values[0].([]any)
	if !isSlice {
		return nil, false, nil
	}
	expectedLen := -1
	floatValues := make([][]float64, 0, len(values))
	for _, v := range values {
		fv := v.([]any)
		if expectedLen == -1 {
			expectedLen = len(fv)
		} else if len(fv) != expectedLen {
			return nil, false, &internal.Error{
				Status:  http.StatusInternalServerError,
				Message: "the _stats function requires that map values be arrays of the same length",
			}
		}
		float := make([]float64, 0, len(fv))
		for _, f := range fv {
			floatValue, ok := f.(float64)
			if !ok {
				return nil, false, nil
			}
			float = append(float, floatValue)
		}
		floatValues = append(floatValues, float)
	}
	return floatValues, true, nil
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
func Stats(_ context.Context, _ [][2]any, values []any, rereduce bool) ([]any, error) {
	if len(values) == 0 {
		return nil, &internal.Error{
			Status:  http.StatusInternalServerError,
			Message: "the _stats function requires at least one value",
		}
	}
	if floatValues, ok, err := toFloatValues(values, rereduce); err != nil {
		return nil, err
	} else if ok {
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

const hllPrecision = 14

// hll is a HyperLogLog sketch for approximate distinct counting.
type hll struct {
	Registers [1 << hllPrecision]uint8
}

// MarshalJSON returns the HLL's cardinality estimate as a JSON number.
func (h *hll) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.Estimate())
}

// Estimate returns the approximate cardinality.
func (h *hll) Estimate() float64 {
	m := float64(len(h.Registers))
	var harmonicSum float64
	zeros := 0
	for _, val := range h.Registers {
		harmonicSum += 1.0 / float64(uint64(1)<<val)
		if val == 0 {
			zeros++
		}
	}
	alpha := 0.7213 / (1 + 1.079/m)
	estimate := alpha * m * m / harmonicSum
	if estimate <= 2.5*m && zeros > 0 {
		estimate = m * math.Log(m/float64(zeros))
	}
	return math.Round(estimate)
}

func (h *hll) add(data []byte) {
	hash := hash64(data)
	idx := hash >> (64 - hllPrecision)
	w := hash<<hllPrecision | (1 << (hllPrecision - 1))
	rho := uint8(bits.LeadingZeros64(w)) + 1
	if rho > h.Registers[idx] {
		h.Registers[idx] = rho
	}
}

func (h *hll) merge(other *hll) {
	for i, val := range other.Registers {
		if val > h.Registers[i] {
			h.Registers[i] = val
		}
	}
}

func hash64(data []byte) uint64 {
	h := fnv.New64a()
	h.Write(data)
	return h.Sum64()
}

// ApproxCountDistinct is the built-in reduce function,
// [_approx_count_distinct]. It uses HyperLogLog to estimate the number of
// distinct keys.
//
// [_approx_count_distinct]: https://docs.couchdb.org/en/stable/ddocs/ddocs.html#approx_count_distinct
func ApproxCountDistinct(_ context.Context, keys [][2]any, values []any, rereduce bool) ([]any, error) {
	h := &hll{}
	if rereduce {
		for _, v := range values {
			other, ok := v.(*hll)
			if !ok {
				continue
			}
			h.merge(other)
		}
		return []any{h}, nil
	}
	for _, key := range keys {
		keyBytes, _ := json.Marshal(key[0])
		h.add(keyBytes)
	}
	return []any{h}, nil
}

// ParseFunc parses the passed javascript string, and returns a Go function that
// implements the reduce function. Built-in functions (_count, _sum, _stats,
// _approx_count_distinct) are returned directly. User-defined functions are
// compiled using the provided Runtime. The logger is used to log any unhandled
// exceptions thrown by user-defined JavaScript functions.
func ParseFunc(javascript string, logger *log.Logger, rt *js.Runtime) (Func, error) {
	switch javascript {
	case "":
		return nil, nil
	case "_count":
		return Count, nil
	case "_sum":
		return Sum, nil
	case "_stats":
		return Stats, nil
	case "_approx_count_distinct":
		return ApproxCountDistinct, nil
	default:
		reduceFunc, err := rt.Reduce(javascript)
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context, keys [][2]any, values []any, rereduce bool) ([]any, error) {
			ret, err := reduceFunc(ctx, keys, values, rereduce)
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
