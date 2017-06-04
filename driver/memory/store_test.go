package memory

import (
	"strings"
	"testing"
)

func TestRandStr(t *testing.T) {
	str := randStr()
	if len(str) != 32 {
		t.Errorf("Expected 32-char string, got %d", len(str))
	}
}

func TestAddRevision(t *testing.T) {
	d := &database{
		docs: make(map[string]*document),
	}
	r := d.addRevision("foo", jsondoc{"_id": "bar"})
	if !strings.HasPrefix(r, "1-") {
		t.Errorf("Expected initial revision to start with '1-', but got '%s'", r)
	}
	if len(r) != 34 {
		t.Errorf("rev (%s) is %d chars long, expected 34", r, len(r))
	}
	r = d.addRevision("foo", jsondoc{"_id": "bar"})
	if !strings.HasPrefix(r, "2-") {
		t.Errorf("Expected second revision to start with '2-', but got '%s'", r)
	}
	if len(r) != 34 {
		t.Errorf("rev (%s) is %d chars long, expected 34", r, len(r))
	}
}
