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

// Operator represents a Mango [operator].
//
// [operator]: https://docs.couchdb.org/en/stable/api/database/find.html#explicit-operators
type Operator string

// [Combination Operators]
//
// [Combination Operators]: https://docs.couchdb.org/en/stable/api/database/find.html#combination-operators
const (
	OpAnd         = Operator("$and")
	OpOr          = Operator("$or")
	OpNot         = Operator("$not")
	OpNor         = Operator("$nor")
	OpAll         = Operator("$all")
	OpElemMatch   = Operator("$elemMatch")
	OpAllMatch    = Operator("$allMatch")
	OpKeyMapMatch = Operator("$keyMapMatch")
)

// [Condition Operators]
//
// [Condition Operators]: https://docs.couchdb.org/en/stable/api/database/find.html#condition-operators
const (
	OpLessThan           = Operator("$lt")
	OpLessThanOrEqual    = Operator("$lte")
	OpEqual              = Operator("$eq")
	OpNotEqual           = Operator("$ne")
	OpGreaterThan        = Operator("$gt")
	OpGreaterThanOrEqual = Operator("$gte")
	OpExists             = Operator("$exists")
	OpType               = Operator("$type")
	OpIn                 = Operator("$in")
	OpNotIn              = Operator("$nin")
	OpSize               = Operator("$size")
	OpMod                = Operator("$mod")
	OpRegex              = Operator("$regex")
)
