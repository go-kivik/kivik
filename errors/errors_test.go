package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/flimzy/diff"
)

func TestStatusf(t *testing.T) {
	e := Statusf(400, "foo %d", 123)
	result := e.(*statusError)
	expected := &statusError{
		message:    "foo 123",
		statusCode: 400,
	}
	if d := diff.Interface(expected, result); d != nil {
		t.Error(d)
	}
}

func TestWrapStatus(t *testing.T) {
	e := WrapStatus(400, errors.New("foo"))
	expected := &wrappedError{
		err:        errors.New("foo"),
		statusCode: 400,
	}
	result := e.(*wrappedError)
	if d := diff.Interface(expected, result); d != nil {
		t.Error(d)
	}
}

func TestErrorJSON(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "statusError not found",
			err:      &statusError{statusCode: http.StatusNotFound, message: "no_db_file"},
			expected: `{"error":"not_found", "reason":"no_db_file"}`,
		},
		{
			name:     "statusError unknown code",
			err:      &statusError{statusCode: 999, message: "somethin' bad happened"},
			expected: `{"error":"unknown", "reason": "somethin' bad happened"}`,
		},
		{
			name:     "statusError unauthorized",
			err:      &statusError{statusCode: http.StatusUnauthorized, message: "You are not a server admin."},
			expected: `{"error":"unauthorized", "reason":"You are not a server admin."}`,
		},
		{
			name:     "statusError precondition failed",
			err:      &statusError{statusCode: http.StatusPreconditionFailed, message: "The database could not be created, the file already exists."},
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
