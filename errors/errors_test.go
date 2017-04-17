package errors

import (
	"errors"
	"testing"
)

func TestErrors(t *testing.T) {
	type errTest struct {
		Name           string
		Func           func() error
		ExpectedStatus int
		ExpectedMsg    string
	}
	tests := []errTest{
		{
			Name:           "Statusf",
			Func:           func() error { return Statusf(500, "Testing %d", 123) },
			ExpectedStatus: 500,
			ExpectedMsg:    "Testing 123",
		},
		{
			Name:           "WrapStatus",
			Func:           func() error { return WrapStatus(500, errors.New("original error")) },
			ExpectedStatus: 500,
			ExpectedMsg:    "original error",
		},
	}
	for _, test := range tests {
		func(test errTest) {
			t.Run(test.Name, func(t *testing.T) {
				err := test.Func()
				if status := StatusCode(err); status != test.ExpectedStatus {
					t.Errorf("Status. Expected %d, Actual %d", test.ExpectedStatus, status)
				}
				if msg := err.Error(); msg != test.ExpectedMsg {
					t.Errorf("Error. Expected '%s', Actual '%s'", test.ExpectedMsg, msg)
				}
			})
		}(test)
	}
}
