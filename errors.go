package kivik

import (
	"net/http"

	"github.com/flimzy/kivik/errors"
)

type kivikError string

func (e kivikError) Error() string {
	return string(e)
}

func (e kivikError) StatusCode() int {
	switch e {
	case ErrNotImplemented:
		return StatusNotImplemented
	case ErrUnauthorized:
		return StatusUnauthorized
	case ErrNotFound:
		return StatusNotFound
	default:
		return 0
	}
}

// ErrNotImplemented is returned as an error if the underlying driver does not
// implement an optional method.
const ErrNotImplemented kivikError = "kivik: method not implemented by driver or backend"

// ErrUnauthorized is a generic Unauthorized error.
const ErrUnauthorized kivikError = "unauthorized"

// ErrNotFound is a generic Not Found error.
const ErrNotFound kivikError = "not found"

// ErrForbidden returns true if the error is the result of an HTTP 403/Forbidden
// response.
func ErrForbidden(err error) bool {
	return errors.StatusCode(err) == http.StatusForbidden
}
