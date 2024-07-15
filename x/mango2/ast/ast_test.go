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

package ast

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"
)

var cmpOpts = cmp.AllowUnexported(unarySelector{}, combinationSelector{}, conditionSelector{})

func TestParse(t *testing.T) {
	type test struct {
		input   string
		want    Selector
		wantErr string
	}

	tests := testy.NewTable()
	tests.Add("empty", test{
		input: "{}",
		want: &combinationSelector{
			op:  OpAnd,
			sel: nil,
		},
	})
	tests.Add("implicit equality", test{
		input: `{"foo": "bar"}`,
		want: &conditionSelector{
			field: "foo",
			op:    OpEqual,
			value: "bar",
		},
	})
	tests.Add("explicit equality", test{
		input: `{"foo": {"$eq": "bar"}}`,
		want: &conditionSelector{
			field: "foo",
			op:    OpEqual,
			value: "bar",
		},
	})
	tests.Add("explicit equality with too many object keys", test{
		input:   `{"foo": {"$eq": "bar", "$ne": "baz"}}`,
		wantErr: "too many keys in object",
	})
	tests.Add("implicit equality with empty object", test{
		input: `{"foo": {}}`,
		want: &conditionSelector{
			field: "foo",
			op:    OpEqual,
			value: map[string]interface{}{},
		},
	})
	tests.Add("explicit invalid comparison operator", test{
		input:   `{"foo": {"$invalid": "bar"}}`,
		wantErr: "invalid operator $invalid",
	})
	tests.Add("explicit equiality against object", test{
		input: `{"foo": {"$eq": {"bar": "baz"}}}`,
		want: &conditionSelector{
			field: "foo",
			op:    OpEqual,
			value: map[string]interface{}{"bar": "baz"},
		},
	})
	tests.Add("less than", test{
		input: `{"foo": {"$lt": 42}}`,
		want: &conditionSelector{
			field: "foo",
			op:    OpLessThan,
			value: float64(42),
		},
	})
	tests.Add("less than or equal", test{
		input: `{"foo": {"$lte": 42}}`,
		want: &conditionSelector{
			field: "foo",
			op:    OpLessThanOrEqual,
			value: float64(42),
		},
	})
	tests.Add("not equal", test{
		input: `{"foo": {"$ne": 42}}`,
		want: &conditionSelector{
			field: "foo",
			op:    OpNotEqual,
			value: float64(42),
		},
	})
	tests.Add("greater than", test{
		input: `{"foo": {"$gt": 42}}`,
		want: &conditionSelector{
			field: "foo",
			op:    OpGreaterThan,
			value: float64(42),
		},
	})
	tests.Add("greater than or equal", test{
		input: `{"foo": {"$gte": 42}}`,
		want: &conditionSelector{
			field: "foo",
			op:    OpGreaterThanOrEqual,
			value: float64(42),
		},
	})
	tests.Add("exists", test{
		input: `{"foo": {"$exists": true}}`,
		want: &conditionSelector{
			field: "foo",
			op:    OpExists,
			value: true,
		},
	})
	tests.Add("exists with non-boolean", test{
		input:   `{"foo": {"$exists": 42}}`,
		wantErr: "$exists: json: cannot unmarshal number into Go value of type bool",
	})
	tests.Add("type", test{
		input: `{"foo": {"$type": "string"}}`,
		want: &conditionSelector{
			field: "foo",
			op:    OpType,
			value: "string",
		},
	})
	tests.Add("type with non-string", test{
		input:   `{"foo": {"$type": 42}}`,
		wantErr: "$type: json: cannot unmarshal number into Go value of type string",
	})
	tests.Add("in", test{
		input: `{"foo": {"$in": [1, 2, 3]}}`,
		want: &conditionSelector{
			field: "foo",
			op:    OpIn,
			value: []interface{}{float64(1), float64(2), float64(3)},
		},
	})
	tests.Add("in with non-array", test{
		input:   `{"foo": {"$in": 42}}`,
		wantErr: "$in: json: cannot unmarshal number into Go value of type []interface {}",
	})
	tests.Add("not in", test{
		input: `{"foo": {"$nin": [1, 2, 3]}}`,
		want: &conditionSelector{
			field: "foo",
			op:    OpNotIn,
			value: []interface{}{float64(1), float64(2), float64(3)},
		},
	})
	tests.Add("not in with non-array", test{
		input:   `{"foo": {"$nin": 42}}`,
		wantErr: "$nin: json: cannot unmarshal number into Go value of type []interface {}",
	})
	tests.Add("size", test{
		input: `{"foo": {"$size": 42}}`,
		want: &conditionSelector{
			field: "foo",
			op:    OpSize,
			value: float64(42),
		},
	})
	tests.Add("size with non-integer", test{
		input:   `{"foo": {"$size": 42.5}}`,
		wantErr: "$size: json: cannot unmarshal number 42.5 into Go value of type uint",
	})

	/*
		TODO:
		- $mod
		- $regex
		- implicit $and
		- $and
		- $or
		- $not
		- $nor
		- $all
		- $elemMatch
		- $allMatch
		- $keyMapMatch


	*/

	tests.Run(t, func(t *testing.T, tt test) {
		got, err := Parse([]byte(tt.input))
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Fatalf("Unexpected error: %s", err)
		}
		if d := cmp.Diff(tt.want, got, cmpOpts); d != "" {
			t.Errorf("Unexpected result (-want +got):\n%s", d)
		}
	})
}
