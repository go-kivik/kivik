package kivik

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/testy"
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

func TestFormatError(t *testing.T) {
	type tst struct {
		err  error
		str  string
		std  string
		full string
	}
	tests := testy.NewTable()
	tests.Add("standard error", tst{
		err:  errors.New("foo"),
		str:  "foo",
		std:  "foo",
		full: "foo",
	})
	tests.Add("not from server", tst{
		err: &Error{HTTPStatus: http.StatusNotFound, Err: errors.New("not found")},
		str: "not found",
		std: "Not Found: not found",
		full: `Not Found:
    kivik generated 404 / Not Found
  - not found`,
	})
	tests.Add("from server", tst{
		err: &Error{HTTPStatus: http.StatusNotFound, FromServer: true, Err: errors.New("not found")},
		str: "not found",
		std: "Not Found: not found",
		full: `Not Found:
    server responded with 404 / Not Found
  - not found`,
	})
	tests.Add("with message", tst{
		err: &Error{HTTPStatus: http.StatusNotFound, Message: "It's missing", Err: errors.New("not found")},
		str: "It's missing: not found",
		std: "It's missing: not found",
		full: `It's missing:
    kivik generated 404 / Not Found
  - not found`,
	})

	tests.Run(t, func(t *testing.T, test tst) {
		if d := diff.Text(test.str, test.err.Error()); d != nil {
			t.Errorf("Error():\n%s", d)
		}
		if d := diff.Text(test.std, fmt.Sprintf("%v", test.err)); d != nil {
			t.Errorf("Standard:\n%s", d)
		}
		if d := diff.Text(test.full, fmt.Sprintf("%+v", test.err)); d != nil {
			t.Errorf("Full:\n%s", d)
		}
	})
}
