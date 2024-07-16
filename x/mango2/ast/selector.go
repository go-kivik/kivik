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
	"fmt"
	"strings"

	"github.com/go-kivik/kivik/v4/x/collate"
)

// Selector represents a node in the Mango Selector.
type Selector interface {
	Op() Operator
	Value() interface{}
	String() string
	// Match(interface{}) (bool, error)
}

type unarySelector struct {
	op  Operator
	sel Selector
}

var _ Selector = (*unarySelector)(nil)

func (u *unarySelector) Op() Operator {
	return u.op
}

func (u *unarySelector) Value() interface{} {
	return u.sel
}

func (u *unarySelector) String() string {
	return fmt.Sprintf("%s %s", u.op, u.sel)
}

type combinationSelector struct {
	op  Operator
	sel []Selector
}

var _ Selector = (*combinationSelector)(nil)

func (c *combinationSelector) Op() Operator {
	return c.op
}

func (c *combinationSelector) Value() interface{} {
	return c.sel
}

func (c *combinationSelector) String() string {
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

type fieldSelector struct {
	field string
	cond  Selector
}

var _ Selector = (*fieldSelector)(nil)

func (f *fieldSelector) Op() Operator {
	return f.cond.Op()
}

func (f *fieldSelector) Value() interface{} {
	return f.cond.Value()
}

func (f *fieldSelector) String() string {
	return fmt.Sprintf("%s %s", f.field, f.cond.String())
}

type conditionSelector struct {
	op    Operator
	value interface{}
}

var _ Selector = (*conditionSelector)(nil)

func (e *conditionSelector) Op() Operator {
	return e.op
}

func (e *conditionSelector) Value() interface{} {
	return e.value
}

func (e *conditionSelector) String() string {
	return fmt.Sprintf("%s %v", e.op, e.value)
}

func (e *conditionSelector) Match(doc interface{}) bool {
	switch e.op {
	case OpEqual:
		return collate.CompareObject(doc, e.value) == 0
	case OpNotEqual:
		return collate.CompareObject(doc, e.value) != 0
	case OpLessThan:
		return collate.CompareObject(doc, e.value) < 0
	case OpLessThanOrEqual:
		return collate.CompareObject(doc, e.value) <= 0
	case OpGreaterThan:
		return collate.CompareObject(doc, e.value) > 0
	case OpGreaterThanOrEqual:
		return collate.CompareObject(doc, e.value) >= 0
	case OpExists:
		return (doc != nil) == e.value.(bool)
	case OpType:
		switch tp := e.value.(string); tp {
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
			_, ok := doc.([]interface{})
			return ok
		case "object":
			_, ok := doc.(map[string]interface{})
			return ok
		default:
			panic("unexpected $type value: " + tp)
		}
	case OpIn:
		for _, v := range e.value.([]interface{}) {
			if collate.CompareObject(doc, v) == 0 {
				return true
			}
		}
		return false
	case OpNotIn:
		for _, v := range e.value.([]interface{}) {
			if collate.CompareObject(doc, v) == 0 {
				return false
			}
		}
		return true
	case OpSize:
		array, ok := doc.([]interface{})
		if !ok {
			return false
		}
		return float64(len(array)) == e.value.(float64)
	}
	return false
}

type elementSelector struct {
	op   Operator
	cond *conditionSelector
}

var _ Selector = (*elementSelector)(nil)

func (e *elementSelector) Op() Operator {
	return e.op
}

func (e *elementSelector) Value() interface{} {
	return e.cond
}

func (e *elementSelector) String() string {
	return fmt.Sprintf("%s {%s}", e.op, e.cond)
}

/*

 - $and []Selector
 - $or []Selector
 - $not Selector
 - $nor []Selector

 - $lt Any JSON
 - $lte Any JSON
 - $eq Any JSON
 - $ne Any JSON
 - $gt Any JSON
 - $gte Any JSON
 - $exists Boolean
 - $type String
 - $in Array
 - $nin Array
 - $size Integer
 - $mod Divisor and Remainder
 - $regex String
 - $all Array
 - $elemMatch Selector
 - $allMatch Selector
 - $keyMapMatch Selector

*/

// cmpValues compares two arbitrary values by converting them to strings.
func cmpValues(a, b interface{}) int {
	return strings.Compare(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
}

// cmpSelectors compares two selectors, for ordering.
func cmpSelectors(a, b Selector) int {
	// Naively sort operators alphabetically.
	if c := strings.Compare(string(a.Op()), string(b.Op())); c != 0 {
		return c
	}
	switch t := a.(type) {
	case *unarySelector:
		u := b.(*unarySelector)
		return cmpSelectors(t.sel, u.sel)
	case *combinationSelector:
		u := b.(*combinationSelector)
		for i := 0; i < len(t.sel) && i < len(u.sel); i++ {
			if c := cmpSelectors(t.sel[i], u.sel[i]); c != 0 {
				return c
			}
		}
		return len(t.sel) - len(u.sel)
	case *fieldSelector:
		u := b.(*fieldSelector)
		if c := strings.Compare(t.field, u.field); c != 0 {
			return c
		}
		return cmpSelectors(t.cond, u.cond)
	case *conditionSelector:
		u := b.(*conditionSelector)
		switch t.op {
		case OpIn, OpNotIn:
			for i := 0; i < len(t.value.([]interface{})) && i < len(u.value.([]interface{})); i++ {
				if c := cmpValues(t.value.([]interface{})[i], u.value.([]interface{})[i]); c != 0 {
					return c
				}
			}
			return len(t.value.([]interface{})) - len(u.value.([]interface{}))
		case OpMod:
			tm := t.value.([2]int)
			um := u.value.([2]int)
			if tm[0] != um[0] {
				return tm[0] - um[0]
			}
			return tm[1] - um[1]
		default:
			return cmpValues(t.value, u.value)
		}
	}
	return 0
}
