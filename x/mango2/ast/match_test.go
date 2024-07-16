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

// Package ast provides the abstract syntax tree for Mango selectors.
package ast

import (
	"regexp"
	"testing"

	"gitlab.com/flimzy/testy"
)

func TestMatch(t *testing.T) {
	type test struct {
		sel  Selector
		doc  interface{}
		want bool
	}

	tests := testy.NewTable()
	tests.Add("nil selector", test{
		sel:  nil,
		doc:  "foo",
		want: true,
	})
	tests.Add("equality", test{
		sel: &conditionSelector{
			op:   OpEqual,
			cond: "foo",
		},
		doc:  "foo",
		want: true,
	})
	tests.Add("!equality", test{
		sel: &conditionSelector{
			op:   OpEqual,
			cond: "foo",
		},
		doc:  "bar",
		want: false,
	})
	tests.Add("inequality", test{
		sel: &conditionSelector{
			op:   OpNotEqual,
			cond: "foo",
		},
		doc:  "bar",
		want: true,
	})
	tests.Add("!inequality", test{
		sel: &conditionSelector{
			op:   OpNotEqual,
			cond: "foo",
		},
		doc:  "foo",
		want: false,
	})
	tests.Add("less than", test{
		sel: &conditionSelector{
			op:   OpLessThan,
			cond: float64(5),
		},
		doc:  float64(4),
		want: true,
	})
	tests.Add("!less than", test{
		sel: &conditionSelector{
			op:   OpLessThan,
			cond: float64(5),
		},
		doc:  float64(10),
		want: false,
	})
	tests.Add("less than or equal", test{
		sel: &conditionSelector{
			op:   OpLessThanOrEqual,
			cond: float64(5),
		},
		doc:  float64(5),
		want: true,
	})
	tests.Add("!less than or equal", test{
		sel: &conditionSelector{
			op:   OpLessThanOrEqual,
			cond: float64(5),
		},
		doc:  float64(8),
		want: false,
	})
	tests.Add("greater than", test{
		sel: &conditionSelector{
			op:   OpGreaterThan,
			cond: float64(5),
		},
		doc:  float64(10),
		want: true,
	})
	tests.Add("!greater than", test{
		sel: &conditionSelector{
			op:   OpGreaterThan,
			cond: float64(5),
		},
		doc:  float64(2),
		want: false,
	})
	tests.Add("greater than or equal", test{
		sel: &conditionSelector{
			op:   OpGreaterThanOrEqual,
			cond: float64(5),
		},
		doc:  float64(5),
		want: true,
	})
	tests.Add("!greater than or equal", test{
		sel: &conditionSelector{
			op:   OpGreaterThanOrEqual,
			cond: float64(5),
		},
		doc:  float64(2),
		want: false,
	})
	tests.Add("exists", test{
		sel: &conditionSelector{
			op:   OpExists,
			cond: true,
		},
		doc:  "foo",
		want: true,
	})
	tests.Add("!exists", test{
		sel: &conditionSelector{
			op:   OpExists,
			cond: false,
		},
		doc:  "foo",
		want: false,
	})
	tests.Add("not exists", test{
		sel: &conditionSelector{
			op:   OpExists,
			cond: false,
		},
		doc:  nil,
		want: true,
	})
	tests.Add("!not exists", test{
		sel: &conditionSelector{
			op:   OpExists,
			cond: true,
		},
		doc:  nil,
		want: false,
	})
	tests.Add("type, null", test{
		sel: &conditionSelector{
			op:   OpType,
			cond: "null",
		},
		doc:  nil,
		want: true,
	})
	tests.Add("!type, null", test{
		sel: &conditionSelector{
			op:   OpType,
			cond: "null",
		},
		doc:  "foo",
		want: false,
	})
	tests.Add("type, boolean", test{
		sel: &conditionSelector{
			op:   OpType,
			cond: "boolean",
		},
		doc:  true,
		want: true,
	})
	tests.Add("!type, boolean", test{
		sel: &conditionSelector{
			op:   OpType,
			cond: "boolean",
		},
		doc:  "foo",
		want: false,
	})
	tests.Add("type, number", test{
		sel: &conditionSelector{
			op:   OpType,
			cond: "number",
		},
		doc:  float64(5),
		want: true,
	})
	tests.Add("!type, number", test{
		sel: &conditionSelector{
			op:   OpType,
			cond: "number",
		},
		doc:  "foo",
		want: false,
	})
	tests.Add("type, string", test{
		sel: &conditionSelector{
			op:   OpType,
			cond: "string",
		},
		doc:  "foo",
		want: true,
	})
	tests.Add("!type, string", test{
		sel: &conditionSelector{
			op:   OpType,
			cond: "string",
		},
		doc:  float64(5),
		want: false,
	})
	tests.Add("type, array", test{
		sel: &conditionSelector{
			op:   OpType,
			cond: "array",
		},
		doc:  []interface{}{"foo"},
		want: true,
	})
	tests.Add("!type, array", test{
		sel: &conditionSelector{
			op:   OpType,
			cond: "array",
		},
		doc:  "foo",
		want: false,
	})
	tests.Add("type, object", test{
		sel: &conditionSelector{
			op:   OpType,
			cond: "object",
		},
		doc:  map[string]interface{}{"foo": "bar"},
		want: true,
	})
	tests.Add("!type, object", test{
		sel: &conditionSelector{
			op:   OpType,
			cond: "object",
		},
		doc:  "foo",
		want: false,
	})
	tests.Add("in", test{
		sel: &conditionSelector{
			op:   OpIn,
			cond: []interface{}{"foo", "bar"},
		},
		doc:  "foo",
		want: true,
	})
	tests.Add("!in", test{
		sel: &conditionSelector{
			op:   OpIn,
			cond: []interface{}{"foo", "bar"},
		},
		doc:  "baz",
		want: false,
	})
	tests.Add("not in", test{
		sel: &conditionSelector{
			op:   OpNotIn,
			cond: []interface{}{"foo", "bar"},
		},
		doc:  "baz",
		want: true,
	})
	tests.Add("!not in", test{
		sel: &conditionSelector{
			op:   OpNotIn,
			cond: []interface{}{"foo", "bar"},
		},
		doc:  "foo",
		want: false,
	})
	tests.Add("size", test{
		sel: &conditionSelector{
			op:   OpSize,
			cond: float64(3),
		},
		doc:  []interface{}{"foo", "bar", "baz"},
		want: true,
	})
	tests.Add("!size", test{
		sel: &conditionSelector{
			op:   OpSize,
			cond: float64(3),
		},
		doc:  []interface{}{"foo", "bar"},
		want: false,
	})
	tests.Add("size, non-array", test{
		sel: &conditionSelector{
			op:   OpSize,
			cond: float64(3),
		},
		doc:  "foo",
		want: false,
	})
	tests.Add("mod", test{
		sel: &conditionSelector{
			op:   OpMod,
			cond: [2]int64{3, 2},
		},
		doc:  float64(8),
		want: true,
	})
	tests.Add("!mod", test{
		sel: &conditionSelector{
			op:   OpMod,
			cond: [2]int64{3, 2},
		},
		doc:  float64(7),
		want: false,
	})
	tests.Add("mod, non-integer", test{
		sel: &conditionSelector{
			op:   OpMod,
			cond: [2]int64{3, 2},
		},
		doc:  float64(7.5),
		want: false,
	})
	tests.Add("mod, non-number", test{
		sel: &conditionSelector{
			op:   OpMod,
			cond: [2]int64{3, 2},
		},
		doc:  "foo",
		want: false,
	})
	tests.Add("regex", test{
		sel: &conditionSelector{
			op:   OpRegex,
			cond: regexp.MustCompile("^foo$"),
		},
		doc:  "foo",
		want: true,
	})
	tests.Add("!regex", test{
		sel: &conditionSelector{
			op:   OpRegex,
			cond: regexp.MustCompile("^foo$"),
		},
		doc:  "bar",
		want: false,
	})
	tests.Add("regexp, non-string", test{
		sel: &conditionSelector{
			op:   OpRegex,
			cond: regexp.MustCompile("^foo$"),
		},
		doc:  float64(5),
		want: false,
	})
	tests.Add("all", test{
		sel: &conditionSelector{
			op:   OpAll,
			cond: []interface{}{"Comedy", "Short"},
		},
		doc: []interface{}{
			"Comedy",
			"Short",
			"Animation",
		},
		want: true,
	})
	tests.Add("!all", test{
		sel: &conditionSelector{
			op:   OpAll,
			cond: []interface{}{"Comedy", "Short"},
		},
		doc: []interface{}{
			"Comedy",
			"Animation",
		},
		want: false,
	})
	tests.Add("all, non-array", test{
		sel: &conditionSelector{
			op:   OpAll,
			cond: []interface{}{"Comedy", "Short"},
		},
		doc:  "Comedy",
		want: false,
	})
	tests.Add("field selector", test{
		sel: &fieldSelector{
			field: "foo",
			cond: &conditionSelector{
				op:   OpEqual,
				cond: "bar",
			},
		},
		doc: map[string]interface{}{
			"foo": "bar",
		},
		want: true,
	})
	tests.Add("!field selector", test{
		sel: &fieldSelector{
			field: "asdf",
			cond: &conditionSelector{
				op:   OpEqual,
				cond: "foo",
			},
		},
		doc: map[string]interface{}{
			"foo": "bar",
		},
		want: false,
	})
	tests.Add("field selector, non-object", test{
		sel: &fieldSelector{
			field: "foo",
			cond: &conditionSelector{
				op:   OpEqual,
				cond: "bar",
			},
		},
		doc:  "bar",
		want: false,
	})

	/*
		TODO:
		$elemMatch
		$allMatch
		$keyMapMatch
		$and
		$or
		$not
		$nor
	*/

	tests.Run(t, func(t *testing.T, tt test) {
		got := Match(tt.sel, tt.doc)
		if got != tt.want {
			t.Errorf("Unexpected result: %v", got)
		}
	})
}
