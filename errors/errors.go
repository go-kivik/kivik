// Package errors provides convenience functions for Kivik drivers to report
// meaningful errors.
package errors

import (
	"errors"
	"fmt"
)

// StatusError is an error message bundled with an HTTP status code.
type StatusError struct {
	statusCode int
	message    string
}

func (se *StatusError) Error() string {
	return fmt.Sprintf("error status %d: %s", se.statusCode, se.message)
}

// StatusCode returns the StatusError's embedded HTTP status code.
func (se *StatusError) StatusCode() int {
	return se.statusCode
}

// StatusCoder is an optional error interface, which returns the error's
// embedded HTTP status code.
type StatusCoder interface {
	StatusCode() int
}

// StatusCode extracts an embedded HTTP status code from an error.
func StatusCode(err error) int {
	if scErr, ok := err.(StatusCoder); ok {
		return scErr.StatusCode()
	}
	return 0
}

// New is a wrapper around the standard errors.New, to avoid the need for
// multiple imports.
func New(msg string) error {
	return errors.New(msg)
}

// Status returns a new error with the designated HTTP status.
func Status(status int, msg string) error {
	return &StatusError{
		statusCode: status,
		message:    msg,
	}
}

// Statusf returns a new error with the designated HTTP status.
func Statusf(status int, format string, args ...interface{}) error {
	return &StatusError{
		statusCode: status,
		message:    fmt.Sprintf(format, args...),
	}
}
