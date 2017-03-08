// +build go1.7,!go1.8

package kt

import "testing"

func TestTName(t *testing.T) {
	name := tName(t)
	if name != "TestTName" {
		t.Errorf("tName() returned '%s', not '%s'", name, "TestTName")
	}
}
