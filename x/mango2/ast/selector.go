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

// Selector represents a node in the Mango Selector.
type Selector interface {
	Op() Operator
	Value() interface{}
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

type conditionSelector struct {
	field string
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

/*

 - $and []Selector
 - $or []Selector
 - $not Selector
 - $nor []Selector
 - $all []Selector
 - $elemMatch Selector
 - $allMatch Selector
 - $keyMapMatch Selector

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

*/
