package memory

import (
	"context"
	"testing"

	"github.com/flimzy/diff"
)

func TestIndexSpecUnmarshalJSON(t *testing.T) {
	type isuTest struct {
		name     string
		input    string
		expected *indexSpec
		err      string
	}
	tests := []isuTest{
		{
			name:     "ddoc only",
			input:    `"foo"`,
			expected: &indexSpec{ddoc: "foo"},
		},
		{
			name:     "ddoc and index",
			input:    `["foo","bar"]`,
			expected: &indexSpec{ddoc: "foo", index: "bar"},
		},
		{
			name:  "invalid json",
			input: "asdf",
			err:   "invalid character 'a' looking for beginning of value",
		},
		{
			name:  "extra fields",
			input: `["foo","bar","baz"]`,
			err:   "invalid index specification",
		},
		{
			name:     "One field",
			input:    `["foo"]`,
			expected: &indexSpec{ddoc: "foo"},
		},
		{
			name:  "Empty array",
			input: `[]`,
			err:   "invalid index specification",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := &indexSpec{}
			err := result.UnmarshalJSON([]byte(test.input))
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
			if d := diff.Interface(test.expected, result); d != "" {
				t.Errorf(d)
			}
		})
	}
}

func TestCreateIndex(t *testing.T) {
	d := &db{}
	err := d.CreateIndex(context.Background(), "foo", "bar", "baz")
	if err != errFindNotImplemented {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestGetIndexes(t *testing.T) {
	d := &db{}
	_, err := d.GetIndexes(context.Background())
	if err != errFindNotImplemented {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestDeleteIndex(t *testing.T) {
	d := &db{}
	err := d.DeleteIndex(context.Background(), "foo", "bar")
	if err != errFindNotImplemented {
		t.Errorf("Unexpected error: %s", err)
	}
}
