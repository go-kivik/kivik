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
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"
)

func TestStats(t *testing.T) {
	t.Parallel()

	type test struct {
		values   []any
		rereduce bool
		want     []any
		wantErr  string
	}

	tests := testy.NewTable()

	tests.Add("single float value", test{
		values: []any{5.0},
		want: []any{stats{
			Sum:    5,
			Min:    5,
			Max:    5,
			Count:  1,
			SumSqr: 25,
		}},
	})
	tests.Add("multiple float values", test{
		values: []any{1.0, 2.0, 3.0, 4.0, 5.0},
		want: []any{stats{
			Sum:    15,
			Min:    1,
			Max:    5,
			Count:  5,
			SumSqr: 55,
		}},
	})
	tests.Add("single zero value", test{
		values: []any{0.0},
		want: []any{stats{
			Sum:    0,
			Min:    0,
			Max:    0,
			Count:  1,
			SumSqr: 0,
		}},
	})
	tests.Add("negative values", test{
		values: []any{-3.0, -1.0, 2.0},
		want: []any{stats{
			Sum:    -2,
			Min:    -3,
			Max:    2,
			Count:  3,
			SumSqr: 14,
		}},
	})
	tests.Add("null value", test{
		values:  []any{nil},
		wantErr: `the _stats function requires that map values be numbers or arrays of numbers, not 'null'`,
	})
	tests.Add("string value", test{
		values:  []any{"hello"},
		wantErr: `the _stats function requires that map values be numbers or arrays of numbers, not '"hello"'`,
	})
	tests.Add("boolean value", test{
		values:  []any{true},
		wantErr: `the _stats function requires that map values be numbers or arrays of numbers, not 'true'`,
	})
	tests.Add("mixed float and null", test{
		values:  []any{1.0, nil},
		wantErr: `the _stats function requires that map values be numbers or arrays of numbers, not 'null'`,
	})
	tests.Add("arrays of floats", test{
		values: []any{
			[]any{1.0, 10.0},
			[]any{2.0, 20.0},
			[]any{3.0, 30.0},
		},
		want: []any{[]stats{
			{Sum: 6, Min: 1, Max: 3, Count: 3, SumSqr: 14},
			{Sum: 60, Min: 10, Max: 30, Count: 3, SumSqr: 1400},
		}},
	})
	tests.Add("single array of floats", test{
		values: []any{
			[]any{5.0, 15.0},
		},
		want: []any{[]stats{
			{Sum: 5, Min: 5, Max: 5, Count: 1, SumSqr: 25},
			{Sum: 15, Min: 15, Max: 15, Count: 1, SumSqr: 225},
		}},
	})
	tests.Add("array with non-float element", test{
		values: []any{
			[]any{1.0, "bad"},
		},
		// Falls through to non-array path, fails mapstructure decode
		wantErr: `the _stats function requires that map values be numbers or arrays of numbers, not '\[1,"bad"\]'`,
	})
	tests.Add("pre-aggregated stats object", test{
		values: []any{
			map[string]any{
				"sum":    10.0,
				"min":    1.0,
				"max":    5.0,
				"count":  3.0,
				"sumsqr": 40.0,
			},
		},
		want: []any{stats{
			Sum:    10,
			Min:    1,
			Max:    5,
			Count:  3, // TODO: Count is 3 from map stats, but then decremented by 1 to 2, then added to len(values)=1, giving 3. Confusing logic.
			SumSqr: 40,
		}},
	})
	tests.Add("mixed floats and pre-aggregated stats", test{
		values: []any{
			2.0,
			map[string]any{
				"sum":    10.0,
				"min":    1.0,
				"max":    5.0,
				"count":  3.0,
				"sumsqr": 40.0,
			},
		},
		want: []any{stats{
			Sum:    12,
			Min:    1,
			Max:    5,
			Count:  4, // len(values)=2, plus mapStats.Count=3, minus 1 for double-count = 4
			SumSqr: 44,
		}},
	})
	tests.Add("pre-aggregated stats missing field", test{
		values: []any{
			map[string]any{
				"sum": 10.0,
				"min": 1.0,
				"max": 5.0,
			},
		},
		wantErr: `user _stats input missing required field count`,
	})
	tests.Add("rereduce single stats", test{
		values: []any{
			stats{Sum: 10, Min: 1, Max: 5, Count: 3, SumSqr: 40},
		},
		rereduce: true,
		want: []any{stats{
			Sum:    10,
			Min:    1,
			Max:    5,
			Count:  3,
			SumSqr: 40,
		}},
	})
	tests.Add("rereduce multiple stats", test{
		values: []any{
			stats{Sum: 10, Min: 1, Max: 5, Count: 3, SumSqr: 40},
			stats{Sum: 20, Min: 2, Max: 8, Count: 4, SumSqr: 120},
		},
		rereduce: true,
		want: []any{stats{
			Sum:    30,
			Min:    1,
			Max:    8,
			Count:  7,
			SumSqr: 160,
		}},
	})
	tests.Add("rereduce stats arrays", test{
		values: []any{
			[]stats{
				{Sum: 6, Min: 1, Max: 3, Count: 3, SumSqr: 14},
				{Sum: 60, Min: 10, Max: 30, Count: 3, SumSqr: 1400},
			},
			[]stats{
				{Sum: 4, Min: 2, Max: 2, Count: 2, SumSqr: 8},
				{Sum: 40, Min: 20, Max: 20, Count: 2, SumSqr: 800},
			},
		},
		rereduce: true,
		want: []any{[]stats{
			{Sum: 10, Min: 1, Max: 3, Count: 5, SumSqr: 22},
			{Sum: 100, Min: 10, Max: 30, Count: 5, SumSqr: 2200},
		}},
	})

	tests.Add("differing-length float arrays", test{
		values: []any{
			[]any{1.0, 2.0},
			[]any{3.0},
		},
		wantErr: `the _stats function requires that map values be arrays of the same length`,
	})

	// TODO: Test empty values slice. slices.Min/Max panic on empty slices,
	// so Stats(nil, []any{}, false) would panic. Verify against CouchDB.

	tests.Run(t, func(t *testing.T, tt test) {
		got, err := Stats(nil, tt.values, tt.rereduce)
		if !testy.ErrorMatchesRE(tt.wantErr, err) {
			t.Errorf("unexpected error: got %v, want /%s/", err, tt.wantErr)
		}
		if err != nil {
			return
		}
		if d := cmp.Diff(tt.want, got, cmp.AllowUnexported(stats{})); d != "" {
			t.Errorf("unexpected result (-want +got):\n%s", d)
		}
	})
}
