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

package kivik

import (
	"errors"
	"net/http"
)

type err int

var (
	_ error       = err(0)
	_ statusCoder = err(0)
)

const (
	// ErrClientClosed is returned by any client operations after [Client.Close]
	// has been called.
	ErrClientClosed err = iota
	// ErrDatabaseClosed is returned by any database operations after [DB.Close]
	// has been called.
	ErrDatabaseClosed
)

const (
	errClientClosed   = "client closed"
	errDatabaseClosed = "database closed"
)

func (e err) Error() string {
	switch e {
	case ErrClientClosed:
		return errClientClosed
	case ErrDatabaseClosed:
		return errDatabaseClosed
	}
	return "unknown error"
}

func (e err) HTTPStatus() int {
	switch e {
	case ErrClientClosed, ErrDatabaseClosed:
		return http.StatusServiceUnavailable
	}
	return http.StatusInternalServerError
}

type statusCoder interface {
	HTTPStatus() int
}

// HTTPStatus returns the HTTP status code embedded in the error, or 500
// (internal server error), if there was no specified status code.  If err is
// nil, HTTPStatus returns 0. This provides a convenient way to determine the
// precise nature of a Kivik-returned error.
//
// For example, to panic for all but NotFound errors:
//
//	err := db.Get(context.TODO(), "docID").ScanDoc(&doc)
//	if kivik.HTTPStatus(err) == http.StatusNotFound {
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
//	    HTTPStatus() int
//	}
func HTTPStatus(err error) int {
	if err == nil {
		return 0
	}
	var coder statusCoder
	for {
		if errors.As(err, &coder) {
			return coder.HTTPStatus()
		}
		if uw := errors.Unwrap(err); uw != nil {
			err = uw
			continue
		}
		return http.StatusInternalServerError
	}
}
