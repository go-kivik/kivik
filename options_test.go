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

package kivik

import (
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"
)

func Test_multiOptions_String(t *testing.T) {
	t.Parallel()

	type tt struct {
		input multiOptions
		want  string
	}

	tests := testy.NewTable()

	tests.Add("nil element in slice", tt{
		input: multiOptions{Param("foo", "bar"), nil, Param("baz", "qux")},
		want:  "map[foo:bar],map[baz:qux]",
	})

	tests.Run(t, func(t *testing.T, test tt) {
		got := test.input.String()
		if d := cmp.Diff(test.want, got); d != "" {
			t.Errorf("Unexpected result (-want, +got):\n%s", d)
		}
	})
}

func TestDuration(t *testing.T) {
	o := Duration("heartbeat", 15*time.Second)

	query := url.Values{}
	o.Apply(&query)
	want := "15000"
	if got := query.Get("heartbeat"); got != want {
		t.Errorf("Unexpected url query value: %s", got)
	}

	opts := map[string]any{}
	o.Apply(opts)
	if got := opts["heartbeat"]; got != want {
		t.Errorf("Unexpected map value: %s", got)
	}
}

func Test_params_Apply(t *testing.T) {
	tests := []struct {
		name   string
		target any
		want   any
		option params
	}{
		{
			name:   "no options",
			target: map[string]any{},
			want:   map[string]any{},
		},
		{
			name:   "map option",
			target: map[string]any{},
			want:   map[string]any{"foo": "bar"},
			option: params{"foo": "bar"},
		},
		{
			name:   "unsupported target type",
			target: "",
			want:   "",
			option: params{"foo": "bar"},
		},
		{
			name:   "query, string",
			target: &url.Values{},
			want:   &url.Values{"foo": []string{"bar"}},
			option: params{"foo": "bar"},
		},
		{
			name:   "query, bool",
			target: &url.Values{},
			want:   &url.Values{"foo": []string{"true"}},
			option: params{"foo": true},
		},
		{
			name:   "query, int",
			target: &url.Values{},
			want:   &url.Values{"foo": []string{"42"}},
			option: params{"foo": 42},
		},
		{
			name:   "query, float64",
			target: &url.Values{},
			want:   &url.Values{"foo": []string{"42.5"}},
			option: params{"foo": 42.5},
		},
		{
			name:   "query, float32",
			target: &url.Values{},
			want:   &url.Values{"foo": []string{"42.5"}},
			option: params{"foo": float32(42.5)},
		},
		{
			name:   "query, []string",
			target: &url.Values{},
			want:   &url.Values{"foo": []string{"bar", "baz"}},
			option: params{"foo": []string{"bar", "baz"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.option.Apply(tt.target)
			if d := cmp.Diff(tt.want, tt.target); d != "" {
				t.Errorf("Unexpected result (-want, +got):\n%s", d)
			}
		})
	}
}
