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

// Package mango provides a Mango query language parser and evaluator.
package mango

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-kivik/kivik/v4/x/collate"
)

// Node represents a node in the Mango Selector.
type Node interface {
	Op() Operator
	Value() any
	String() string
	Match(any) bool
}

type notNode struct {
	sel Node
}

var _ Node = (*notNode)(nil)

func (*notNode) Op() Operator {
	return OpNot
}

func (n *notNode) Value() any {
	return n.sel
}

func (n *notNode) String() string {
	return fmt.Sprintf("%s %s", OpNot, n.sel)
}

func (n *notNode) Match(doc any) bool {
	return !n.sel.Match(doc)
}

type combinationNode struct {
	op  Operator
	sel []Node
}

var _ Node = (*combinationNode)(nil)

func (c *combinationNode) Op() Operator {
	return c.op
}

func (c *combinationNode) Value() any {
	return c.sel
}

func (c *combinationNode) String() string {
	var sb strings.Builder
	sb.WriteString(string(c.op))
	sb.WriteString(" [")
	for i, sel := range c.sel {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%v", sel))
	}
	sb.WriteString("]")
	return sb.String()
}

func (c *combinationNode) Match(doc any) bool {
	switch c.op {
	case OpAnd:
		for _, sel := range c.sel {
			if !sel.Match(doc) {
				return false
			}
		}
		return true
	case OpOr:
		for _, sel := range c.sel {
			if sel.Match(doc) {
				return true
			}
		}
		return false
	case OpNor:
		for _, sel := range c.sel {
			if sel.Match(doc) {
				return false
			}
		}
		return true
	}
	panic("not implemented")
}

type fieldNode struct {
	field string
	cond  Node
}

var _ Node = (*fieldNode)(nil)

func (f *fieldNode) Op() Operator {
	return f.cond.Op()
}

func (f *fieldNode) Value() any {
	return f.cond.Value()
}

func (f *fieldNode) String() string {
	return fmt.Sprintf("%s %s", f.field, f.cond.String())
}

func (f *fieldNode) Match(doc any) bool {
	val := doc

	// Traverse nested fields (e.g. "foo.bar.baz")
	segments := SplitKeys(f.field)
	for _, segment := range segments {
		m, ok := val.(map[string]any)
		if !ok {
			return false
		}

		val = m[segment]
	}

	// Even if the field does not exist we need to pass it to the condition expression because of `$exists`
	return f.cond.Match(val)
}

type conditionNode struct {
	op   Operator
	cond any
}

var _ Node = (*conditionNode)(nil)

func (e *conditionNode) Op() Operator {
	return e.op
}

func (e *conditionNode) Value() any {
	return e.cond
}

func (e *conditionNode) String() string {
	return fmt.Sprintf("%s %v", e.op, e.cond)
}

func (e *conditionNode) Match(doc any) bool {
	switch e.op {
	case OpEqual:
		return collate.CompareObject(doc, e.cond) == 0
	case OpNotEqual:
		return collate.CompareObject(doc, e.cond) != 0
	case OpLessThan:
		return collate.CompareObject(doc, e.cond) < 0
	case OpLessThanOrEqual:
		return collate.CompareObject(doc, e.cond) <= 0
	case OpGreaterThan:
		return collate.CompareObject(doc, e.cond) > 0
	case OpGreaterThanOrEqual:
		return collate.CompareObject(doc, e.cond) >= 0
	case OpExists:
		return (doc != nil) == e.cond.(bool)
	case OpType:
		switch tp := e.cond.(string); tp {
		case "null":
			return doc == nil
		case "boolean":
			_, ok := doc.(bool)
			return ok
		case "number":
			_, ok := doc.(float64)
			return ok
		case "string":
			_, ok := doc.(string)
			return ok
		case "array":
			_, ok := doc.([]any)
			return ok
		case "object":
			_, ok := doc.(map[string]any)
			return ok
		default:
			panic("unexpected $type value: " + tp)
		}
	case OpIn:
		for _, v := range e.cond.([]any) {
			if collate.CompareObject(doc, v) == 0 {
				return true
			}
		}
		return false
	case OpNotIn:
		for _, v := range e.cond.([]any) {
			if collate.CompareObject(doc, v) == 0 {
				return false
			}
		}
		return true
	case OpSize:
		array, ok := doc.([]any)
		if !ok {
			return false
		}
		return float64(len(array)) == e.cond.(float64)
	case OpMod:
		num, ok := doc.(float64)
		if !ok {
			return false
		}
		if num != float64(int(num)) {
			return false
		}
		mod := e.cond.([2]int64)
		return int64(num)%mod[0] == mod[1]
	case OpRegex:
		str, ok := doc.(string)
		if !ok {
			return false
		}
		return e.cond.(*regexp.Regexp).MatchString(str)
	case OpAll:
		array, ok := doc.([]any)
		if !ok {
			return false
		}
		for _, v := range e.cond.([]any) {
			if !contains(array, v) {
				return false
			}
		}
		return true
	}
	return false
}

