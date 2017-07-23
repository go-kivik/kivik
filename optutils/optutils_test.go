package optutils

import "testing"

func TestInt64(t *testing.T) {
	opts := map[string]interface{}{
		"int":            int(1235),
		"int8":           int8(123),
		"int16":          int16(1234),
		"int32":          int32(12345),
		"int64":          int64(123456),
		"uint":           uint(2235),
		"uint8":          uint8(223),
		"uint16":         uint16(2234),
		"uint32":         uint32(22345),
		"uint64":         uint64(223456),
		"uint64overflow": uint64(18446744073709551615),
	}
	type iTest struct {
		name     string
		key      string
		expected int64
		err      string
	}
	tests := []iTest{
		{
			key: "notfound",
			err: "key not found",
		},
		{
			key:      "int",
			expected: 1235,
		},
		{
			key:      "int8",
			expected: 123,
		},
		{
			key:      "int16",
			expected: 1234,
		},
		{
			key:      "int32",
			expected: 12345,
		},
		{
			key:      "int64",
			expected: 123456,
		},
		{
			key:      "uint",
			expected: 2235,
		},
		{
			key:      "uint8",
			expected: 223,
		},
		{
			key:      "uint16",
			expected: 2234,
		},
		{
			key:      "uint32",
			expected: 22345,
		},
		{
			key:      "uint64",
			expected: 223456,
		},
		{
			key: "uint64overflow",
			err: "",
		},
	}
	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			result, err := Int64(opts, test.key)
			var msg string
			if err != nil {
				msg = err.Error()
			}
			if msg != test.err {
				t.Errorf("Unexpected error: %s", msg)
			}
			if result != test.expected {
				t.Errorf("Unexpected result: %d", result)
			}
		})
	}
}
