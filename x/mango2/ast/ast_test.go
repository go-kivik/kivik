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
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"
)

var cmpOpts = []cmp.Option{
	cmp.AllowUnexported(unarySelector{}, combinationSelector{}, conditionSelector{}),
}

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
	tests.Add("mod", test{
		input: `{"foo": {"$mod": [2, 1]}}`,
		want: &conditionSelector{
			field: "foo",
			op:    OpMod,
			value: [2]int{2, 1},
		},
	})
	tests.Add("mod with non-array", test{
		input:   `{"foo": {"$mod": 42}}`,
		wantErr: "$mod: json: cannot unmarshal number into Go value of type [2]int",
	})
	tests.Add("mod with zero divisor", test{
		input:   `{"foo": {"$mod": [0, 1]}}`,
		wantErr: "$mod: divisor must be non-zero",
	})
	tests.Add("regex", test{
		input: `{"foo": {"$regex": "^bar$"}}`,
		want: &conditionSelector{
			field: "foo",
			op:    OpRegex,
			value: regexp.MustCompile("^bar$"),
		},
	})
	tests.Add("regexp non-string", test{
		input:   `{"foo": {"$regex": 42}}`,
		wantErr: "$regex: json: cannot unmarshal number into Go value of type string",
	})
	tests.Add("regexp invalid", test{
		input:   `{"foo": {"$regex": "["}}`,
		wantErr: "$regex: error parsing regexp: missing closing ]: `[`",
	})
	tests.Add("implicit $and", test{
		input: `{"foo":"bar","baz":"qux"}`,
		want: &combinationSelector{
			op: OpAnd,
			sel: []Selector{
				&conditionSelector{
					field: "baz",
					op:    OpEqual,
					value: "qux",
				},
				&conditionSelector{
					field: "foo",
					op:    OpEqual,
					value: "bar",
				},
			},
		},
	})
	tests.Add("explicit $and", test{
		input: `{"$and":[{"foo":"bar"},{"baz":"qux"}]}`,
		want: &combinationSelector{
			op: OpAnd,
			sel: []Selector{
				&conditionSelector{
					field: "foo",
					op:    OpEqual,
					value: "bar",
				},
				&conditionSelector{
					field: "baz",
					op:    OpEqual,
					value: "qux",
				},
			},
		},
	})
	tests.Add("nested implicit and explicit $and", test{
		input: `{"$and":[{"foo":"bar"},{"baz":"qux"}, {"quux":"corge","grault":"garply"}]}`,
		want: &combinationSelector{
			op: OpAnd,
			sel: []Selector{
				&conditionSelector{
					field: "foo",
					op:    OpEqual,
					value: "bar",
				},
				&conditionSelector{
					field: "baz",
					op:    OpEqual,
					value: "qux",
				},
				&combinationSelector{
					op: OpAnd,
					sel: []Selector{
						&conditionSelector{
							field: "grault",
							op:    OpEqual,
							value: "garply",
						},
						&conditionSelector{
							field: "quux",
							op:    OpEqual,
							value: "corge",
						},
					},
				},
			},
		},
	})
	tests.Add("$or", test{
		input: `{"$or":[{"foo":"bar"},{"baz":"qux"}]}`,
		want: &combinationSelector{
			op: OpOr,
			sel: []Selector{
				&conditionSelector{
					field: "foo",
					op:    OpEqual,
					value: "bar",
				},
				&conditionSelector{
					field: "baz",
					op:    OpEqual,
					value: "qux",
				},
			},
		},
	})
	tests.Add("invalid operator", test{
		input:   `{"$invalid": "bar"}`,
		wantErr: "unknown operator $invalid",
	})
	tests.Add("$not", test{
		input: `{"$not": {"foo":"bar"}}`,
		want: &unarySelector{
			op: OpNot,
			sel: &conditionSelector{
				field: "foo",
				op:    OpEqual,
				value: "bar",
			},
		},
	})

	/*
		TODO:
		- $nor
		- $all
		- $elemMatch
		- $allMatch
		- $keyMapMatch

		- $mod with non-integer values returns 404 (WTF) https://docs.couchdb.org/en/stable/api/database/find.html#condition-operators

	*/

	tests.Run(t, func(t *testing.T, tt test) {
		got, err := Parse([]byte(tt.input))
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Fatalf("Unexpected error: %s", err)
		}
		if err != nil {
			return
		}
		if d := cmp.Diff(tt.want.String(), got.String(), cmpOpts...); d != "" {
			t.Errorf("Unexpected result (-want +got):\n%s", d)
		}
	})
}
