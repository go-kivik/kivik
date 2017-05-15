package bindings

import (
	"testing"

	"github.com/gopherjs/gopherjs/js"
)

type statuser interface {
	StatusCode() int
}

func TestNewPouchError(t *testing.T) {
	type npeTest struct {
		Name           string
		Object         *js.Object
		ExpectedStatus int
		Expected       string
	}
	tests := []npeTest{
		{
			Name:     "Null",
			Object:   nil,
			Expected: "",
		},
		{
			Name: "NameAndReasonNoStatus",
			Object: func() *js.Object {
				o := js.Global.Get("Object").New()
				o.Set("reason", "error reason")
				o.Set("name", "error name")
				return o
			}(),
			ExpectedStatus: 500,
			Expected:       "error name: error reason",
		},
		{
			Name: "ECONNREFUSED",
			Object: func() *js.Object {
				o := js.Global.Get("Object").New()
				o.Set("code", "ECONNREFUSED")
				o.Set("errno", "ECONNREFUSED")
				o.Set("syscall", "connect")
				o.Set("address", "127.0.0.1")
				o.Set("port", "5984")
				o.Set("status", "500")
				return o
			}(),
			ExpectedStatus: 500,
			Expected:       "connect: connection refused",
		},
	}
	for _, test := range tests {
		func(test npeTest) {
			t.Run(test.Name, func(t *testing.T) {
				result := NewPouchError(test.Object)
				var msg string
				if result != nil {
					msg = result.Error()
				}
				if msg != test.Expected {
					t.Errorf("Expected error: %s\n  Actual error: %s", test.Expected, msg)
				}
				if result == nil {
					return
				}
				status := result.(statuser).StatusCode()
				if status != test.ExpectedStatus {
					t.Errorf("Expected status %d, got %d", test.ExpectedStatus, status)
				}
			})
		}(test)
	}
}
