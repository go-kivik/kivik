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

	"github.com/go-kivik/kivik/v4"
)

func TestReduce(t *testing.T) {
	type test struct {
		input      RowIterator
		fn         Func
		groupLevel int
		want       []Row
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
			{ID: "1", Key: "foo", Value: nil, First: 1, Last: 1},
		},
		fn: Count,
		want: []Row{
			{Value: 1.0, First: 1, Last: 1},
		},
	})
	tests.Add("count two rows", test{
		input: &Rows{
			{ID: "1", Key: "foo", Value: nil, First: 1, Last: 1},
			{ID: "2", Key: "foo", Value: nil, First: 2, Last: 2},
		},
		fn: Count,
		want: []Row{
			{Value: 2.0, First: 1, Last: 2},
		},
	})
	tests.Add("max group_level", test{
		input: &Rows{
			{ID: "a", Key: []any{1.0, 2.0, 3.0}, Value: nil, First: 1, Last: 1},
			{ID: "b", Key: []any{1.0, 2.0, 3.0}, Value: nil, First: 2, Last: 2},
			{ID: "c", Key: []any{1.0, 2.0, 4.0}, Value: nil, First: 3, Last: 3},
			{ID: "d", Key: []any{1.0, 2.0, 5.0}, Value: nil, First: 4, Last: 4},
		},
		groupLevel: -1,
		fn:         Count,
		want: []Row{
			{Key: []any{1.0, 2.0, 3.0}, Value: 2.0, First: 1, Last: 2},
			{Key: []any{1.0, 2.0, 4.0}, Value: 1.0, First: 3, Last: 3},
			{Key: []any{1.0, 2.0, 5.0}, Value: 1.0, First: 4, Last: 4},
		},
	})
	tests.Add("max group_level with mixed depth keys", test{
		input: &Rows{
			{ID: "a", Key: []any{1.0, 2.0, 3.0, 4.0, 5.0}, Value: nil, First: 1, Last: 1},
			{ID: "b", Key: []any{1.0, 2.0, 3.0}, Value: nil, First: 2, Last: 2},
			{ID: "c", Key: []any{1.0, 2.0, 4.0}, Value: nil, First: 3, Last: 3},
			{ID: "d", Key: []any{1.0, 2.0, 5.0}, Value: nil, First: 4, Last: 4},
		},
		groupLevel: -1,
		fn:         Count,
		want: []Row{
			{Key: []any{1.0, 2.0, 3.0, 4.0, 5.0}, Value: 1.0, First: 1, Last: 1},
			{Key: []any{1.0, 2.0, 3.0}, Value: 1.0, First: 2, Last: 2},
			{Key: []any{1.0, 2.0, 4.0}, Value: 1.0, First: 3, Last: 3},
			{Key: []any{1.0, 2.0, 5.0}, Value: 1.0, First: 4, Last: 4},
		},
	})
	tests.Add("explicit group_level with mixed-depth keys", test{
		input: &Rows{
			{ID: "a", Key: []any{1.0, 2.0, 3.0, 4.0, 5.0}, Value: nil, First: 1, Last: 1},
			{ID: "b", Key: []any{1.0, 2.0, 3.0}, Value: nil, First: 2, Last: 2},
			{ID: "c", Key: []any{1.0, 2.0, 4.0}, Value: nil, First: 3, Last: 3},
			{ID: "d", Key: []any{1.0, 2.0, 5.0}, Value: nil, First: 4, Last: 4},
		},
		groupLevel: 3,
		fn:         Count,
		want: []Row{
			{Key: []any{1.0, 2.0, 3.0}, Value: 2.0, First: 1, Last: 2},
			{Key: []any{1.0, 2.0, 4.0}, Value: 1.0, First: 3, Last: 3},
			{Key: []any{1.0, 2.0, 5.0}, Value: 1.0, First: 4, Last: 4},
		},
	})
	tests.Add("mix map and pre-reduced inputs", test{
		input: &Rows{
			{Key: []any{1.0, 2.0, 3.0, 4.0}, Value: 3.0, First: 1, Last: 3},
			{Key: []any{1.0, 2.0, 3.0, 6.0}, Value: 1.0, First: 4, Last: 4},
			{ID: "b", Key: []any{1.0, 2.0, 3.0}, Value: nil, First: 5, Last: 5},
			{ID: "c", Key: []any{1.0, 2.0, 4.0}, Value: nil, First: 6, Last: 6},
		},
		groupLevel: 3,
		fn:         Count,
		want: []Row{
			{Key: []any{1.0, 2.0, 3.0}, Value: 5.0, First: 1, Last: 5},
			{Key: []any{1.0, 2.0, 4.0}, Value: 1.0, First: 6, Last: 6},
		},
		wantCache: [][]Row{
			{{Key: []any{1.0, 2.0, 3.0}, Value: 4.0, First: 1, Last: 4}}, // Merge of first two rows, rereduce=true
			{{Key: []any{1.0, 2.0, 3.0}, Value: 1.0, First: 5, Last: 5}}, // Single map row, rereduce=false
			{{Key: []any{1.0, 2.0, 4.0}, Value: 1.0, First: 6, Last: 6}}, // Single map row, rereduce=false, final
			{{Key: []any{1.0, 2.0, 3.0}, Value: 5.0, First: 1, Last: 5}}, // Merge of the first two reduce outputs, final
		},
	})

	tests.Run(t, func(t *testing.T, tt test) {
		var cache [][]Row
		cb := func(rows []Row) {
			cache = append(cache, rows)
		}
		got, err := Reduce(tt.input, tt.fn, tt.groupLevel, cb)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("Unexpected error: %v", err)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status code: %d", status)
		}
		if err != nil {
			return
		}
		if d := cmp.Diff(tt.want, got); d != "" {
			t.Errorf("Unexpected output (-want +got):\n%s", d)
		}
		if tt.wantCache != nil {
			if d := cmp.Diff(tt.wantCache, cache); d != "" {
				t.Errorf("Unexpected cache (-want +got):\n%s", d)
			}
		}
	})
}
