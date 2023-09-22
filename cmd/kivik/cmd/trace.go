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

package cmd

import (
	"bytes"
	"net/http"
	"net/http/httputil"

	"github.com/go-kivik/kivik/v4/couchdb/chttp"
)

func (r *root) clientTrace() *chttp.ClientTrace {
	r.trace = &chttp.ClientTrace{}
	return r.trace
}

func (r *root) setTrace() {
	if r.verbose {
		r.trace.HTTPRequestBody = r.traceHTTPRequestBody
		r.trace.HTTPResponseBody = r.traceHTTPResponseBody
		return
	}
	if r.dumpHeader {
		r.trace.HTTPResponse = r.traceHTTPResponse
	}
}

func (r *root) captureResponseHeader(h *http.Header) {
	orig := r.trace.HTTPResponse
	r.trace.HTTPResponse = func(resp *http.Response) {
		*h = resp.Header
		if orig != nil {
			orig(resp)
		}
	}
}

func (r *root) traceHTTPResponse(resp *http.Response) {
	dump, _ := httputil.DumpResponse(resp, false)
	for _, line := range bytes.Split(dump, []byte("\n")) {
		if len(line) > 0 {
			r.log.Infof("< %s", string(line))
		}
	}
}

func (r *root) traceHTTPRequestBody(req *http.Request) {
	dump, _ := httputil.DumpRequest(req, true)
	for _, line := range bytes.Split(dump, []byte("\n")) {
		if len(line) > 0 {
			r.log.Infof("> %s", string(line))
		}
	}
}

func (r *root) traceHTTPResponseBody(resp *http.Response) {
	dump, _ := httputil.DumpResponse(resp, true)
	for _, line := range bytes.Split(dump, []byte("\n")) {
		if len(line) > 0 {
			r.log.Infof("< %s", string(line))
		}
	}
}
