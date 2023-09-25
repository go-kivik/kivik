package mango

import (
	"fmt"
	"sort"
	"testing"

	"gitlab.com/flimzy/testy"
)

type Selectors []Selector

var _ sort.Interface = &Selectors{}

func (s Selectors) Len() int           { return len(s) }
func (s Selectors) Less(i, j int) bool { return s[i].field < s[j].field }
func (s Selectors) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func TestUnmarshal(t *testing.T) {
	type uTest struct {
		name     string
		input    string
		expected Selector
		err      string
	}
	tests := []uTest{
		{
			name:     "Empty selector",
			input:    `{}`,
			expected: Selector{},
		},
		{
			name:  "Invalid operator",
			input: `{"foo":{"$invalid":"bar"}}`,
			err:   "unknown mango operator '$invalid'",
		},
		{
			name:  "invalid JSON",
			input: "xxx",
			err:   "invalid character 'x' looking for beginning of value",
		},
		{
			// http://docs.couchdb.org/en/2.0.0/api/database/find.html#selector-basics
			name:     "basic",
			input:    `{"director":"Lars von Trier"}`,
			expected: Selector{op: opEq, field: "director", value: "Lars von Trier"},
		},
		{
			// http://docs.couchdb.org/en/2.0.0/api/database/find.html#selector-with-2-fields
			name:  "selector with two fields",
			input: `{"name": "Paul", "location": "Boston"}`,
			expected: Selector{
				op: opAnd,
				sel: []Selector{
					{op: opEq, field: "location", value: "Boston"},
					{op: opEq, field: "name", value: "Paul"},
				},
			},
		},
		{
			name:     "explicit $eq",
			input:    `{"director":{"$eq":"Lars von Trier"}}`,
			expected: Selector{op: opEq, field: "director", value: "Lars von Trier"},
		},
		{
			name:     "explicit $gt",
			input:    `{"director":{"$gt":"Lars von Trier"}}`,
			expected: Selector{op: opGT, field: "director", value: "Lars von Trier"},
		},
		{
			name:     "explicit $gte",
			input:    `{"director":{"$gte":"Lars von Trier"}}`,
			expected: Selector{op: opGTE, field: "director", value: "Lars von Trier"},
		},
		{
			name:     "explicit $lt",
			input:    `{"director":{"$lt":"Lars von Trier"}}`,
			expected: Selector{op: opLT, field: "director", value: "Lars von Trier"},
		},
		{
			name:     "explicit $lte",
			input:    `{"director":{"$lte":"Lars von Trier"}}`,
			expected: Selector{op: opLTE, field: "director", value: "Lars von Trier"},
		},
		{
			name:     "find test",
			input:    `{"_id":{"$gt":null}}`,
			expected: Selector{op: opGT, field: "_id", value: nil},
		},
		// {
		// 	// http://docs.couchdb.org/en/2.0.0/api/database/find.html#subfields
		// 	name:  "subfields 1",
		// 	input: `{"imdb": {"rating": 8}}`,
		// 	expected: Selector{
		// 		op:      opEq,
		// 		field:   "imdb.rating",
		// 		pattern: []byte("8"),
		// 	},
		// },
	}
	for _, op := range []operator{opLT, opLTE, opEq, opNE, opGTE, opGT} {
		tests = append(tests, uTest{
			name:  string(op),
			input: fmt.Sprintf(`{"director": {"%s": "Lars von Trier"}}`, op),
			expected: Selector{
				op:    op,
				field: "director",
				value: "Lars von Trier",
			},
		})
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := &Selector{}
			err := result.UnmarshalJSON([]byte(test.input))
			var msg string
			if err != nil {
				msg = err.Error()
			}
			if msg != test.err {
				t.Errorf("Unexpected error: %s", msg)
			}
			if err != nil {
				return
			}
			sort.Sort(Selectors(result.sel))
			if d := testy.DiffInterface(test.expected, *result); d != nil {
				t.Error(d)
			}
		})
	}
}

func mustNew(data string) *Selector {
	s, e := New(data)
	if e != nil {
		panic(e)
	}
	return s
}

func TestMatches(t *testing.T) {
	type mTest struct {
		name     string
		sel      *Selector
		doc      couchDoc
		expected bool
		err      string
	}
	tests := []mTest{
		{
			name: "invalid op",
			sel:  &Selector{op: "$invalid"},
			err:  "unknown mango operator '$invalid'",
		},
		{
			name:     "empty selecotor",
			sel:      mustNew("{}"),
			doc:      couchDoc{"foo": "bar"},
			expected: true,
		},
		{
			name:     "exact match hit",
			sel:      mustNew(`{"foo":"bar"}`),
			doc:      couchDoc{"foo": "bar"},
			expected: true,
		},
		{
			name:     "exact match miss",
			sel:      mustNew(`{"foo":"bar"}`),
			doc:      couchDoc{"foo": "baz"},
			expected: false,
		},
		{
			name:     "missing field",
			sel:      mustNew(`{"foo":"bar"}`),
			doc:      couchDoc{"boo": "baz"},
			expected: false,
		},
		{
			name:     "compound match hit",
			sel:      mustNew(`{"foo":"bar","baz":"qux"}`),
			doc:      couchDoc{"foo": "bar", "baz": "qux"},
			expected: true,
		},
		{
			name:     "compound match, one miss",
			sel:      mustNew(`{"foo":"bar","baz":"qux"}`),
			doc:      couchDoc{"foo": "bar", "baz": "quxx"},
			expected: false,
		},
		{
			name:     "explicit $eq",
			sel:      mustNew(`{"foo":{"$eq":"bar"}}`),
			doc:      couchDoc{"foo": "bar", "baz": "quxx"},
			expected: true,
		},
		{
			name:     "$gt",
			sel:      mustNew(`{"foo":{"$gt":"bar"}}`),
			doc:      couchDoc{"foo": "bar"},
			expected: false,
		},
		{
			name:     "$gte",
			sel:      mustNew(`{"foo":{"$gte":"bar"}}`),
			doc:      couchDoc{"foo": "bar"},
			expected: true,
		},
		{
			name:     "$lte",
			sel:      mustNew(`{"foo":{"$lte":"bar"}}`),
			doc:      couchDoc{"foo": "bar"},
			expected: true,
		},
		{
			name:     "$lt",
			sel:      mustNew(`{"foo":{"$lt":"bar"}}`),
			doc:      couchDoc{"foo": "bar"},
			expected: false,
		},
		{
			name:     "$lt zzz",
			sel:      mustNew(`{"foo":{"$lt":"bar"}}`),
			doc:      couchDoc{"foo": "zzz"},
			expected: false,
		},
		{
			name:     "$lt aaa",
			sel:      mustNew(`{"foo":{"$lt":"bar"}}`),
			doc:      couchDoc{"foo": "aaa"},
			expected: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.sel.Matches(test.doc)
			var msg string
			if err != nil {
				msg = err.Error()
			}
			if msg != test.err {
				t.Errorf("Unexpected error: %s", err)
			}
			if err != nil {
				return
			}
			if result != test.expected {
				t.Errorf("Expected %t, got %t", test.expected, result)
			}
		})
	}
}
