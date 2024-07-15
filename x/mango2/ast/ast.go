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
	"encoding/json"
	"errors"
	"fmt"
)

// Parse parses s into a Mango Selector tree.
func Parse(input []byte) (Selector, error) {
	var tmp map[string]json.RawMessage
	if err := json.Unmarshal(input, &tmp); err != nil {
		return nil, err
	}
	if len(tmp) == 0 {
		return &combinationSelector{
			op:  OpAnd,
			sel: nil,
		}, nil
	}
	if len(tmp) == 1 {
		for k, v := range tmp {
			op, value, err := opAndValue(v)
			if err != nil {
				return nil, err
			}
			return &conditionSelector{
				field: k,
				op:    op,
				value: value,
			}, nil
		}
	}
	panic("not implemented")
}

// opAndValue is called when the input is an object in a context where a
// comparison operator is expected. It returns the operator and value,
// defaulting to [OpEqual] if no operator is specified.
func opAndValue(input json.RawMessage) (Operator, interface{}, error) {
	if input[0] != '{' {
		var value interface{}
		if err := json.Unmarshal(input, &value); err != nil {
			return "", nil, err
		}
		return OpEqual, value, nil
	}
	var tmp map[string]interface{}
	if err := json.Unmarshal(input, &tmp); err != nil {
		return "", nil, err
	}
	switch len(tmp) {
	case 0:
		return OpEqual, map[string]interface{}{}, nil
	case 1:
		for k, v := range tmp {
			switch op := Operator(k); op {
			case OpEqual, OpLessThan, OpLessThanOrEqual, OpNotEqual,
				OpGreaterThan, OpGreaterThanOrEqual:
				return op, v, nil
			}
			return "", nil, fmt.Errorf("invalid operator %s", k)
		}
	default:
		return "", nil, errors.New("too many keys in object")
	}
	panic("impossible")
}
