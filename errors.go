package kivik

import (
	"net/http"

	"github.com/flimzy/kivik/errors"
)

type kivikError string

func (e kivikError) Error() string {
	return string(e)
}

// NotImplemented is returned as an error if the underlying driver does not
// implement an optional method.
const NotImplemented kivikError = "kivik: method not implemented by driver"

// ErrNoAuthenticator is returned if Authenticate() is called, and no authenticator
// is set.
const ErrNoAuthenticator kivikError = "kivik: no authenticator set"

// ErrNotFound returns true if the error is the result of an HTTP 404/Not Found
// response.
func ErrNotFound(err error) bool {
	return errors.StatusCode(err) == http.StatusNotFound
}

// ErrForbidden returns true if the error is the result of an HTTP 403/Forbidden
// response.
func ErrForbidden(err error) bool {
	return errors.StatusCode(err) == http.StatusForbidden
}
