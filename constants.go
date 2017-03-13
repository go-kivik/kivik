package kivik

import "net/http"

// SessionCookieName is the name of the CouchDB session cookie.
const SessionCookieName = "AuthSession"

// Media types commonly used by CouchDB
const (
	TypeJSON = "application/json"
	TypeText = "text/plain"
)

// HTTP methods supported by CouchDB. This is almost an exact copy of the
// methods in the standard http package, with the addition of MethodCopy, and
// a few methods left out which are not used by CouchDB.
const (
	MethodGet    = http.MethodGet
	MethodHead   = http.MethodHead
	MethodPost   = http.MethodPost
	MethodPut    = http.MethodPut
	MethodDelete = http.MethodDelete
	MethodCopy   = "COPY"
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
