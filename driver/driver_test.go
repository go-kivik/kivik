package driver

import (
	"fmt"
	"testing"
)

func TestMD5sumStringer(t *testing.T) {
	var md5sum = MD5sum{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	fmt.Printf("X: %s\n", md5sum)
	result := fmt.Sprintf("%s", md5sum)
	expected := "0102030405060708090a0b0c0d0e0f10"
	if result != expected {
		t.Errorf("Unexpected result: %s", result)
	}
}
