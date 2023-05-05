package kivik

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Error represents an error returned by Kivik.
//
// This type definition is not guaranteed to remain stable, or even exported.
// When examining errors programatically, you should rely instead on the
// StatusCode() function in this package, rather than on directly observing
// the fields of this type.
type Error struct {
	// HTTPStatus is the HTTP status code associated with this error. Normally
	// this is the actual HTTP status returned by the server, but in some cases
	// it may be generated by Kivik directly. Check the FromServer value if
	// the distinction matters to you.
	HTTPStatus int

	// FromServer is set to true if the error was returned by the server.
	// This field is deprecated and will soon be removed.
	FromServer bool

	// Message is the error message.
	Message string

	// Err is the originating error, if any.
	Err error
}

var (
	_ error       = &Error{}
	_ statusCoder = &Error{}
	_ causer      = &Error{}
)

func (e *Error) Error() string {
	if e.Err == nil {
		return e.msg()
	}
	if e.Message == "" {
		return e.Err.Error()
	}
	return e.Message + ": " + e.Err.Error()
}

// StatusCode returns the HTTP status code associated with the error, or 500
// (internal server error), if none.
func (e *Error) StatusCode() int {
	if e.HTTPStatus == 0 {
		return http.StatusInternalServerError
	}
	return e.HTTPStatus
}

// Cause satisfies the github.com/pkg/errors.causer interface by returning e.Err.
func (e *Error) Cause() error {
	return e.Err
}

// Unwrap satisfies the errors.Wrapper interface.
func (e *Error) Unwrap() error {
	return e.Err
}

// Format implements fmt.Formatter
func (e *Error) Format(f fmt.State, c rune) {
	parts := make([]string, 0, 3)
	if e.Message != "" {
		parts = append(parts, e.Message)
	}
	switch c {
	case 'v':
		if f.Flag('+') {
			var prefix string
			if e.FromServer {
				prefix = "server responded with"
			} else {
				prefix = "kivik generated"
			}
			parts = append(parts, fmt.Sprintf("%s %d / %s", prefix, e.HTTPStatus, http.StatusText(e.HTTPStatus)))
		}
	}
	if e.Err != nil {
		parts = append(parts, e.Err.Error())
	}
	_, _ = fmt.Fprint(f, strings.Join(parts, ": "))
}

func (e *Error) msg() string {
	switch e.Message {
	case "":
		return http.StatusText(e.StatusCode())
	default:
		return e.Message
	}
}

type statusCoder interface {
	StatusCode() int
}

type causer interface {
	Cause() error
}

// StatusCode returns the HTTP status code embedded in the error, or 500
// (internal server error), if there was no specified status code.  If err is
// nil, StatusCode returns 0. This provides a convenient way to determine the
// precise nature of a Kivik-returned error.
//
// For example, to panic for all but NotFound errors:
//
//	err := db.Get(context.TODO(), "docID").ScanDoc(&doc)
//	if kivik.StatusCode(err) == kivik.StatusNotFound {
//	    return
//	}
//	if err != nil {
//	    panic(err)
//	}
//
// This method uses the statusCoder interface, which is not exported by this
// package, but is considered part of the stable public API.  Driver
// implementations are expected to return errors which conform to this
// interface.
//
//	type statusCoder interface {
//	    StatusCode() (httpStatusCode int)
//	}
func StatusCode(err error) int {
	if err == nil {
		return 0
	}
	var coder statusCoder
	for {
		if errors.As(err, &coder) {
			return coder.StatusCode()
		}
		if uw := errors.Unwrap(err); uw != nil {
			err = uw
			continue
		}
		if c, ok := err.(causer); ok {
			err = c.Cause()
			continue
		}
		return http.StatusInternalServerError
	}
}
