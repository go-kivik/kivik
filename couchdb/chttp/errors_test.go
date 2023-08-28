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
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	kivik "github.com/go-kivik/kivik/v4"
)

func TestHTTPErrorError(t *testing.T) {
	tests := []struct {
		name     string
		input    *HTTPError
		expected string
	}{
		{
			name: "No reason",
			input: &HTTPError{
				Response: &http.Response{StatusCode: 400},
			},
			expected: "Bad Request",
		},
		{
			name: "Reason, HTTP code",
			input: &HTTPError{
				Response: &http.Response{StatusCode: 400},
				Reason:   "Bad stuff",
			},
			expected: "Bad Request: Bad stuff",
		},
		{
			name: "Non-HTTP code",
			input: &HTTPError{
				Response: &http.Response{StatusCode: 604},
				Reason:   "Bad stuff",
			},
			expected: "Bad stuff",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.input.Error()
			if result != test.expected {
				t.Errorf("Unexpected result: %s", result)
			}
		})
	}
}

func TestResponseError(t *testing.T) {
	tests := []struct {
		name     string
		resp     *http.Response
		status   int
		err      string
		expected interface{}
	}{
		{
			name:     "non error",
			resp:     &http.Response{StatusCode: 200},
			expected: nil,
		},
		{
			name: "HEAD error",
			resp: &http.Response{
				StatusCode: http.StatusNotFound,
				Request:    &http.Request{Method: "HEAD"},
				Body:       Body(""),
			},
			status: http.StatusNotFound,
			err:    "Not Found",
			expected: &kivik.Error{
				Status: http.StatusNotFound,
				Err: &HTTPError{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
					},
				},
			},
		},
		{
			name: "2.0.0 error",
			resp: &http.Response{
				StatusCode: http.StatusBadRequest,
				Header: http.Header{
					"Cache-Control":       {"must-revalidate"},
					"Content-Length":      {"194"},
					"Content-Type":        {"application/json"},
					"Date":                {"Fri, 27 Oct 2017 15:34:07 GMT"},
					"Server":              {"CouchDB/2.0.0 (Erlang OTP/17)"},
					"X-Couch-Request-ID":  {"92d05bd015"},
					"X-CouchDB-Body-Time": {"0"},
				},
				ContentLength: 194,
				Body:          Body(`{"error":"illegal_database_name","reason":"Name: '_foo'. Only lowercase characters (a-z), digits (0-9), and any of the characters _, $, (, ), +, -, and / are allowed. Must begin with a letter."}`),
				Request:       &http.Request{Method: "PUT"},
			},
			status: http.StatusBadRequest,
			err:    "Bad Request: Name: '_foo'. Only lowercase characters (a-z), digits (0-9), and any of the characters _, $, (, ), +, -, and / are allowed. Must begin with a letter.",
			expected: &kivik.Error{
				Status: http.StatusBadRequest,
				Err: &HTTPError{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
					},
					Reason: "Name: '_foo'. Only lowercase characters (a-z), digits (0-9), and any of the characters _, $, (, ), +, -, and / are allowed. Must begin with a letter.",
				},
			},
		},
		{
			name: "invalid json error",
			resp: &http.Response{
				StatusCode: http.StatusBadRequest,
				Header: http.Header{
					"Server":         {"CouchDB/1.6.1 (Erlang OTP/17)"},
					"Date":           {"Fri, 27 Oct 2017 15:42:34 GMT"},
					"Content-Type":   {"application/json"},
					"Content-Length": {"194"},
					"Cache-Control":  {"must-revalidate"},
				},
				ContentLength: 194,
				Body:          Body("invalid json"),
				Request:       &http.Request{Method: "PUT"},
			},
			status: http.StatusBadRequest,
			err:    "Bad Request",
			expected: &kivik.Error{
				Status: http.StatusBadRequest,
				Err: &HTTPError{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ResponseError(test.resp)
			testy.StatusError(t, test.err, test.status, err)
			if he, ok := err.(*HTTPError); ok {
				he.Response = nil
				err = he
			}
			if d := testy.DiffInterface(test.expected, err); d != nil {
				t.Error(d)
			}
		})
	}
}