func contains(haystack []any, needle any) bool {
	for _, v := range haystack {
		if collate.CompareObject(v, needle) == 0 {
			return true
		}
	}
	return false
}

type elementNode struct {
	op   Operator
	cond *conditionNode
}

var _ Node = (*elementNode)(nil)

func (e *elementNode) Op() Operator {
	return e.op
}

func (e *elementNode) Value() any {
	return e.cond
}

func (e *elementNode) String() string {
	return fmt.Sprintf("%s {%s}", e.op, e.cond)
}

func (e *elementNode) Match(doc any) bool {
	switch e.op {
	case OpElemMatch:
		array, ok := doc.([]any)
		if !ok {
			return false
		}
		for _, v := range array {
			if e.cond.Match(v) {
				return true
			}
		}
		return false
	case OpAllMatch:
		array, ok := doc.([]any)
		if !ok {
			return false
		}
		for _, v := range array {
			if !e.cond.Match(v) {
				return false
			}
		}
		return true
	case OpKeyMapMatch:
		object, ok := doc.(map[string]any)
		if !ok {
			return false
		}
		for k := range object {
			if k == e.cond.cond.(string) {
				return e.cond.Match(k)
			}
		}
		return false
	}
	panic("unready")
}

// cmpValues compares two arbitrary values by converting them to strings.
func cmpValues(a, b any) int {
	return strings.Compare(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
}

// cmpSelectors compares two selectors, for ordering.
func cmpSelectors(a, b Node) int {
	// Naively sort operators alphabetically.
	if c := strings.Compare(string(a.Op()), string(b.Op())); c != 0 {
		return c
	}
	switch t := a.(type) {
	case *notNode:
		u := b.(*notNode)
		return cmpSelectors(t.sel, u.sel)
	case *combinationNode:
		u := b.(*combinationNode)
		for i := 0; i < len(t.sel) && i < len(u.sel); i++ {
			if c := cmpSelectors(t.sel[i], u.sel[i]); c != 0 {
				return c
			}
		}
		return len(t.sel) - len(u.sel)
	case *fieldNode:
		u := b.(*fieldNode)
		if c := strings.Compare(t.field, u.field); c != 0 {
			return c
		}
		return cmpSelectors(t.cond, u.cond)
	case *conditionNode:
		u := b.(*conditionNode)
		switch t.op {
		case OpIn, OpNotIn:
			for i := 0; i < len(t.cond.([]any)) && i < len(u.cond.([]any)); i++ {
				if c := cmpValues(t.cond.([]any)[i], u.cond.([]any)[i]); c != 0 {
					return c
				}
			}
			return len(t.cond.([]any)) - len(u.cond.([]any))
		case OpMod:
			tm := t.cond.([2]int)
			um := u.cond.([2]int)
			if tm[0] != um[0] {
				return tm[0] - um[0]
			}
			return tm[1] - um[1]
		default:
			return cmpValues(t.cond, u.cond)
		}
	}
	return 0
}

// SplitKeys splits a field into its component keys. For example,
// "foo.bar" is split into `["foo", "bar"]`. Escaped dots are not treated
// as separators, so `"foo\\.bar"` becomes `["foo.bar"]`.
func SplitKeys(field string) []string {
	var escaped bool
	result := []string{}
	word := make([]byte, 0, len(field))
	for _, ch := range field {
		if escaped {
			word = append(word, byte(ch))
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '.' {
			result = append(result, string(word))
			word = word[:0]
			continue
		}
		word = append(word, byte(ch))
	}
	if escaped {
		word = append(word, '\\')
	}
	return append(result, string(word))
}
