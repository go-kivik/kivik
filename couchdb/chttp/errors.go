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
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
)

// HTTPError is an error that represents an HTTP transport error.
type HTTPError struct {
	// Response is the HTTP response received by the client.  The response body
	// should already be closed, but the response and request headers and other
	// metadata will typically be in tact for debugging purposes.
	Response *http.Response `json:"-"`

	// Reason is the server-supplied error reason.
	Reason string `json:"reason"`
}

func (e *HTTPError) Error() string {
	if e.Reason == "" {
		return http.StatusText(e.HTTPStatus())
	}
	if statusText := http.StatusText(e.HTTPStatus()); statusText != "" {
		return fmt.Sprintf("%s: %s", statusText, e.Reason)
	}
	return e.Reason
}

// HTTPStatus returns the embedded status code.
func (e *HTTPError) HTTPStatus() int {
	return e.Response.StatusCode
}

// ResponseError returns an error from an *http.Response if the status code
// indicates an error.
func ResponseError(resp *http.Response) error {
	if resp.StatusCode < 400 { // nolint:gomnd
		return nil
	}
	if resp.Body != nil {
		defer CloseBody(resp.Body)
	}
	httpErr := &HTTPError{
		Response: resp,
	}
	if resp != nil && resp.Request != nil && resp.Request.Method != "HEAD" && resp.ContentLength != 0 {
		if ct, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type")); ct == typeJSON {
			_ = json.NewDecoder(resp.Body).Decode(httpErr)
		}
	}
	return httpErr
}
