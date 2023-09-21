// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
)

// Exit status codes
//
// See https://man.openbsd.org/sysexits.3
const (
	// ErrUsageError indicates an incorrect command, option, or unparseable
	// configuration or command line options.
	ErrUsage = 2
	// ErrUnknown indicates that the server responded with an HTTP status > 500.
	// Probably an indication of a proxy server interfering.
	ErrUnknown = 3
	// ErrInternalServerError indicates that the server responded with a 500
	// error.
	ErrInternalServerError = 4

	// ErrBadReuqest indicates that the server responded with a 400 error.
	ErrBadRequest = 10
	// ErrUnauthorized indicates that the server responded with a 401 error.
	ErrUnauthorized = 11
	// ErrForbidden indicates that the server responded with a 403 error.
	ErrForbidden = 13
	// ErrNotFound indicates that the server responded with a 404 error.
	ErrNotFound = 14
	// ErrMethodNotAllowed indicates that the server responded with a 405 error.
	ErrMethodNotAllowed = 15
	// ErrNotAcceptable indicates that the server responded with a 406 error.
	ErrNotAcceptable = 16
	// ErrConflict indicates that the server responded with a 409 error.
	ErrConflict = 19
	// ErrPreconditionFailed indicates that the server responded with a 412
	// error.
	ErrPreconditionFailed = 22
	// ErrRequestEntityTooLarge indicates that the server responded with a 413
	// error.
	ErrRequestEntityTooLarge = 23
	// ErrUnsupportedMediaType indicates that the server responded with a 415
	// error.
	ErrUnsupportedMediaType = 25
	// ErrRequestedRangeNotSatisfiable indicates that the server responded with
	// a 416 error.
	ErrRequestedRangeNotSatisfiable = 26
	// ErrExpectationFailed indicates that the server responded with a 417 error.
	ErrExpectationFailed = 27

	// ErrData indicates an input file is invalid, such as malformed JSON or
	// YAML.
	ErrData = 65
	// ErrNoInput indicates that an input file does not exist or cannot be read.
	ErrNoInput = 66
	// ErrNoHost indicates that the host does not exist or cannot be looked up.
	ErrNoHost = 67
	// ErrUnavailable indicates that the server could not be reached, such as
	// a connection refused.
	ErrUnavailable = 69
	// ErrCantCreate indicates that an output file cannot be created.
	ErrCantCreate = 73
	// ErrIO indicates an I/O error while reading from or writing to a file or
	// the network.
	ErrIO = 74
	// ErrProtocol indicates a protocol error, such as a CouchDB server
	// returning a non-JSON response.
	ErrProtocol = 76
)

type statusErr struct {
	error
	code int
}

func (e *statusErr) Error() string {
	return e.error.Error()
}

func (e *statusErr) Unwrap() error {
	return e.error
}

func (e *statusErr) ExitStatus() int {
	return e.code
}

func WithCode(err error, code int) error {
	return &statusErr{
		error: err,
		code:  code,
	}
}

// New calls errors.New.
func New(text string) error {
	return errors.New(text)
}

func InspectErrorCode(err error) int {
	if err == nil {
		return 0
	}
	exitErr := new(statusErr)
	if errors.As(err, &exitErr) {
		return exitErr.ExitStatus()
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return ErrUnavailable
	}

	jsonSyntax := new(json.SyntaxError)
	if errors.As(err, &jsonSyntax) {
		return ErrProtocol
	}

	var kivikErr interface {
		HTTPStatus() int
	}
	if errors.As(err, &kivikErr) {
		return fromHTTPStatus(kivikErr.HTTPStatus())
	}

	return 0
}

func fromHTTPStatus(status int) int {
	switch {
	case status == http.StatusInternalServerError:
		return ErrInternalServerError
	case status >= 400 && status < 500:
		return status - 390 // nolint:gomnd
	default:
		return ErrUnknown
	}
}

// HTTPStatus converts status to an error code, and passes it to Code().
func HTTPStatus(status int, err ...interface{}) error {
	return Code(fromHTTPStatus(status), err...)
}

// HTTPStatusf converts status to an error code, and passes it to Codef().
func HTTPStatusf(status int, format string, args ...interface{}) error {
	return Codef(fromHTTPStatus(status), format, args...)
}

// Code returns a new error with an error code. If err is an existing error, it
// is wrapped with the error code. All other values are passed to fmt.Sprint.
//
// If err is a single nil value, nil is returned.
func Code(code int, err ...interface{}) error {
	if len(err) == 1 {
		if err[0] == nil {
			return nil
		}
		if e, ok := err[0].(error); ok {
			return &statusErr{
				error: e,
				code:  code,
			}
		}
	}
	return &statusErr{
		error: errors.New(fmt.Sprint(err...)),
		code:  code,
	}
}

// Codef wraps the output of fmt.Errorf with a code.
func Codef(code int, format string, args ...interface{}) error {
	return &statusErr{
		error: fmt.Errorf(format, args...),
		code:  code,
	}
}

// As calls errors.As.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Is calls errors.Is.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// Unwrap calls errors.Unwrap.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}
