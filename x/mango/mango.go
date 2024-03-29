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
		var op operator
		var field string
		var value interface{}
		field = k
		if v[0] == '{' {
			var e error
			op, value, e = opPattern(v)
			if e != nil {
				return e
			}
		}
		if op == "" {
			op = opEq
			if e := json.Unmarshal(v, &value); e != nil {
				return e
			}
		}
		sels = append(sels, Selector{
			op:    op,
			field: field,
			value: value,
		})
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

func opPattern(data []byte) (op operator, value interface{}, err error) {
	var x map[operator]json.RawMessage
	if e := json.Unmarshal(data, &x); e != nil {
		return operator(""), nil, e
	}
	if len(x) != 1 {
		panic("got more than one result")
	}
	for k, v := range x {
		switch k {
		case opEq, opNE, opLT, opLTE, opGT, opGTE:
			var value interface{}
			if e := json.Unmarshal(v, &value); e != nil {
				return "", nil, e
			}
			return k, value, nil
		default:
			return "", nil, fmt.Errorf("unknown mango operator '%s'", k)
		}
	}
	return operator(""), nil, nil
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
			m, e := sel.Matches(doc)
			if e != nil || !m {
				return m, e
			}
		}
		return true, nil
	default:
		return false, fmt.Errorf("unknown mango operator '%s'", s.op)
	}
	return true, nil
}
