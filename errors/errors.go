// Package errors provides convenience functions for Kivik drivers to report
// meaningful errors.
package errors

import "fmt"

// StatusError is an error message bundled with an HTTP status code.
type StatusError struct {
	StatusCode int
	Message    string
}

func (se *StatusError) Error() string {
	return fmt.Sprintf("error status %d: %s", se.StatusCode, se.Message)
}

// Status returns a new error with the designated HTTP status.
func Status(status int, msg string) error {
	return &StatusError{
		StatusCode: status,
		Message:    msg,
	}
}

// Statusf returns a new error with the designated HTTP status.
func Statusf(status int, format string, args ...interface{}) error {
	return &StatusError{
		StatusCode: status,
		Message:    fmt.Sprintf(format, args...),
	}
}
