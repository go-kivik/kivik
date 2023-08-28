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

package chttp

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-kivik/couchdb/v4/internal"
	"github.com/go-kivik/kivik/v4"
)

// Options are optional parameters which may be sent with a request.
type Options struct {
	// Accept sets the request's Accept header. Defaults to "application/json".
	// To specify any, use "*/*".
	Accept string

	// ContentType sets the requests's Content-Type header. Defaults to "application/json".
	ContentType string

	// ContentLength, if set, sets the ContentLength of the request
	ContentLength int64

	// Body sets the body of the request.
	Body io.ReadCloser

	// GetBody is a function to set the body, and can be used on retries. If
	// set, Body is ignored.
	GetBody func() (io.ReadCloser, error)

	// JSON is an arbitrary data type which is marshaled to the request's body.
	// It an error to set both Body and JSON on the same request. When this is
	// set, ContentType is unconditionally set to 'application/json'. Note that
	// for large JSON payloads, it can be beneficial to do your own JSON stream
	// encoding, so that the request can be live on the wire during JSON
	// encoding.
	JSON interface{}

	// FullCommit adds the X-Couch-Full-Commit: true header to requests
	FullCommit bool

	// IfNoneMatch adds the If-None-Match header. The value will be quoted if
	// it is not already.
	IfNoneMatch string

	// Query is appended to the exiting url, if present. If the passed url
	// already contains query parameters, the values in Query are appended.
	// No merging takes place.
	Query url.Values

	// Header is a list of default headers to be set on the request.
	Header http.Header

	// NoGzip disables gzip compression on the request body.
	NoGzip bool
}

// NewOptions converts a kivik options map into
func NewOptions(opts map[string]interface{}) (*Options, error) {
	fullCommit, err := fullCommit(opts)
	if err != nil {
		return nil, err
	}
	ifNoneMatch, err := ifNoneMatch(opts)
	if err != nil {
		return nil, err
	}
	return &Options{
		FullCommit:  fullCommit,
		IfNoneMatch: ifNoneMatch,
	}, nil
}

func fullCommit(opts map[string]interface{}) (bool, error) {
	fc, ok := opts[internal.OptionFullCommit]
	if !ok {
		return false, nil
	}
	fcBool, ok := fc.(bool)
	if !ok {
		return false, &kivik.Error{Status: http.StatusBadRequest, Err: fmt.Errorf("kivik: option '%s' must be bool, not %T", internal.OptionFullCommit, fc)}
	}
	delete(opts, internal.OptionFullCommit)
	return fcBool, nil
}

func ifNoneMatch(opts map[string]interface{}) (string, error) {
	inm, ok := opts[internal.OptionIfNoneMatch]
	if !ok {
		return "", nil
	}
	inmString, ok := inm.(string)
	if !ok {
		return "", &kivik.Error{Status: http.StatusBadRequest, Err: fmt.Errorf("kivik: option '%s' must be string, not %T", internal.OptionIfNoneMatch, inm)}
	}
	delete(opts, internal.OptionIfNoneMatch)
	if inmString[0] != '"' {
		return `"` + inmString + `"`, nil
	}
	return inmString, nil
}
