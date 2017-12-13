package driver

import (
	"encoding/json"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/testy"
)

func TestChangesUnmarshal(t *testing.T) {
	input := `[
                {"rev": "6-460637e73a6288cb24d532bf91f32969"},
                {"rev": "5-eeaa298781f60b7bcae0c91bdedd1b87"}
            ]`
	var changes ChangedRevs
	if err := json.Unmarshal([]byte(input), &changes); err != nil {
		t.Fatalf("unmarshal failed: %s", err)
	}
	if len(changes) != 2 {
		t.Errorf("Expected 2 results, got %d", len(changes))
	}
	expected := []string{"6-460637e73a6288cb24d532bf91f32969", "5-eeaa298781f60b7bcae0c91bdedd1b87"}
	if d := diff.AsJSON(expected, changes); d != nil {
		t.Errorf("Results differ from expected:\n%s\n", d)
	}
}

func TestSequenceIDUnmarshal(t *testing.T) {
	tests := []struct {
		name  string
		input string

		expected SequenceID
		err      string
	}{
		{
			name:     "Couch 1.6",
			input:    "123",
			expected: "123",
		},
		{
			name:     "Couch 2.0",
			input:    `"1-seqfoo"`,
			expected: "1-seqfoo",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var seq SequenceID
			err := json.Unmarshal([]byte(test.input), &seq)
			testy.Error(t, test.err, err)
			if seq != test.expected {
				t.Errorf("Unexpected result: %s", seq)
			}
		})
	}
}
