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

package mango

import (
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"
)

var cmpOpts = []cmp.Option{
	cmp.AllowUnexported(notSelector{}, combinationSelector{}, conditionSelector{}),
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
		want: &fieldSelector{
			field: "foo",
			cond: &conditionSelector{
				op:   OpEqual,
				cond: "bar",
			},
		},
	})
	tests.Add("explicit equality", test{
		input: `{"foo": {"$eq": "bar"}}`,
		want: &fieldSelector{
			field: "foo",
			cond: &conditionSelector{
				op:   OpEqual,
				cond: "bar",
			},
		},
	})
	tests.Add("explicit equality with too many object keys", test{
		input:   `{"foo": {"$eq": "bar", "$ne": "baz"}}`,
		wantErr: "too many keys in object",
	})
	tests.Add("implicit equality with empty object", test{
		input: `{"foo": {}}`,
		want: &fieldSelector{
			field: "foo",
			cond: &conditionSelector{
				op:   OpEqual,
				cond: map[string]interface{}{},
			},
		},
	})
	tests.Add("explicit invalid comparison operator", test{
		input:   `{"foo": {"$invalid": "bar"}}`,
		wantErr: "invalid operator $invalid",
	})
	tests.Add("explicit equiality against object", test{
		input: `{"foo": {"$eq": {"bar": "baz"}}}`,
		want: &fieldSelector{
			field: "foo",
			cond: &conditionSelector{
				op:   OpEqual,
				cond: map[string]interface{}{"bar": "baz"},
			},
		},
	})
	tests.Add("less than", test{
		input: `{"foo": {"$lt": 42}}`,
		want: &fieldSelector{
			field: "foo",
			cond: &conditionSelector{
				op:   OpLessThan,
				cond: float64(42),
			},
		},
	})
	tests.Add("less than or equal", test{
		input: `{"foo": {"$lte": 42}}`,
		want: &fieldSelector{
			field: "foo",
			cond: &conditionSelector{
				op:   OpLessThanOrEqual,
				cond: float64(42),
			},
		},
	})
	tests.Add("not equal", test{
		input: `{"foo": {"$ne": 42}}`,
		want: &fieldSelector{
			field: "foo",
			cond: &conditionSelector{
				op:   OpNotEqual,
				cond: float64(42),
			},
		},
	})
	tests.Add("greater than", test{
		input: `{"foo": {"$gt": 42}}`,
		want: &fieldSelector{
			field: "foo",
			cond: &conditionSelector{
				op:   OpGreaterThan,
				cond: float64(42),
			},
		},
	})
	tests.Add("greater than or equal", test{
		input: `{"foo": {"$gte": 42}}`,
		want: &fieldSelector{
			field: "foo",
			cond: &conditionSelector{
				op:   OpGreaterThanOrEqual,
				cond: float64(42),
			},
		},
	})
	tests.Add("exists", test{
		input: `{"foo": {"$exists": true}}`,
		want: &fieldSelector{
			field: "foo",
			cond: &conditionSelector{
				op:   OpExists,
				cond: true,
			},
		},
	})
	tests.Add("exists with non-boolean", test{
		input:   `{"foo": {"$exists": 42}}`,
		wantErr: "$exists: json: cannot unmarshal number into Go value of type bool",
	})
	tests.Add("type", test{
		input: `{"foo": {"$type": "string"}}`,
		want: &fieldSelector{
			field: "foo",
			cond: &conditionSelector{
				op:   OpType,
				cond: "string",
			},
		},
	})
	tests.Add("type with non-string", test{
		input:   `{"foo": {"$type": 42}}`,
		wantErr: "$type: json: cannot unmarshal number into Go value of type string",
	})
	tests.Add("in", test{
		input: `{"foo": {"$in": [1, 2, 3]}}`,
		want: &fieldSelector{
			field: "foo",
			cond: &conditionSelector{
				op:   OpIn,
				cond: []interface{}{float64(1), float64(2), float64(3)},
			},
		},
	})
	tests.Add("in with non-array", test{
		input:   `{"foo": {"$in": 42}}`,
		wantErr: "$in: json: cannot unmarshal number into Go value of type []interface {}",
	})
	tests.Add("not in", test{
		input: `{"foo": {"$nin": [1, 2, 3]}}`,
		want: &fieldSelector{
			field: "foo",
			cond: &conditionSelector{
				op:   OpNotIn,
				cond: []interface{}{float64(1), float64(2), float64(3)},
			},
		},
	})
	tests.Add("not in with non-array", test{
		input:   `{"foo": {"$nin": 42}}`,
		wantErr: "$nin: json: cannot unmarshal number into Go value of type []interface {}",
	})
	tests.Add("size", test{
		input: `{"foo": {"$size": 42}}`,
		want: &fieldSelector{
			field: "foo",
			cond: &conditionSelector{
				op:   OpSize,
				cond: float64(42),
			},
		},
	})
	tests.Add("size with non-integer", test{
		input:   `{"foo": {"$size": 42.5}}`,
		wantErr: "$size: json: cannot unmarshal number 42.5 into Go value of type uint",
	})
	tests.Add("mod", test{
		input: `{"foo": {"$mod": [2, 1]}}`,
		want: &fieldSelector{
			field: "foo",
			cond: &conditionSelector{
				op:   OpMod,
				cond: [2]int64{2, 1},
			},
		},
	})
	tests.Add("mod with non-array", test{
		input:   `{"foo": {"$mod": 42}}`,
		wantErr: "$mod: json: cannot unmarshal number into Go value of type [2]int64",
	})
	tests.Add("mod with zero divisor", test{
		input:   `{"foo": {"$mod": [0, 1]}}`,
		wantErr: "$mod: divisor must be non-zero",
	})
	tests.Add("regex", test{
		input: `{"foo": {"$regex": "^bar$"}}`,
		want: &fieldSelector{
			field: "foo",
			cond: &conditionSelector{
				op:   OpRegex,
				cond: regexp.MustCompile("^bar$"),
			},
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
				&fieldSelector{
					field: "baz",
					cond: &conditionSelector{
						op:   OpEqual,
						cond: "qux",
					},
				},
				&fieldSelector{
					field: "foo",
					cond: &conditionSelector{
						op:   OpEqual,
						cond: "bar",
					},
				},
			},
		},
	})
	tests.Add("explicit $and", test{
		input: `{"$and":[{"foo":"bar"},{"baz":"qux"}]}`,
		want: &combinationSelector{
			op: OpAnd,
			sel: []Selector{
				&fieldSelector{
					field: "foo",
					cond: &conditionSelector{
						op:   OpEqual,
						cond: "bar",
					},
				},
				&fieldSelector{
					field: "baz",
					cond: &conditionSelector{
						op:   OpEqual,
						cond: "qux",
					},
				},
			},
		},
	})
	tests.Add("nested implicit and explicit $and", test{
		input: `{"$and":[{"foo":"bar"},{"baz":"qux"}, {"quux":"corge","grault":"garply"}]}`,
		want: &combinationSelector{
			op: OpAnd,
			sel: []Selector{
				&fieldSelector{
					field: "foo",
					cond: &conditionSelector{
						op:   OpEqual,
						cond: "bar",
					},
				},
				&fieldSelector{
					field: "baz",
					cond: &conditionSelector{
						op:   OpEqual,
						cond: "qux",
					},
				},
				&combinationSelector{
					op: OpAnd,
					sel: []Selector{
						&fieldSelector{
							field: "grault",
							cond: &conditionSelector{
								op:   OpEqual,
								cond: "garply",
							},
						},
						&fieldSelector{
							field: "quux",
							cond: &conditionSelector{
								op:   OpEqual,
								cond: "corge",
							},
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
				&fieldSelector{
					field: "foo",
					cond: &conditionSelector{
						op:   OpEqual,
						cond: "bar",
					},
				},
				&fieldSelector{
					field: "baz",
					cond: &conditionSelector{
						op:   OpEqual,
						cond: "qux",
					},
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
		want: &notSelector{
			sel: &fieldSelector{
				field: "foo",
				cond: &conditionSelector{
					op:   OpEqual,
					cond: "bar",
				},
			},
		},
	})
	tests.Add("$not with invalid selector", test{
		input:   `{"$not": []}`,
		wantErr: "$not: json: cannot unmarshal array into Go value of type map[string]json.RawMessage",
	})
	tests.Add("$and with invalid selector array", test{
		input:   `{"$and": {}}`,
		wantErr: "$and: json: cannot unmarshal object into Go value of type []json.RawMessage",
	})
	tests.Add("$and with invalid selector", test{
		input:   `{"$and": [42]}`,
		wantErr: "$and: json: cannot unmarshal number into Go value of type map[string]json.RawMessage",
	})
	tests.Add("$nor", test{
		input: `{"$nor":[{"foo":"bar"},{"baz":"qux"}]}`,
		want: &combinationSelector{
			op: OpNor,
			sel: []Selector{
				&fieldSelector{
					field: "foo",
					cond: &conditionSelector{
						op:   OpEqual,
						cond: "bar",
					},
				},
				&fieldSelector{
					field: "baz",
					cond: &conditionSelector{
						op:   OpEqual,
						cond: "qux",
					},
				},
			},
		},
	})
	tests.Add("$all", test{
		input: `{"foo": {"$all": ["bar", "baz"]}}`,
		want: &fieldSelector{
			field: "foo",
			cond: &conditionSelector{
				op:   OpAll,
				cond: []interface{}{"bar", "baz"},
			},
		},
	})
	tests.Add("$all with non-array", test{
		input:   `{"foo": {"$all": "bar"}}`,
		wantErr: "$all: json: cannot unmarshal string into Go value of type []interface {}",
	})
	tests.Add("$elemMatch", test{
		input: `{"genre": {"$elemMatch": {"$eq": "Horror"}}}`,
		want: &fieldSelector{
			field: "genre",
			cond: &elementSelector{
				op: OpElemMatch,
				cond: &conditionSelector{
					op:   OpEqual,
					cond: "Horror",
				},
			},
		},
	})
	tests.Add("$allMatch", test{
		input: `{"genre": {"$allMatch": {"$eq": "Horror"}}}`,
		want: &fieldSelector{
			field: "genre",
			cond: &elementSelector{
				op: OpAllMatch,
				cond: &conditionSelector{
					op:   OpEqual,
					cond: "Horror",
				},
			},
		},
	})
	tests.Add("$keyMapMatch", test{
		input: `{"cameras": {"$keyMapMatch": {"$eq": "secondary"}}}`,
		want: &fieldSelector{
			field: "cameras",
			cond: &elementSelector{
				op: OpKeyMapMatch,
				cond: &conditionSelector{
					op:   OpEqual,
					cond: "secondary",
				},
			},
		},
	})
	tests.Add("element selector with invalid selector", test{
		input:   `{"cameras": {"$keyMapMatch": 42}}`,
		wantErr: "$keyMapMatch: json: cannot unmarshal number into Go value of type map[string]json.RawMessage",
	})

	/*
		TODO:
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
