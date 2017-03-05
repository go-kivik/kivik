// Package errors provides convenience functions for Kivik drivers to report
// meaningful errors.
package errors

import (
	"fmt"

	"github.com/pkg/errors"
)

// HTTP response codes permitted by the CouchDB API.
// See http://docs.couchdb.org/en/1.6.1/api/basics.html#http-status-codes
const (
	StatusNoError                      = 0
	StatusOK                           = 200
	StatusCreated                      = 201
	StatusAccepted                     = 202
	StatusNotModified                  = 304
	StatusBadRequest                   = 400
	StatusUnauthorized                 = 401
	StatusForbidden                    = 403
	StatusNotFound                     = 404
	StatusResourceNotAllowed           = 405
	StatusConflict                     = 409
	StatusPreconditionFailed           = 412
	StatusBadContentType               = 415
	StatusRequestedRangeNotSatisfiable = 416
	StatusExpectationFailed            = 417
	StatusInternalServerError          = 500
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
	return StatusNoError
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

type wrappedError struct {
	err        error
	statusCode int
}

func (e *wrappedError) Error() string {
	return e.err.Error()
}

func (e *wrappedError) StatusCode() int {
	return e.statusCode
}

// WrapStatus bundles an existing error with a status code.
func WrapStatus(status int, err error) error {
	if err == nil {
		return nil
	}
	return &wrappedError{
		err:        err,
		statusCode: status,
	}
}

// Wrap is a wrapper around pkg/errors.Wrap()
func Wrap(err error, msg string) error {
	return errors.Wrap(err, msg)
}

// Wrapf is a wrapper around pkg/errors.Wrapf()
func Wrapf(err error, format string, args ...interface{}) error {
	return errors.Wrapf(err, format, args...)
}

// Cause is a wrapper around pkg/errors.Cause()
func Cause(err error) error {
	return errors.Cause(err)
}

// Errorf is a wrapper around pkg/errors.Errorf()
func Errorf(format string, args ...interface{}) error {
	return errors.Errorf(format, args...)
}
