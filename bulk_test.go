package kivik

import (
	"testing"

	"github.com/flimzy/diff"
)

func TestDocsInterfaceSlice(t *testing.T) {
	type diTest struct {
		name     string
		input    interface{}
		expected interface{}
		error    string
	}
	str := "foo"
	intSlice := []int{1, 2, 3}
	tests := []diTest{
		{
			name:     "Nil",
			input:    nil,
			expected: nil,
			error:    "must be slice or array, got <nil>",
		},
		{
			name:     "InterfaceSlice",
			input:    []interface{}{map[string]string{"foo": "bar"}},
			expected: []interface{}{map[string]string{"foo": "bar"}},
		},
		{
			name:  "String",
			input: "foo",
			error: "must be slice or array, got string",
		},
		{
			name:     "IntSlice",
			input:    []int{1, 2, 3},
			expected: []interface{}{1, 2, 3},
		},
		{
			name:     "IntArray",
			input:    [3]int{1, 2, 3},
			expected: []interface{}{1, 2, 3},
		},
		{
			name:  "StringPointer",
			input: &str,
			error: "must be slice or array, got *string",
		},
		{
			name:     "SlicePointer",
			input:    &intSlice,
			expected: []interface{}{1, 2, 3},
		},
		{
			name: "JSONDoc",
			input: []interface{}{
				map[string]string{"foo": "bar"},
				[]byte(`{"foo":"bar"}`),
			},
			expected: []interface{}{
				map[string]string{"foo": "bar"},
				map[string]string{"foo": "bar"},
			},
		},
	}
	for _, test := range tests {
		func(test diTest) {
			t.Run(test.name, func(t *testing.T) {
				result, err := docsInterfaceSlice(test.input)
				var msg string
				if err != nil {
					msg = err.Error()
				}
				if msg != test.error {
					t.Errorf("Unexpected error: %s", err)
				}
				if d := diff.AsJSON(test.expected, result); d != "" {
					t.Errorf("%s", d)
				}
			})
		}(test)
	}
}
