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

func TestFieldToJSONPath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple field",
			input: "name",
			want:  `$."name"`,
		},
		{
			name:  "nested field",
			input: "address.city",
			want:  `$."address"."city"`,
		},
		{
			name:  "escaped dot",
			input: `foo\.bar`,
			want:  `$."foo.bar"`,
		},
		{
			name:  "mixed nested and escaped",
			input: `address.foo\.bar.zip`,
			want:  `$."address"."foo.bar"."zip"`,
		},
		{
			name:  "double quote in field name",
			input: `say"hello`,
			want:  `$."say\"hello"`,
		},
		{
			name:  "backslash in field name",
			input: `foo\\bar`,
			want:  `$."foo\\bar"`,
		},
		{
			name:  "dot and double quote",
			input: `foo\.b"ar`,
			want:  `$."foo.b\"ar"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FieldToJSONPath(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractIndexFields(t *testing.T) {
	tests := []struct {
		name     string
		indexDef []byte
		want     []string
		wantErr  string
	}{
		{
			name:     "string fields",
			indexDef: []byte(`{"fields":["name","age"]}`),
			want:     []string{"name", "age"},
		},
		{
			name:     "object fields with direction",
			indexDef: []byte(`{"fields":[{"name":"asc"},{"age":"desc"}]}`),
			want:     []string{"name", "age"},
		},
		{
			name:     "mixed string and object fields",
			indexDef: []byte(`{"fields":["name",{"age":"desc"}]}`),
			want:     []string{"name", "age"},
		},
		{
			name:     "invalid JSON",
			indexDef: []byte(`not json`),
			wantErr:  "invalid character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractIndexFields(tt.indexDef)
			if tt.wantErr != "" {
				if err == nil || !containsString(err.Error(), tt.wantErr) {
					t.Errorf("unexpected error: got %v, want match for %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if d := cmp.Diff(tt.want, got); d != "" {
				t.Errorf("unexpected result:\n%s", d)
			}
		})
	}
}

func TestNormalizeIndexFields(t *testing.T) {
	tests := []struct {
		name     string
		indexDef string
		want     []map[string]string
		wantErr  string
	}{
		{
			name:     "string fields",
			indexDef: `{"fields":["name","age"]}`,
			want:     []map[string]string{{"name": "asc"}, {"age": "asc"}},
		},
		{
			name:     "object fields",
			indexDef: `{"fields":[{"name":"asc"},{"age":"desc"}]}`,
			want:     []map[string]string{{"name": "asc"}, {"age": "desc"}},
		},
		{
			name:     "invalid JSON",
			indexDef: `not json`,
			wantErr:  "invalid character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeIndexFields(tt.indexDef)
			if tt.wantErr != "" {
				if err == nil || !containsString(err.Error(), tt.wantErr) {
					t.Errorf("unexpected error: got %v, want match for %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if d := cmp.Diff(tt.want, got); d != "" {
				t.Errorf("unexpected result:\n%s", d)
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
