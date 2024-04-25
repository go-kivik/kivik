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

//go:build !js
// +build !js

package sqlite

import (
	"encoding/json"
	"math/rand"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_couchdbCmp(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			name: "js types", // See https://docs.couchdb.org/en/stable/ddocs/views/collation.html#collation-specification
			want: []string{
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
				`{"b":2, "a":1}`, // Member order does matter for collation. CouchDB preserves member order but doesn't require that clients will. this test might fail if used with a js engine that doesn't preserve order.
				`{"b":2, "c":2}`,
			},
		},
		{
			name: "7-bit ASCII",
			want: []string{
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Shuffle the input
			input := make([]json.RawMessage, len(tt.want))
			for i, v := range tt.want {
				input[i] = json.RawMessage(v)
			}
			rand.Shuffle(len(input), func(i, j int) { input[i], input[j] = input[j], input[i] })

			sort.Sort(couchdbKeys(input))

			got := make([]string, len(input))
			for i, v := range input {
				got[i] = string(v)
			}
			if d := cmp.Diff(tt.want, got); d != "" {
				t.Errorf("Unexpected result:\n%s", d)
			}
		})
	}
}
