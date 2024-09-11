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
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSplitKeys(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{
			input: "foo.bar.baz",
			want:  []string{"foo", "bar", "baz"},
		},
		{
			input: "foo",
			want:  []string{"foo"},
		},
		{
			input: "",
			want:  []string{""},
		},
		{
			input: "foo\\.bar",
			want:  []string{"foo.bar"},
		},
		{
			input: "foo\\\\.bar",
			want:  []string{"foo\\", "bar"},
		},
		{
			input: "foo\\",
			want:  []string{"foo\\"},
		},
		{
			input: "foo.",
			want:  []string{"foo", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := SplitKeys(tt.input)
			if d := cmp.Diff(tt.want, got); d != "" {
				t.Errorf("unexpected keys (-want, +got): %s", d)
			}
		})
	}
}
