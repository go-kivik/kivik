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
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
)

// Selector represents a Mango Selector tree.
type Selector struct {
	root Node
}

// Match returns true if doc matches the selector.
func (s *Selector) Match(doc interface{}) bool {
	return s.root.Match(doc)
}

// UnmarshalJSON parses the JSON-encoded data and stores the result in s.
func (s *Selector) UnmarshalJSON(data []byte) error {
	node, err := Parse(data)
	if err != nil {
		return err
	}
	s.root = node
	return nil
}

// Parse parses s into a Mango Selector tree.
func Parse(input []byte) (Node, error) {
	var tmp map[string]json.RawMessage
	if err := json.Unmarshal(input, &tmp); err != nil {
		return nil, err
	}
	if len(tmp) == 0 {
		// Empty object is an implicit $and
		return &combinationNode{
			op:  OpAnd,
			sel: nil,
		}, nil
	}
	sels := make([]Node, 0, len(tmp))
	for k, v := range tmp {
		switch op := Operator(k); op {
		case OpAnd, OpOr, OpNor:
			var sel []json.RawMessage
			if err := json.Unmarshal(v, &sel); err != nil {
				return nil, fmt.Errorf("%s: %w", k, err)
			}
			subsels := make([]Node, 0, len(sel))
			for _, s := range sel {
				sel, err := Parse(s)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", k, err)
				}
				subsels = append(subsels, sel)
			}

			sels = append(sels, &combinationNode{
				op:  op,
				sel: subsels,
			})
		case OpNot:
			sel, err := Parse(v)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", k, err)
			}
			sels = append(sels, &notNode{
				sel: sel,
			})
		case OpEqual, OpLessThan, OpLessThanOrEqual, OpNotEqual,
			OpGreaterThan, OpGreaterThanOrEqual:
			op, value, err := opAndValue(v)
			if err != nil {
				return nil, err
			}
			sels = append(sels, &conditionNode{
				op:   op,
				cond: value,
			})
		default:
			if op[0] == '$' {
				return nil, fmt.Errorf("unknown operator %s", op)
			}
			op, value, err := opAndValue(v)
			if err != nil {
				return nil, err
			}

			switch op {
			case OpElemMatch, OpAllMatch, OpKeyMapMatch:
				sels = append(sels, &fieldNode{
					field: k,
					cond: &elementNode{
						op:   op,
						cond: value.(*conditionNode),
					},
				})
			default:
				sels = append(sels, &fieldNode{
					field: k,
					cond: &conditionNode{
						op:   op,
						cond: value,
					},
				})
			}
		}
	}
	if len(sels) == 1 {
		return sels[0], nil
	}

	// Sort the selectors to ensure deterministic output.
	sort.Slice(sels, func(i, j int) bool {
		return cmpSelectors(sels[i], sels[j]) < 0
	})

	return &combinationNode{
		op:  OpAnd,
		sel: sels,
	}, nil
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
	var tmp map[string]json.RawMessage
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
				var value interface{}
				err := json.Unmarshal(v, &value)
				return op, value, err
			case OpExists:
				var value bool
				if err := json.Unmarshal(v, &value); err != nil {
					return "", nil, fmt.Errorf("%s: %w", k, err)
				}
				return OpExists, value, nil
			case OpType:
				var value string
				if err := json.Unmarshal(v, &value); err != nil {
					return "", nil, fmt.Errorf("%s: %w", k, err)
				}
				return OpType, value, nil
			case OpIn, OpNotIn:
				var value []interface{}
				if err := json.Unmarshal(v, &value); err != nil {
					return "", nil, fmt.Errorf("%s: %w", k, err)
				}
				return op, value, nil
			case OpSize:
				var value uint
				if err := json.Unmarshal(v, &value); err != nil {
					return "", nil, fmt.Errorf("%s: %w", k, err)
				}
				return OpSize, float64(value), nil
			case OpMod:
				var value [2]int64
				if err := json.Unmarshal(v, &value); err != nil {
					return "", nil, fmt.Errorf("%s: %w", k, err)
				}
				if value[0] == 0 {
					return "", nil, errors.New("$mod: divisor must be non-zero")
				}
				return OpMod, value, nil
			case OpRegex:
				var pattern string
				if err := json.Unmarshal(v, &pattern); err != nil {
					return "", nil, fmt.Errorf("%s: %w", k, err)
				}
				re, err := regexp.Compile(pattern)
				if err != nil {
					return "", nil, fmt.Errorf("%s: %w", k, err)
				}
				return OpRegex, re, nil
			case OpAll:
				var value []interface{}
				if err := json.Unmarshal(v, &value); err != nil {
					return "", nil, fmt.Errorf("%s: %w", k, err)
				}
				return OpAll, value, nil
			case OpElemMatch, OpAllMatch, OpKeyMapMatch:
				sel, err := Parse(v)
				if err != nil {
					return "", nil, fmt.Errorf("%s: %w", k, err)
				}
				return op, sel, nil
			}
			return "", nil, fmt.Errorf("invalid operator %s", k)
		}
	default:
		return "", nil, errors.New("too many keys in object")
	}
	panic("impossible")
}
