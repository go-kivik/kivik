// +build go1.7,!go1.8

package kt

import (
	"reflect"
	"testing"
	"unsafe"
)

// This method extracts the unexported 'name' field from a testing.T value.
//
// This is super evil. Don't try this at home.
//
// Adapted from http://stackoverflow.com/a/17982725/13860
func tName(t *testing.T) string {
	pv := reflect.ValueOf(t)
	v := reflect.Indirect(pv)
	name := v.FieldByName("name")
	namePtr := unsafe.Pointer(name.UnsafeAddr())
	realName := (*string)(namePtr)
	return *realName
}
