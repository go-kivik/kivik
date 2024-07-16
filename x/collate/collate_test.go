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

package collate

import (
	"math/rand"
	"sort"
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
