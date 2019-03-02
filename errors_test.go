package kivik

import (
	"errors"
	"testing"
)

func TestStatusCoder(t *testing.T) {
	type scTest struct {
		Name     string
		Err      error
		Expected int
	}
	tests := []scTest{
		{
			Name:     "nil",
			Expected: 0,
		},
		{
			Name:     "Standard error",
			Err:      errors.New("foo"),
			Expected: 500,
		},
		{
			Name:     "StatusCoder",
			Err:      &Error{HTTPStatus: 400, Err: errors.New("bad request")},
			Expected: 400,
		},
	}
	for _, test := range tests {
		func(test scTest) {
			t.Run(test.Name, func(t *testing.T) {
				result := StatusCode(test.Err)
				if result != test.Expected {
					t.Errorf("Unexpected result. Expected %d, got %d", test.Expected, result)
				}
			})
		}(test)
	}
}
