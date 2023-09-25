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

type operator string

const (
	opNone = operator("")

	// Combination operators - http://docs.couchdb.org/en/2.0.0/api/database/find.html#combination-operators
	opAnd = operator("$and")
	// opOr        = operator("$or")
	// opNot       = operator("$not")
	// opNor       = operator("$nor")
	// opAll       = operator("$all")
	// opElemMatch = operator("$elemMatch")

	// Condition operators - http://docs.couchdb.org/en/2.0.0/api/database/find.html#condition-operators
	opLT  = operator("$lt")
	opLTE = operator("$lte")
	opEq  = operator("$eq")
	opNE  = operator("$ne")
	opGTE = operator("$gte")
	opGT  = operator("$gt")
	// opExists = operator("$exists")
	// opType   = operator("$type")
	// opIn     = operator("$in")
	// opNIn    = operator("$nin")
	// opSize   = operator("$size")
	// opMod    = operator("$mod")
	// opRegex  = operator("$regex")
)
