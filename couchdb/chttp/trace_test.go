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
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab.com/flimzy/testy"
)

func TestHTTPResponse(t *testing.T) {
	tests := []struct {
		name      string
		trace     func(t *testing.T) *ClientTrace
		resp      *http.Response
		finalResp *http.Response
	}{
		{
			name:      "no hook defined",
			trace:     func(_ *testing.T) *ClientTrace { return &ClientTrace{} },
			resp:      &http.Response{StatusCode: 200},
			finalResp: &http.Response{StatusCode: 200},
		},
		{
			name: "HTTPResponseBody/cloned response",
			trace: func(t *testing.T) *ClientTrace {
				return &ClientTrace{
					HTTPResponseBody: func(r *http.Response) {
						if r.StatusCode != 200 {
							t.Errorf("Unexpected status code: %d", r.StatusCode)
						}
						r.StatusCode = 0
						defer r.Body.Close() // nolint: errcheck
						if _, err := io.ReadAll(r.Body); err != nil {
							t.Fatal(err)
						}
					},
				}
			},
			resp:      &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("testing"))},
			finalResp: &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("testing"))},
		},
		{
			name: "HTTPResponse/cloned response",
			trace: func(t *testing.T) *ClientTrace {
				return &ClientTrace{
					HTTPResponse: func(r *http.Response) {
						if r.StatusCode != 200 {
							t.Errorf("Unexpected status code: %d", r.StatusCode)
						}
						r.StatusCode = 0
						if r.Body != nil {
							t.Errorf("non-nil body")
						}
					},
				}
			},
			resp:      &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("testing"))},
			finalResp: &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("testing"))},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			trace := test.trace(t)
			trace.httpResponseBody(test.resp)
			trace.httpResponse(test.resp)
			if d := testy.DiffHTTPResponse(test.finalResp, test.resp); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestHTTPRequest(t *testing.T) {
	tests := []struct {
		name     string
		trace    func(t *testing.T) *ClientTrace
		req      *http.Request
		finalReq *http.Request
	}{
		{
			name:     "no hook defined",
			trace:    func(_ *testing.T) *ClientTrace { return &ClientTrace{} },
			req:      httptest.NewRequest("PUT", "/", io.NopCloser(strings.NewReader("testing"))),
			finalReq: httptest.NewRequest("PUT", "/", io.NopCloser(strings.NewReader("testing"))),
		},
		{
			name: "HTTPRequesteBody/cloned response",
			trace: func(t *testing.T) *ClientTrace {
				return &ClientTrace{
					HTTPRequestBody: func(r *http.Request) {
						if r.Method != "PUT" {
							t.Errorf("Unexpected method: %s", r.Method)
						}
						r.Method = "unf"     // nolint: goconst
						defer r.Body.Close() // nolint: errcheck
						if _, err := io.ReadAll(r.Body); err != nil {
							t.Fatal(err)
						}
					},
				}
			},
			req:      httptest.NewRequest("PUT", "/", io.NopCloser(strings.NewReader("testing"))),
			finalReq: httptest.NewRequest("PUT", "/", io.NopCloser(strings.NewReader("testing"))),
		},
		{
			name: "HTTPRequeste/cloned response",
			trace: func(t *testing.T) *ClientTrace {
				return &ClientTrace{
					HTTPRequest: func(r *http.Request) {
						if r.Method != "PUT" {
							t.Errorf("Unexpected method: %s", r.Method)
						}
						r.Method = "unf"
						if r.Body != nil {
							t.Errorf("non-nil body")
						}
					},
				}
			},
			req:      httptest.NewRequest("PUT", "/", io.NopCloser(strings.NewReader("testing"))),
			finalReq: httptest.NewRequest("PUT", "/", io.NopCloser(strings.NewReader("testing"))),
		},
		{
			name: "HTTPRequesteBody/no body",
			trace: func(t *testing.T) *ClientTrace {
				return &ClientTrace{
					HTTPRequestBody: func(r *http.Request) {
						if r.Method != "GET" {
							t.Errorf("Unexpected method: %s", r.Method)
						}
						r.Method = "unf"
						if r.Body != nil {
							t.Errorf("non-nil body")
						}
					},
				}
			},
			req: func() *http.Request {
				req, _ := http.NewRequest("GET", "/", nil)
				return req
			}(),
			finalReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "/", nil)
				req.Header.Add("Host", "example.com")
				return req
			}(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			trace := test.trace(t)
			trace.httpRequestBody(test.req)
			trace.httpRequest(test.req)
			if d := testy.DiffHTTPRequest(test.finalReq, test.req); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestReplayReadCloser(t *testing.T) {
	tests := []struct {
		name     string
		input    io.ReadCloser
		expected string
		readErr  string
		closeErr string
	}{
		{
			name:     "no errors",
			input:    io.NopCloser(strings.NewReader("testing")),
			expected: "testing",
		},
		{
			name:     "read error",
			input:    io.NopCloser(&errReader{Reader: strings.NewReader("testi"), err: errors.New("read error 1")}),
			expected: "testi",
			readErr:  "read error 1",
		},
		{
			name:     "close error",
			input:    &errCloser{Reader: strings.NewReader("testin"), err: errors.New("close error 1")},
			expected: "testin",
			closeErr: "close error 1",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			content, err := io.ReadAll(test.input.(io.Reader))
			closeErr := test.input.Close()
			rc := newReplay(content, err, closeErr)

			result, resultErr := io.ReadAll(rc.(io.Reader))
			resultCloseErr := rc.Close()
			if d := testy.DiffText(test.expected, result); d != nil {
				t.Error(d)
			}
			testy.Error(t, test.readErr, resultErr)
			testy.Error(t, test.closeErr, resultCloseErr)
		})
	}
}
