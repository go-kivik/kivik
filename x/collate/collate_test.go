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

// GopherJS does not work with go-cmp diffs
//go:build !js

package collate

import (
	"encoding/json"
	"math/rand"
	"sort"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCompareString(t *testing.T) {
	want := []string{
		// "\"`\"", `"^"`, // TODO: These don't sort according to CouchDB rules
		`"_"`, `"-"`, `","`, `";"`, `":"`, `"!"`, `"?"`,
		`"."`, `"'"`, `"""`, `"("`, `")"`, `"["`, `"]"`, `"{"`, `"}"`,
		`"@"`, `"*"`, `"/"`, `"\"`, `"&"`, `"#"`, `"%"`, `"+"`, `"<"`,
		`"="`, `">"`, `"|"`, `"~"`, `"$"`, `"0"`, `"1"`, `"2"`, `"3"`,
		`"4"`, `"5"`, `"6"`, `"7"`, `"8"`, `"9"`,
		`"a"`, `"A"`, `"b"`, `"B"`, `"c"`, `"C"`, `"d"`, `"D"`, `"e"`,
		`"E"`, `"f"`, `"F"`, `"g"`, `"G"`, `"h"`, `"H"`, `"i"`, `"I"`,
		`"j"`, `"J"`, `"k"`, `"K"`, `"l"`, `"L"`, `"m"`, `"M"`, `"n"`,
		`"N"`, `"o"`, `"O"`, `"p"`, `"P"`, `"q"`, `"Q"`, `"r"`, `"R"`,
		`"s"`, `"S"`, `"t"`, `"T"`, `"u"`, `"U"`, `"v"`, `"V"`, `"w"`,
		`"W"`, `"x"`, `"X"`, `"y"`, `"Y"`, `"z"`, `"Z"`,
	}
	input := make([]string, len(want))
	copy(input, want)
	// Shuffle the input
	rand.Shuffle(len(input), func(i, j int) { input[i], input[j] = input[j], input[i] })

	sort.Slice(input, func(i, j int) bool {
		return CompareString(input[i], input[j]) < 0
	})

	if d := cmp.Diff(want, input); d != "" {
		t.Errorf("Unexpected result:\n%s", d)
	}
}

func TestCompareObject(t *testing.T) {
	want := []any{
		nil,
		false,
		true,

		// then numbers
		float64(1),
		float64(2),
		float64(3.0),
		float64(4),

		// then text, case sensitive
		"a",
		"A",
		"aa",
		"b",
		"B",
		"ba",
		"bb",

		// then arrays. compared element by element until different.
		// Longer arrays sort after their prefixes
		[]any{"a"},
		[]any{"b"},
		[]any{"b", "c"},
		[]any{"b", "c", "a"},
		[]any{"b", "d"},
		[]any{"b", "d", "e"},

		// then object, compares each key value in the list until different.
		// larger objects sort after their subset objects.
		map[string]any{"a": float64(1)},
		map[string]any{"a": float64(2)},
		map[string]any{"b": float64(1)},
		map[string]any{"b": float64(2)},
		// TODO: See #952
		// map[string]any{"b": float64(2), "a": float64(1)}, // Member order does matter for collation. CouchDB preserves member order but doesn't require that clients will. this test might fail if used with a js engine that doesn't preserve order.
		map[string]any{"b": float64(2), "c": float64(2)},
	}

	input := make([]any, len(want))
	copy(input, want)
	// Shuffle the input
	rand.Shuffle(len(input), func(i, j int) { input[i], input[j] = input[j], input[i] })

	sort.Slice(input, func(i, j int) bool {
		return CompareObject(input[i], input[j]) < 0
	})

	if d := cmp.Diff(want, input); d != "" {
		t.Errorf("Unexpected result:\n%s", d)
	}
}

func Test_CompareJSON(t *testing.T) {
	want := []string{
		"",
		`null`,
		`false`,
		`true`,

		// then numbers
		`1`,
		`2`,
		`3.0`,
		`4`,

		// then text, case sensitive
		`"a"`,
		`"A"`,
		`"aa"`,
		`"b"`,
		`"B"`,
		`"ba"`,
		`"bb"`,

		// then arrays. compared element by element until different.
		// Longer arrays sort after their prefixes
		`["a"]`,
		`["b"]`,
		`["b","c"]`,
		`["b","c", "a"]`,
		`["b","d"]`,
		`["b","d", "e"]`,

		// then object, compares each key value in the list until different.
		// larger objects sort after their subset objects.
		`{"a":1}`,
		`{"a":2}`,
		`{"b":1}`,
		`{"b":2}`,
		// TODO: See #952
		// `{"b":2, "a":1}`, // Member order does matter for collation. CouchDB preserves member order but doesn't require that clients will. this test might fail if used with a js engine that doesn't preserve order.
		`{"b":2, "c":2}`,
		`{"za":1, "zb":2, "zc":3, "zd":4}`,
		`{"za":1, "zb":2, "zd":5, "zc":3}`,
		`{"za":1, "zb":2, "zd":6, "zc":3}`,
		`{"zd":7, "zb":2, "za":1, "zc":3}`,
		`{"za":1, "zd":8, "zb":2, "zc":3}`,
	}

	for i := 0; i < 10; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			// Clone and shuffle the input
			input := make([]json.RawMessage, len(want))
			for i, v := range want {
				input[i] = json.RawMessage(v)
			}
			rand.Shuffle(len(input), func(i, j int) { input[i], input[j] = input[j], input[i] })

			sort.Sort(couchdbKeys(input))

			got := make([]string, len(input))
			for i, v := range input {
				got[i] = string(v)
			}
			if d := cmp.Diff(want, got); d != "" {
				t.Errorf("Unexpected result:\n%s", d)
			}
		})
	}
}

// This test triggers the race detector if the collator isn't protected
func Test_collator_race(*testing.T) {
	go CompareJSON(json.RawMessage(`"a"`), json.RawMessage(`"b"`))
	go CompareJSON(json.RawMessage(`"a"`), json.RawMessage(`"b"`))
}
