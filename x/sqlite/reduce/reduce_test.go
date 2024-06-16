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
	"io"
	"log"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
)

func TestReduce(t *testing.T) {
	type test struct {
		input      Reducer
		javascript string
		groupLevel int
		batchSize  int
		want       Rows
		wantCache  [][]Row
		wantErr    string
		wantStatus int
	}

	tests := testy.NewTable()
	tests.Add("no inputs", test{
		input: &Rows{},
		want:  []Row{},
	})
	tests.Add("count single row", test{
		input: &Rows{
			{ID: "1", FirstKey: "foo", Value: nil, FirstPK: 1, LastPK: 1},
		},
		javascript: "_count",
		want: []Row{
			{Value: 1.0, FirstPK: 1, LastPK: 1},
		},
	})
	tests.Add("non-array key with grouping", test{
		input: &Rows{
			{ID: "1", FirstKey: "foo", Value: nil, FirstPK: 1, LastPK: 1},
		},
		javascript: "_count",
		groupLevel: -1,
		want: []Row{
			{FirstKey: "foo", Value: 1.0, FirstPK: 1, LastPK: 1},
		},
	})
	tests.Add("single-element array key with grouping", test{
		input: &Rows{
			{ID: "1", FirstKey: []any{"foo"}, Value: nil, FirstPK: 1, LastPK: 1},
		},
		javascript: "_count",
		groupLevel: -1,
		want: []Row{
			{FirstKey: []any{"foo"}, Value: 1.0, FirstPK: 1, LastPK: 1},
		},
	})
	tests.Add("count two rows", test{
		input: &Rows{
			{ID: "1", FirstKey: "foo", Value: nil, FirstPK: 1, LastPK: 1},
			{ID: "2", FirstKey: "foo", Value: nil, FirstPK: 2, LastPK: 2},
		},
		javascript: "_count",
		want: []Row{
			{Value: 2.0, FirstPK: 1, LastPK: 2},
		},
	})
	tests.Add("max group_level", test{
		input: &Rows{
			{ID: "a", FirstKey: []any{1.0, 2.0, 3.0}, Value: nil, FirstPK: 1, LastPK: 1},
			{ID: "b", FirstKey: []any{1.0, 2.0, 3.0}, Value: nil, FirstPK: 2, LastPK: 2},
			{ID: "c", FirstKey: []any{1.0, 2.0, 4.0}, Value: nil, FirstPK: 3, LastPK: 3},
			{ID: "d", FirstKey: []any{1.0, 2.0, 5.0}, Value: nil, FirstPK: 4, LastPK: 4},
		},
		groupLevel: -1,
		javascript: "_count",
		want: []Row{
			{FirstKey: []any{1.0, 2.0, 3.0}, Value: 2.0, FirstPK: 1, LastPK: 2},
			{FirstKey: []any{1.0, 2.0, 4.0}, Value: 1.0, FirstPK: 3, LastPK: 3},
			{FirstKey: []any{1.0, 2.0, 5.0}, Value: 1.0, FirstPK: 4, LastPK: 4},
		},
	})
	tests.Add("max group_level with mixed depth keys", test{
		input: &Rows{
			{ID: "a", FirstKey: []any{1.0, 2.0, 3.0, 4.0, 5.0}, Value: nil, FirstPK: 1, LastPK: 1},
			{ID: "b", FirstKey: []any{1.0, 2.0, 3.0}, Value: nil, FirstPK: 2, LastPK: 2},
			{ID: "c", FirstKey: []any{1.0, 2.0, 4.0}, Value: nil, FirstPK: 3, LastPK: 3},
			{ID: "d", FirstKey: []any{1.0, 2.0, 5.0}, Value: nil, FirstPK: 4, LastPK: 4},
		},
		groupLevel: -1,
		javascript: "_count",
		want: []Row{
			{FirstKey: []any{1.0, 2.0, 3.0, 4.0, 5.0}, Value: 1.0, FirstPK: 1, LastPK: 1},
			{FirstKey: []any{1.0, 2.0, 3.0}, Value: 1.0, FirstPK: 2, LastPK: 2},
			{FirstKey: []any{1.0, 2.0, 4.0}, Value: 1.0, FirstPK: 3, LastPK: 3},
			{FirstKey: []any{1.0, 2.0, 5.0}, Value: 1.0, FirstPK: 4, LastPK: 4},
		},
	})
	tests.Add("explicit group_level with mixed-depth keys", test{
		input: &Rows{
			{ID: "a", FirstKey: []any{1.0, 2.0, 3.0, 4.0, 5.0}, Value: nil, FirstPK: 1, LastPK: 1},
			{ID: "b", FirstKey: []any{1.0, 2.0, 3.0}, Value: nil, FirstPK: 2, LastPK: 2},
			{ID: "c", FirstKey: []any{1.0, 2.0, 4.0}, Value: nil, FirstPK: 3, LastPK: 3},
			{ID: "d", FirstKey: []any{1.0, 2.0, 5.0}, Value: nil, FirstPK: 4, LastPK: 4},
		},
		groupLevel: 3,
		javascript: "_count",
		want: []Row{
			{FirstKey: []any{1.0, 2.0, 3.0}, Value: 2.0, FirstPK: 1, LastPK: 2},
			{FirstKey: []any{1.0, 2.0, 4.0}, Value: 1.0, FirstPK: 3, LastPK: 3},
			{FirstKey: []any{1.0, 2.0, 5.0}, Value: 1.0, FirstPK: 4, LastPK: 4},
		},
	})
	tests.Add("mix map and pre-reduced inputs", test{
		input: &Rows{
			{FirstKey: []any{1.0, 2.0, 3.0, 4.0}, Value: 3.0, FirstPK: 1, LastPK: 3},
			{FirstKey: []any{1.0, 2.0, 3.0, 6.0}, Value: 1.0, FirstPK: 4, LastPK: 4},
			{ID: "b", FirstKey: []any{1.0, 2.0, 3.0}, Value: nil, FirstPK: 5, LastPK: 5},
			{ID: "c", FirstKey: []any{1.0, 2.0, 4.0}, Value: nil, FirstPK: 6, LastPK: 6},
		},
		groupLevel: 3,
		javascript: "_count",
		want: []Row{
			{FirstKey: []any{1.0, 2.0, 3.0}, Value: 5.0, FirstPK: 1, LastPK: 5},
			{FirstKey: []any{1.0, 2.0, 4.0}, Value: 1.0, FirstPK: 6, LastPK: 6},
		},
		wantCache: [][]Row{
			{{FirstKey: []any{1.0, 2.0, 3.0}, Value: 4.0, FirstPK: 1, LastPK: 4}}, // Merge of first two rows, rereduce=true
			{{FirstKey: []any{1.0, 2.0, 3.0}, Value: 1.0, FirstPK: 5, LastPK: 5}}, // Single map row, rereduce=false
			{{FirstKey: []any{1.0, 2.0, 4.0}, Value: 1.0, FirstPK: 6, LastPK: 6}}, // Single map row, rereduce=false, final
			{{FirstKey: []any{1.0, 2.0, 3.0}, Value: 5.0, FirstPK: 1, LastPK: 5}}, // Merge of the first two reduce outputs, final
		},
	})
	tests.Add("group level 0", test{
		input: &Rows{
			{ID: "a", FirstKey: []any{1.0, 2.0, 3.0}, Value: nil, FirstPK: 1, LastPK: 1},
			{ID: "b", FirstKey: []any{1.0, 2.0, 3.0}, Value: nil, FirstPK: 2, LastPK: 2},
			{ID: "c", FirstKey: []any{1.0, 2.0, 4.0}, Value: nil, FirstPK: 3, LastPK: 3},
			{ID: "d", FirstKey: []any{1.0, 2.0, 5.0}, Value: nil, FirstPK: 4, LastPK: 4},
		},
		groupLevel: 0,
		javascript: "_count",
		want: []Row{
			{FirstKey: nil, Value: 4.0, FirstPK: 1, LastPK: 4},
		},
	})
	tests.Add("group level 0, cached", test{
		input: &Rows{
			{FirstKey: nil, Value: 4.0, FirstPK: 1, LastPK: 4},
		},
		groupLevel: 0,
		javascript: "_count",
		want: []Row{
			{FirstKey: nil, Value: 4.0, FirstPK: 1, LastPK: 4},
		},
	})
	tests.Add("group level 0, partially cached", test{
		input: &Rows{
			{FirstKey: nil, Value: 4.0, FirstPK: 1, LastPK: 4},
			{ID: "e", FirstKey: []any{1.0, 2.0, 4.0}, Value: nil, FirstPK: 5, LastPK: 5},
			{ID: "f", FirstKey: []any{1.0, 2.0, 5.0}, Value: nil, FirstPK: 6, LastPK: 6},
		},
		groupLevel: 0,
		javascript: "_count",
		want: []Row{
			{FirstKey: nil, Value: 6.0, FirstPK: 1, LastPK: 6},
		},
	})
	tests.Add("batched reduce", test{
		input: &Rows{
			{ID: "a", FirstKey: []any{1.0, 2.0, 4.0}, Value: nil, FirstPK: 1},
			{ID: "b", FirstKey: []any{1.0, 2.0, 5.0}, Value: nil, FirstPK: 2},
			{ID: "c", FirstKey: []any{1.0, 2.0, 4.0}, Value: nil, FirstPK: 3},
			{ID: "d", FirstKey: []any{1.0, 2.0, 5.0}, Value: nil, FirstPK: 4},
			{ID: "e", FirstKey: []any{1.0, 2.0, 4.0}, Value: nil, FirstPK: 5},
			{ID: "f", FirstKey: []any{1.0, 2.0, 5.0}, Value: nil, FirstPK: 6},
		},
		groupLevel: 0,
		javascript: "_count",
		batchSize:  2,
		want: []Row{
			{FirstKey: nil, Value: 6.0, FirstPK: 1, LastPK: 6},
		},
		wantCache: [][]Row{
			{{FirstKey: nil, Value: 2.0, FirstPK: 1, LastPK: 2}},
			{{FirstKey: nil, Value: 2.0, FirstPK: 3, LastPK: 4}},
			{{FirstKey: nil, Value: 2.0, FirstPK: 5, LastPK: 6}},
			{{FirstKey: nil, Value: 4.0, FirstPK: 1, LastPK: 4}},
			{{FirstKey: nil, Value: 6.0, FirstPK: 1, LastPK: 6}},
		},
	})

	tests.Run(t, func(t *testing.T, tt test) {
		var cache [][]Row
		cb := func(_ uint, rows []Row) {
			cache = append(cache, rows)
		}
		batchSize := tt.batchSize
		if batchSize == 0 {
			batchSize = defaultBatchSize
		}
		got, err := reduceWithBatchSize(tt.input, tt.javascript, log.New(io.Discard, "", 0), tt.groupLevel, cb, batchSize)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("Unexpected error: %v", err)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status code: %d", status)
		}
		if err != nil {
			return
		}
		if d := cmp.Diff(tt.want, *got); d != "" {
			t.Errorf("Unexpected output (-want +got):\n%s", d)
		}
		if tt.wantCache != nil {
			if d := cmp.Diff(tt.wantCache, cache); d != "" {
				t.Errorf("Unexpected cache (-want +got):\n%s", d)
			}
		}
	})
}
