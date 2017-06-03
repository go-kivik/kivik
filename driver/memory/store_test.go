package memory

import "testing"

func TestRandStr(t *testing.T) {
	str := randStr()
	if len(str) != 32 {
		t.Errorf("Expected 32-char string, got %d", len(str))
	}
}
