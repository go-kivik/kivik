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
	"net/http"

	"github.com/go-kivik/kivik/v4/internal"
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

	// Various not-implemented errors, that are returned, but don't need to be exposed directly.
	findNotImplemented
	clusterNotImplemented
	openRevsNotImplemented
	securityNotImplemented
	configNotImplemented
	replicationNotImplemented
)

const (
	errClientClosedText           = "client closed"
	errDatabaseClosedText         = "database closed"
	findNotImplementedText        = "kivik: driver does not support Find interface"
	clusterNotImplementedText     = "kivik: driver does not support cluster operations"
	openRevsNotImplementedText    = "kivik: driver does not support OpenRevs interface"
	securityNotImplementedText    = "kivik: driver does not support Security interface"
	configNotImplementedText      = "kivik: driver does not support Config interface"
	replicationNotImplementedText = "kivik: driver does not support replication"
)

func (e err) Error() string {
	switch e {
	case ErrClientClosed:
		return errClientClosedText
	case ErrDatabaseClosed:
		return errDatabaseClosedText
	case findNotImplemented:
		return findNotImplementedText
	case clusterNotImplemented:
		return clusterNotImplementedText
	case openRevsNotImplemented:
		return openRevsNotImplementedText
	case securityNotImplemented:
		return securityNotImplementedText
	case configNotImplemented:
		return configNotImplementedText
	case replicationNotImplemented:
		return replicationNotImplementedText
	}
	return "kivik: unknown error"
}

func (e err) HTTPStatus() int {
	switch e {
	case ErrClientClosed, ErrDatabaseClosed:
		return http.StatusServiceUnavailable
	case findNotImplemented, clusterNotImplemented, openRevsNotImplemented,
		securityNotImplemented, configNotImplemented, replicationNotImplemented:
		return http.StatusNotImplemented
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
	return internal.HTTPStatus(err)
}
