package kivik

import (
	"net/http"

	"github.com/flimzy/kivik/errors"
)

// HTTP response codes permitted by the CouchDB API.
// See http://docs.couchdb.org/en/1.6.1/api/basics.html#http-status-codes
const (
	StatusNoError                      = 0
	StatusOK                           = 200
	StatusCreated                      = 201
	StatusAccepted                     = 202
	StatusFound                        = 302
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
	StatusNotImplemented               = 501
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
