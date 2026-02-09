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

	"gitlab.com/flimzy/testy"
)

func TestMatch(t *testing.T) {
	type test struct {
		sel  Node
		doc  any
		want bool
	}

	tests := testy.NewTable()
	tests.Add("nil selector", test{
		sel:  nil,
		doc:  "foo",
		want: true,
	})
	tests.Add("equality", test{
		sel: &conditionNode{
			op:   OpEqual,
			cond: "foo",
		},
		doc:  "foo",
		want: true,
	})
	tests.Add("!equality", test{
		sel: &conditionNode{
			op:   OpEqual,
			cond: "foo",
		},
		doc:  "bar",
		want: false,
	})
	tests.Add("inequality", test{
		sel: &conditionNode{
			op:   OpNotEqual,
			cond: "foo",
		},
		doc:  "bar",
		want: true,
	})
	tests.Add("!inequality", test{
		sel: &conditionNode{
			op:   OpNotEqual,
			cond: "foo",
		},
		doc:  "foo",
		want: false,
	})
	tests.Add("less than", test{
		sel: &conditionNode{
			op:   OpLessThan,
			cond: float64(5),
		},
		doc:  float64(4),
		want: true,
	})
	tests.Add("!less than", test{
		sel: &conditionNode{
			op:   OpLessThan,
			cond: float64(5),
		},
		doc:  float64(10),
		want: false,
	})
	tests.Add("less than or equal", test{
		sel: &conditionNode{
			op:   OpLessThanOrEqual,
			cond: float64(5),
		},
		doc:  float64(5),
		want: true,
	})
	tests.Add("!less than or equal", test{
		sel: &conditionNode{
			op:   OpLessThanOrEqual,
			cond: float64(5),
		},
		doc:  float64(8),
		want: false,
	})
	tests.Add("greater than", test{
		sel: &conditionNode{
			op:   OpGreaterThan,
			cond: float64(5),
		},
		doc:  float64(10),
		want: true,
	})
	tests.Add("!greater than", test{
		sel: &conditionNode{
			op:   OpGreaterThan,
			cond: float64(5),
		},
		doc:  float64(2),
		want: false,
	})
	tests.Add("greater than or equal", test{
		sel: &conditionNode{
			op:   OpGreaterThanOrEqual,
			cond: float64(5),
		},
		doc:  float64(5),
		want: true,
	})
	tests.Add("!greater than or equal", test{
		sel: &conditionNode{
			op:   OpGreaterThanOrEqual,
			cond: float64(5),
		},
		doc:  float64(2),
		want: false,
	})
	tests.Add("exists", test{
		sel: &fieldNode{
			field: "foo",
			cond:  &conditionNode{op: OpExists, cond: true},
		},
		doc: map[string]any{
			"foo": "bar",
		},
		want: true,
	})
	tests.Add("!exists", test{
		sel: &fieldNode{
			field: "baz",
			cond:  &conditionNode{op: OpExists, cond: true},
		},
		doc: map[string]any{
			"foo": "bar",
		},
		want: false,
	})
	tests.Add("not exists", test{
		sel: &fieldNode{
			field: "baz",
			cond:  &conditionNode{op: OpExists, cond: false},
		},
		doc: map[string]any{
			"foo": "bar",
		},
		want: true,
	})
	tests.Add("!not exists", test{
		sel: &fieldNode{
			field: "baz",
			cond:  &conditionNode{op: OpExists, cond: true},
		},
		doc: map[string]any{
			"foo": "bar",
		},
		want: false,
	})
	tests.Add("type, null", test{
		sel: &conditionNode{
			op:   OpType,
			cond: "null",
		},
		doc:  nil,
		want: true,
	})
	tests.Add("!type, null", test{
		sel: &conditionNode{
			op:   OpType,
			cond: "null",
		},
		doc:  "foo",
		want: false,
	})
	tests.Add("type, boolean", test{
		sel: &conditionNode{
			op:   OpType,
			cond: "boolean",
		},
		doc:  true,
		want: true,
	})
	tests.Add("!type, boolean", test{
		sel: &conditionNode{
			op:   OpType,
			cond: "boolean",
		},
		doc:  "foo",
		want: false,
	})
	tests.Add("type, number", test{
		sel: &conditionNode{
			op:   OpType,
			cond: "number",
		},
		doc:  float64(5),
		want: true,
	})
	tests.Add("!type, number", test{
		sel: &conditionNode{
			op:   OpType,
			cond: "number",
		},
		doc:  "foo",
		want: false,
	})
	tests.Add("type, string", test{
		sel: &conditionNode{
			op:   OpType,
			cond: "string",
		},
		doc:  "foo",
		want: true,
	})
	tests.Add("!type, string", test{
		sel: &conditionNode{
			op:   OpType,
			cond: "string",
		},
		doc:  float64(5),
		want: false,
	})
	tests.Add("type, array", test{
		sel: &conditionNode{
			op:   OpType,
			cond: "array",
		},
		doc:  []any{"foo"},
		want: true,
	})
	tests.Add("!type, array", test{
		sel: &conditionNode{
			op:   OpType,
			cond: "array",
		},
		doc:  "foo",
		want: false,
	})
	tests.Add("type, object", test{
		sel: &conditionNode{
			op:   OpType,
			cond: "object",
		},
		doc:  map[string]any{"foo": "bar"},
		want: true,
	})
	tests.Add("!type, object", test{
		sel: &conditionNode{
			op:   OpType,
			cond: "object",
		},
		doc:  "foo",
		want: false,
	})
	tests.Add("in", test{
		sel: &conditionNode{
			op:   OpIn,
			cond: []any{"foo", "bar"},
		},
		doc:  "foo",
		want: true,
	})
	tests.Add("!in", test{
		sel: &conditionNode{
			op:   OpIn,
			cond: []any{"foo", "bar"},
		},
		doc:  "baz",
		want: false,
	})
	tests.Add("not in", test{
		sel: &conditionNode{
			op:   OpNotIn,
			cond: []any{"foo", "bar"},
		},
		doc:  "baz",
		want: true,
	})
	tests.Add("!not in", test{
		sel: &conditionNode{
			op:   OpNotIn,
			cond: []any{"foo", "bar"},
		},
		doc:  "foo",
		want: false,
	})
	tests.Add("size", test{
		sel: &conditionNode{
			op:   OpSize,
			cond: float64(3),
		},
		doc:  []any{"foo", "bar", "baz"},
		want: true,
	})
	tests.Add("!size", test{
		sel: &conditionNode{
			op:   OpSize,
			cond: float64(3),
		},
		doc:  []any{"foo", "bar"},
		want: false,
	})
	tests.Add("size, non-array", test{
		sel: &conditionNode{
			op:   OpSize,
			cond: float64(3),
		},
		doc:  "foo",
		want: false,
	})
	tests.Add("mod", test{
		sel: &conditionNode{
			op:   OpMod,
			cond: [2]int64{3, 2},
		},
		doc:  float64(8),
		want: true,
	})
	tests.Add("!mod", test{
		sel: &conditionNode{
			op:   OpMod,
			cond: [2]int64{3, 2},
		},
		doc:  float64(7),
		want: false,
	})
	tests.Add("mod, non-integer", test{
		sel: &conditionNode{
			op:   OpMod,
			cond: [2]int64{3, 2},
		},
		doc:  float64(7.5),
		want: false,
	})
	tests.Add("mod, non-number", test{
		sel: &conditionNode{
			op:   OpMod,
			cond: [2]int64{3, 2},
		},
		doc:  "foo",
		want: false,
	})
	tests.Add("regex", test{
		sel: &conditionNode{
			op:   OpRegex,
			cond: regexp.MustCompile("^foo$"),
		},
		doc:  "foo",
		want: true,
	})
	tests.Add("!regex", test{
		sel: &conditionNode{
			op:   OpRegex,
			cond: regexp.MustCompile("^foo$"),
		},
		doc:  "bar",
		want: false,
	})
	tests.Add("regexp, non-string", test{
		sel: &conditionNode{
			op:   OpRegex,
			cond: regexp.MustCompile("^foo$"),
		},
		doc:  float64(5),
		want: false,
	})
	tests.Add("all", test{
		sel: &conditionNode{
			op:   OpAll,
			cond: []any{"Comedy", "Short"},
		},
		doc: []any{
			"Comedy",
			"Short",
			"Animation",
		},
		want: true,
	})
	tests.Add("!all", test{
		sel: &conditionNode{
			op:   OpAll,
			cond: []any{"Comedy", "Short"},
		},
		doc: []any{
			"Comedy",
			"Animation",
		},
		want: false,
	})
	tests.Add("all, non-array", test{
		sel: &conditionNode{
			op:   OpAll,
			cond: []any{"Comedy", "Short"},
		},
		doc:  "Comedy",
		want: false,
	})
	tests.Add("field selector", test{
		sel: &fieldNode{
			field: "foo",
			cond: &conditionNode{
				op:   OpEqual,
				cond: "bar",
			},
		},
		doc: map[string]any{
			"foo": "bar",
		},
		want: true,
	})
	tests.Add("!field selector", test{
		sel: &fieldNode{
			field: "asdf",
			cond: &conditionNode{
				op:   OpEqual,
				cond: "foo",
			},
		},
		doc: map[string]any{
			"foo": "bar",
		},
		want: false,
	})
	tests.Add("field selector, non-object", test{
		sel: &fieldNode{
			field: "foo",
			cond: &conditionNode{
				op:   OpEqual,
				cond: "bar",
			},
		},
		doc:  "bar",
		want: false,
	})
	tests.Add("field selector, nested", test{
		sel: &fieldNode{
			field: "foo.bar.baz",
			cond: &conditionNode{
				op:   OpEqual,
				cond: "hello",
			},
		},
		doc: map[string]any{
			"foo": map[string]any{
				"bar": map[string]any{
					"baz": "hello",
				},
			},
		},
		want: true,
	})
	tests.Add("field selector, nested, non-object", test{
		sel: &fieldNode{
			field: "foo.bar.baz",
			cond: &conditionNode{
				op:   OpEqual,
				cond: "hello",
			},
		},
		doc: map[string]any{
			"foo": "hello",
		},
		want: false,
	})
	tests.Add("!field selector, nested", test{
		sel: &fieldNode{
			field: "foo.bar.baz",
			cond: &conditionNode{
				op:   OpEqual,
				cond: "hello",
			},
		},
		doc: map[string]any{
			"foo": map[string]any{
				"bar": map[string]any{
					"buzz": "hello",
				},
			},
		},
		want: false,
	})
	tests.Add("elemMatch", test{
		sel: &fieldNode{
			field: "foo",
			cond: &elementNode{
				op: OpElemMatch,
				cond: &conditionNode{
					op:   OpEqual,
					cond: "Horror",
				},
			},
		},
		doc: map[string]any{
			"foo": []any{
				"Comedy",
				"Horror",
			},
		},
		want: true,
	})
	tests.Add("!elemMatch", test{
		sel: &fieldNode{
			field: "genre",
			cond: &elementNode{
				op: OpElemMatch,
				cond: &conditionNode{
					op:   OpEqual,
					cond: "Horror",
				},
			},
		},
		doc: map[string]any{
			"genre": []any{
				"Comedy",
			},
		},
		want: false,
	})
	tests.Add("elemMatch, non-array", test{
		sel: &fieldNode{
			field: "genre",
			cond: &elementNode{
				op: OpElemMatch,
				cond: &conditionNode{
					op:   OpEqual,
					cond: "Horror",
				},
			},
		},
		doc: map[string]any{
			"genre": "Comedy",
		},
		want: false,
	})
	tests.Add("allMatch", test{
		sel: &fieldNode{
			field: "genre",
			cond: &elementNode{
				op: OpAllMatch,
				cond: &conditionNode{
					op:   OpEqual,
					cond: "Horror",
				},
			},
		},
		doc: map[string]any{
			"genre": []any{
				"Horror",
				"Horror",
			},
		},
		want: true,
	})
	tests.Add("!allMatch", test{
		sel: &fieldNode{
			field: "genre",
			cond: &elementNode{
				op: OpAllMatch,
				cond: &conditionNode{
					op:   OpEqual,
					cond: "Horror",
				},
			},
		},
		doc: map[string]any{
			"genre": []any{
				"Horror",
				"Comedy",
			},
		},
		want: false,
	})
	tests.Add("allMatch, non-array", test{
		sel: &fieldNode{
			field: "genre",
			cond: &elementNode{
				op: OpAllMatch,
				cond: &conditionNode{
					op:   OpEqual,
					cond: "Horror",
				},
			},
		},
		doc: map[string]any{
			"genre": "Horror",
		},
		want: false,
	})
	tests.Add("keyMapMatch", test{
		sel: &fieldNode{
			field: "cameras",
			cond: &elementNode{
				op: OpKeyMapMatch,
				cond: &conditionNode{
					op:   OpEqual,
					cond: "secondary",
				},
			},
		},
		doc: map[string]any{
			"cameras": map[string]any{
				"primary":   "Canon",
				"secondary": "Nikon",
			},
		},
		want: true,
	})
	tests.Add("!keyMapMatch", test{
		sel: &fieldNode{
			field: "cameras",
			cond: &elementNode{
				op: OpKeyMapMatch,
				cond: &conditionNode{
					op:   OpEqual,
					cond: "secondary",
				},
			},
		},
		doc: map[string]any{
			"cameras": map[string]any{
				"primary": "Canon",
			},
		},
		want: false,
	})
	tests.Add("keyMapMatch, non-object", test{
		sel: &fieldNode{
			field: "cameras",
			cond: &elementNode{
				op: OpKeyMapMatch,
				cond: &conditionNode{
					op:   OpEqual,
					cond: "secondary",
				},
			},
		},
		doc: map[string]any{
			"cameras": []any{"Canon", "Nikon"},
		},
		want: false,
	})
	tests.Add("and", test{
		sel: &combinationNode{
			op: OpAnd,
			sel: []Node{
				&fieldNode{
					field: "foo",
					cond: &conditionNode{
						op:   OpEqual,
						cond: "bar",
					},
				},
				&fieldNode{
					field: "baz",
					cond: &conditionNode{
						op:   OpEqual,
						cond: "qux",
					},
				},
			},
		},
		doc: map[string]any{
			"foo": "bar",
			"baz": "qux",
		},
		want: true,
	})
	tests.Add("!and", test{
		sel: &combinationNode{
			op: OpAnd,
			sel: []Node{
				&fieldNode{
					field: "foo",
					cond: &conditionNode{
						op:   OpEqual,
						cond: "bar",
					},
				},
				&fieldNode{
					field: "baz",
					cond: &conditionNode{
						op:   OpEqual,
						cond: "qux",
					},
				},
			},
		},
		doc: map[string]any{
			"baz": "qux",
		},
		want: false,
	})
	tests.Add("or", test{
		sel: &combinationNode{
			op: OpOr,
			sel: []Node{
				&fieldNode{
					field: "foo",
					cond: &conditionNode{
						op:   OpEqual,
						cond: "bar",
					},
				},
				&fieldNode{
					field: "baz",
					cond: &conditionNode{
						op:   OpEqual,
						cond: "qux",
					},
				},
			},
		},
		doc: map[string]any{
			"foo": "bar",
			"baz": "quux",
		},
		want: true,
	})
	tests.Add("!or", test{
		sel: &combinationNode{
			op: OpOr,
			sel: []Node{
				&fieldNode{
					field: "foo",
					cond: &conditionNode{
						op:   OpEqual,
						cond: "bar",
					},
				},
				&fieldNode{
					field: "baz",
					cond: &conditionNode{
						op:   OpEqual,
						cond: "qux",
					},
				},
			},
		},
		doc: map[string]any{
			"foo": "baz",
			"baz": "quux",
		},
		want: false,
	})
	tests.Add("not", test{
		sel: &notNode{
			sel: &fieldNode{
				field: "foo",
				cond: &conditionNode{
					op:   OpEqual,
					cond: "bar",
				},
			},
		},
		doc: map[string]any{
			"foo": "baz",
		},
		want: true,
	})
	tests.Add("!not", test{
		sel: &notNode{
			sel: &fieldNode{
				field: "foo",
				cond: &conditionNode{
					op:   OpEqual,
					cond: "bar",
				},
			},
		},
		doc: map[string]any{
			"foo": "bar",
		},
		want: false,
	})
	tests.Add("nor", test{
		sel: &combinationNode{
			op: OpNor,
			sel: []Node{
				&fieldNode{
					field: "foo",
					cond: &conditionNode{
						op:   OpEqual,
						cond: "bar",
					},
				},
				&fieldNode{
					field: "baz",
					cond: &conditionNode{
						op:   OpEqual,
						cond: "qux",
					},
				},
			},
		},
		doc: map[string]any{
			"foo": "baz",
			"baz": "quux",
		},
		want: true,
	})
	tests.Add("!nor", test{
		sel: &combinationNode{
			op: OpNor,
			sel: []Node{
				&fieldNode{
					field: "foo",
					cond: &conditionNode{
						op:   OpEqual,
						cond: "bar",
					},
				},
				&fieldNode{
					field: "baz",
					cond: &conditionNode{
						op:   OpEqual,
						cond: "qux",
					},
				},
			},
		},
		doc: map[string]any{
			"foo": "bar",
			"baz": "quux",
		},
		want: false,
	})

	tests.Run(t, func(t *testing.T, tt test) {
		got := Match(tt.sel, tt.doc)
		if got != tt.want {
			t.Errorf("Unexpected result: %v", got)
		}
	})
}
