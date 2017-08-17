package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/flimzy/diff"
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

func TestErrorJSON(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "StatusError not found",
			err:      &StatusError{statusCode: http.StatusNotFound, message: "no_db_file"},
			expected: `{"error":"not_found", "reason":"no_db_file"}`,
		},
		{
			name:     "StatusError unknown code",
			err:      &StatusError{statusCode: 999, message: "somethin' bad happened"},
			expected: `{"error":"unknown", "reason": "somethin' bad happened"}`,
		},
		{
			name:     "StatusError unauthorized",
			err:      &StatusError{statusCode: http.StatusUnauthorized, message: "You are not a server admin."},
			expected: `{"error":"unauthorized", "reason":"You are not a server admin."}`,
		},
		{
			name:     "StatusError precondition failed",
			err:      &StatusError{statusCode: http.StatusPreconditionFailed, message: "The database could not be created, the file already exists."},
			expected: `{"error":"precondition_failed", "reason":"The database could not be created, the file already exists."}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := json.Marshal(test.err)
			if err != nil {
				t.Fatal(err)
			}
			if d := diff.JSON([]byte(test.expected), result); d != nil {
				t.Error(d)
			}
		})
	}
}
