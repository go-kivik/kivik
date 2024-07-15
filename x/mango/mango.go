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
	"fmt"

	"github.com/go-kivik/kivik/v4/x/mango/collate"
)

// Selector represents a CouchDB Find query selector. See
// http://docs.couchdb.org/en/2.0.0/api/database/find.html#find-selectors
type Selector struct {
	op    operator
	field string
	value interface{}
	sel   []Selector
}

// New returns a new selector, parsed from data.
func New(data string) (*Selector, error) {
	s := &Selector{}
	err := json.Unmarshal([]byte(data), &s)
	return s, err
}

// UnmarshalJSON unmarshals a JSON selector as described in the CouchDB
// documentation.
// http://docs.couchdb.org/en/2.0.0/api/database/find.html#selector-syntax
func (s *Selector) UnmarshalJSON(data []byte) error {
	var x map[string]json.RawMessage
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if len(x) == 0 {
		return nil
	}
	sels := make([]Selector, 0, len(x))
	for k, v := range x {
		var sel Selector
		sel.field = k
		switch v[0] {
		case '{':
			var err error
			sel.op, sel.value, err = opObjectPattern(v)
			if err != nil {
				return err
			}
		case '[':
			sel.op = operator(k)
			if err := json.Unmarshal(v, &sel.sel); err != nil {
				return err
			}
		}
		if sel.op == "" {
			sel.op = opEq
			if err := json.Unmarshal(v, &sel.value); err != nil {
				return err
			}
		}
		sels = append(sels, sel)
	}
	if len(sels) == 1 {
		*s = sels[0]
	} else {
		*s = Selector{
			op:  opAnd,
			sel: sels,
		}
	}
	return nil
}

func opObjectPattern(data []byte) (op operator, value interface{}, err error) {
	var x map[operator]json.RawMessage
	if err := json.Unmarshal(data, &x); err != nil {
		return operator(""), nil, err
	}
	if len(x) != 1 {
		panic("got more than one result")
	}
	for k, v := range x {
		switch k {
		case opEq, opNE, opLT, opLTE, opGT, opGTE:
			var value interface{}
			err := json.Unmarshal(v, &value)
			return k, value, err
		default:
			if len(k) > 0 && k[0] == '$' {
				return "", nil, fmt.Errorf("unknown mango operator '%s'", k)
			}
			return opNone, v, nil
		}
	}
	return opNone, nil, nil
}

type couchDoc map[string]interface{}

// Matches returns true if the provided doc matches the selector.
func (s *Selector) Matches(doc couchDoc) (bool, error) {
	c := &collate.Raw{}
	switch s.op {
	case opNone:
		return true, nil
	case opEq, opGT, opGTE, opLT, opLTE:
		v, ok := doc[s.field]
		if !ok {
			return false, nil
		}
		switch s.op {
		case opEq:
			return c.Eq(v, s.value), nil
		case opGT:
			return c.GT(v, s.value), nil
		case opGTE:
			return c.GTE(v, s.value), nil
		case opLT:
			return c.LT(v, s.value), nil
		case opLTE:
			return c.LTE(v, s.value), nil
		}
	case opAnd:
		for _, sel := range s.sel {
			match, err := sel.Matches(doc)
			if err != nil || !match {
				return match, err
			}
		}
		return true, nil
	case opOr:
		for _, sel := range s.sel {
			match, err := sel.Matches(doc)
			if err != nil {
				return false, err
			}
			if match {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("unknown mango operator '%s'", s.op)
	}
	return true, nil
}
