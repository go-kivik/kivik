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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"gitlab.com/flimzy/testy"
	"golang.org/x/net/publicsuffix"

	kivik "github.com/go-kivik/kivik/v4"
)

var defaultUA = func() string {
	c := &Client{}
	return c.userAgent()
}()

func TestNew(t *testing.T) {
	type tt struct {
		dsn      string
		expected *Client
		status   int
		err      string
	}

	tests := testy.NewTable()
	tests.Add("invalid url", tt{
		dsn:    "http://foo.com/%xx",
		status: http.StatusBadRequest,
		err:    `parse "?http://foo.com/%xx"?: invalid URL escape "%xx"`,
	})
	tests.Add("no url", tt{
		dsn:    "",
		status: http.StatusBadRequest,
		err:    "no URL specified",
	})
	tests.Add("no auth", tt{
		dsn: "http://foo.com/",
		expected: &Client{
			Client: &http.Client{},
			rawDSN: "http://foo.com/",
			dsn: &url.URL{
				Scheme: "http",
				Host:   "foo.com",
				Path:   "/",
			},
		},
	})
	tests.Add("auth success", func(t *testing.T) interface{} {
		h := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"userCtx":{"name":"user"}}`) // nolint: errcheck
		}
		s := httptest.NewServer(http.HandlerFunc(h))
		authDSN, _ := url.Parse(s.URL)
		dsn, _ := url.Parse(s.URL + "/")
		authDSN.User = url.UserPassword("user", "password")
		jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		c := &Client{
			Client: &http.Client{Jar: jar},
			rawDSN: authDSN.String(),
			dsn:    dsn,
		}
		auth := &CookieAuth{
			Username:  "user",
			Password:  "password",
			client:    c,
			transport: http.DefaultTransport,
		}
		c.auth = auth
		c.Client.Transport = auth

		return tt{
			dsn:      authDSN.String(),
			expected: c,
		}
	})
	tests.Add("default url scheme", tt{
		dsn: "foo.com",
		expected: &Client{
			Client: &http.Client{},
			rawDSN: "foo.com",
			dsn: &url.URL{
				Scheme: "http",
				Host:   "foo.com",
				Path:   "/",
			},
		},
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		result, err := New(&http.Client{}, tt.dsn, nil)
		statusErrorRE(t, tt.err, tt.status, err)
		result.UserAgents = nil // Determinism
		if d := testy.DiffInterface(tt.expected, result); d != nil {
			t.Error(d)
		}
	})
}

func TestParseDSN(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *url.URL
		status   int
		err      string
	}{
		{
			name:  "happy path",
			input: "http://foo.com/",
			expected: &url.URL{
				Scheme: "http",
				Host:   "foo.com",
				Path:   "/",
			},
		},
		{
			name:  "default scheme",
			input: "foo.com",
			expected: &url.URL{
				Scheme: "http",
				Host:   "foo.com",
				Path:   "/",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := parseDSN(test.input)
			statusErrorRE(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Fatal(d)
			}
		})
	}
}

func TestDSN(t *testing.T) {
	expected := "foo"
	client := &Client{rawDSN: expected}
	result := client.DSN()
	if result != expected {
		t.Errorf("Unexpected result: %s", result)
	}
}

func TestFixPath(t *testing.T) {
	tests := []struct {
		Input    string
		Expected string
	}{
		{Input: "foo", Expected: "/foo"},
		{Input: "foo?oink=yes", Expected: "/foo"},
		{Input: "foo/bar", Expected: "/foo/bar"},
		{Input: "foo%2Fbar", Expected: "/foo%2Fbar"},
	}
	for _, test := range tests {
		req, _ := http.NewRequest("GET", "http://localhost/"+test.Input, nil)
		fixPath(req, test.Input)
		if req.URL.EscapedPath() != test.Expected {
			t.Errorf("Path for '%s' not fixed.\n\tExpected: %s\n\t  Actual: %s\n", test.Input, test.Expected, req.URL.EscapedPath())
		}
	}
}

func TestEncodeBody(t *testing.T) {
	type encodeTest struct {
		name  string
		input interface{}

		expected string
		status   int
		err      string
	}
	tests := []encodeTest{
		{
			name:     "Null",
			input:    nil,
			expected: "null",
		},
		{
			name: "Struct",
			input: struct {
				Foo string `json:"foo"`
			}{Foo: "bar"},
			expected: `{"foo":"bar"}`,
		},
		{
			name:   "JSONError",
			input:  func() {}, // Functions cannot be marshaled to JSON
			status: http.StatusBadRequest,
			err:    "json: unsupported type: func()",
		},
		{
			name:     "raw json input",
			input:    json.RawMessage(`{"foo":"bar"}`),
			expected: `{"foo":"bar"}`,
		},
		{
			name:     "byte slice input",
			input:    []byte(`{"foo":"bar"}`),
			expected: `{"foo":"bar"}`,
		},
		{
			name:     "string input",
			input:    `{"foo":"bar"}`,
			expected: `{"foo":"bar"}`,
		},
	}
	for _, test := range tests {
		func(test encodeTest) {
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()
				r := EncodeBody(test.input)
				defer r.Close() // nolint: errcheck
				body, err := io.ReadAll(r)
				testy.StatusError(t, test.err, test.status, err)
				result := strings.TrimSpace(string(body))
				if result != test.expected {
					t.Errorf("Result\nExpected: %s\n  Actual: %s\n", test.expected, result)
				}
			})
		}(test)
	}
}

func TestSetHeaders(t *testing.T) {
	type shTest struct {
		Name     string
		Options  *Options
		Expected http.Header
	}
	tests := []shTest{
		{
			Name: "NoOpts",
			Expected: http.Header{
				"Accept":       {"application/json"},
				"Content-Type": {"application/json"},
			},
		},
		{
			Name:    "Content-Type",
			Options: &Options{ContentType: "image/gif"},
			Expected: http.Header{
				"Accept":       {"application/json"},
				"Content-Type": {"image/gif"},
			},
		},
		{
			Name:    "Accept",
			Options: &Options{Accept: "image/gif"},
			Expected: http.Header{
				"Accept":       {"image/gif"},
				"Content-Type": {"application/json"},
			},
		},
		{
			Name:    "FullCommit",
			Options: &Options{FullCommit: true},
			Expected: http.Header{
				"Accept":              {"application/json"},
				"Content-Type":        {"application/json"},
				"X-Couch-Full-Commit": {"true"},
			},
		},
		{
			Name: "Destination",
			Options: &Options{Header: http.Header{
				HeaderDestination: []string{"somewhere nice"},
			}},
			Expected: http.Header{
				"Accept":       {"application/json"},
				"Content-Type": {"application/json"},
				"Destination":  {"somewhere nice"},
			},
		},
		{
			Name:    "If-None-Match",
			Options: &Options{IfNoneMatch: `"foo"`},
			Expected: http.Header{
				"Accept":        {"application/json"},
				"Content-Type":  {"application/json"},
				"If-None-Match": {`"foo"`},
			},
		},
		{
			Name:    "Unquoted If-None-Match",
			Options: &Options{IfNoneMatch: `foo`},
			Expected: http.Header{
				"Accept":        {"application/json"},
				"Content-Type":  {"application/json"},
				"If-None-Match": {`"foo"`},
			},
		},
	}
	for _, test := range tests {
		func(test shTest) {
			t.Run(test.Name, func(t *testing.T) {
				t.Parallel()
				req, err := http.NewRequest("GET", "/", nil)
				if err != nil {
					panic(err)
				}
				setHeaders(req, test.Options)
				if d := testy.DiffInterface(test.Expected, req.Header); d != nil {
					t.Errorf("Headers:\n%s\n", d)
				}
			})
		}(test)
	}
}

func TestSetQuery(t *testing.T) {
	tests := []struct {
		name     string
		req      *http.Request
		opts     *Options
		expected *http.Request
	}{
		{
			name:     "nil query",
			req:      &http.Request{URL: &url.URL{}},
			expected: &http.Request{URL: &url.URL{}},
		},
		{
			name:     "empty query",
			req:      &http.Request{URL: &url.URL{RawQuery: "a=b"}},
			opts:     &Options{Query: url.Values{}},
			expected: &http.Request{URL: &url.URL{RawQuery: "a=b"}},
		},
		{
			name:     "options query",
			req:      &http.Request{URL: &url.URL{}},
			opts:     &Options{Query: url.Values{"foo": []string{"a"}}},
			expected: &http.Request{URL: &url.URL{RawQuery: "foo=a"}},
		},
		{
			name:     "merged queries",
			req:      &http.Request{URL: &url.URL{RawQuery: "bar=b"}},
			opts:     &Options{Query: url.Values{"foo": []string{"a"}}},
			expected: &http.Request{URL: &url.URL{RawQuery: "bar=b&foo=a"}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setQuery(test.req, test.opts)
			if d := testy.DiffInterface(test.expected, test.req); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestETag(t *testing.T) {
	tests := []struct {
		name     string
		input    *http.Response
		expected string
		found    bool
	}{
		{
			name:     "nil response",
			input:    nil,
			expected: "",
			found:    false,
		},
		{
			name:     "No etag",
			input:    &http.Response{},
			expected: "",
			found:    false,
		},
		{
			name: "ETag",
			input: &http.Response{
				Header: http.Header{
					"ETag": {`"foo"`},
				},
			},
			expected: "foo",
			found:    true,
		},
		{
			name: "Etag",
			input: &http.Response{
				Header: http.Header{
					"Etag": {`"bar"`},
				},
			},
			expected: "bar",
			found:    true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, found := ETag(test.input)
			if result != test.expected {
				t.Errorf("Unexpected result: %s", result)
			}
			if found != test.found {
				t.Errorf("Unexpected found: %v", found)
			}
		})
	}
}

func TestGetRev(t *testing.T) {
	tests := []struct {
		name          string
		resp          *http.Response
		expected, err string
	}{
		{
			name: "error response",
			resp: &http.Response{
				StatusCode: 400,
				Request:    &http.Request{Method: "POST"},
				Body:       io.NopCloser(strings.NewReader("")),
			},
			err: "Bad Request",
		},
		{
			name: "no ETag header",
			resp: &http.Response{
				StatusCode: 200,
				Request:    &http.Request{Method: "POST"},
				Body:       io.NopCloser(strings.NewReader("")),
			},
			err: "unable to determine document revision: EOF",
		},
		{
			name: "normalized Etag header",
			resp: &http.Response{
				StatusCode: 200,
				Request:    &http.Request{Method: "POST"},
				Header:     http.Header{"Etag": {`"12345"`}},
				Body:       io.NopCloser(strings.NewReader("")),
			},
			expected: `12345`,
		},
		{
			name: "satndard ETag header",
			resp: &http.Response{
				StatusCode: 200,
				Request:    &http.Request{Method: "POST"},
				Header:     http.Header{"ETag": {`"12345"`}},
				Body:       Body(""),
			},
			expected: `12345`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetRev(test.resp)
			testy.Error(t, test.err, err)
			if result != test.expected {
				t.Errorf("Got %s, expected %s", result, test.expected)
			}
		})
	}
}

func TestDoJSON(t *testing.T) {
	tests := []struct {
		name         string
		method, path string
		opts         *Options
		client       *Client
		expected     interface{}
		status       int
		err          string
	}{
		{
			name:   "network error",
			method: "GET",
			client: newTestClient(nil, errors.New("net error")),
			status: http.StatusBadGateway,
			err:    `Get "?http://example.com"?: net error`,
		},
		{
			name:   "error response",
			method: "GET",
			client: newTestClient(&http.Response{
				StatusCode: 401,
				Header: http.Header{
					"Content-Type":   {"application/json"},
					"Content-Length": {"67"},
				},
				ContentLength: 67,
				Body:          Body(`{"error":"unauthorized","reason":"Name or password is incorrect."}`),
				Request:       &http.Request{Method: "GET"},
			}, nil),
			status: http.StatusUnauthorized,
			err:    "Unauthorized: Name or password is incorrect.",
		},
		{
			name:   "invalid JSON in response",
			method: "GET",
			client: newTestClient(&http.Response{
				StatusCode: 200,
				Header: http.Header{
					"Content-Type":   {"application/json"},
					"Content-Length": {"67"},
				},
				ContentLength: 67,
				Body:          Body(`invalid response`),
				Request:       &http.Request{Method: "GET"},
			}, nil),
			status: http.StatusBadGateway,
			err:    "invalid character 'i' looking for beginning of value",
		},
		{
			name:   "success",
			method: "GET",
			client: newTestClient(&http.Response{
				StatusCode: 200,
				Header: http.Header{
					"Content-Type":   {"application/json"},
					"Content-Length": {"15"},
				},
				ContentLength: 15,
				Body:          Body(`{"foo":"bar"}`),
				Request:       &http.Request{Method: "GET"},
			}, nil),
			expected: map[string]interface{}{"foo": "bar"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var i interface{}
			err := test.client.DoJSON(context.Background(), test.method, test.path, test.opts, &i)
			testy.StatusErrorRE(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, i); d != nil {
				t.Errorf("JSON result differs:\n%s\n", d)
			}
		})
	}
}

func TestNewRequest(t *testing.T) {
	tests := []struct {
		name         string
		method, path string
		expected     *http.Request
		client       *Client
		status       int
		err          string
	}{
		{
			name:   "invalid URL",
			client: newTestClient(nil, nil),
			method: "GET",
			path:   "%xx",
			status: http.StatusBadRequest,
			err:    `parse "?%xx"?: invalid URL escape "%xx"`,
		},
		{
			name:   "invalid method",
			method: "FOO BAR",
			client: newTestClient(nil, nil),
			status: http.StatusBadRequest,
			err:    `net/http: invalid method "FOO BAR"`,
		},
		{
			name:   "success",
			method: "GET",
			path:   "foo",
			client: newTestClient(nil, nil),
			expected: &http.Request{
				Method: "GET",
				URL: func() *url.URL {
					url := newTestClient(nil, nil).dsn
					url.Path = "/foo"
					return url
				}(),
				Proto:      "HTTP/1.1",
				ProtoMajor: 1,
				ProtoMinor: 1,
				Header: http.Header{
					"User-Agent": []string{defaultUA},
				},
				Host: "example.com",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := test.client.NewRequest(context.Background(), test.method, test.path, nil, nil)
			statusErrorRE(t, test.err, test.status, err)
			test.expected = test.expected.WithContext(req.Context()) // determinism
			if d := testy.DiffInterface(test.expected, req); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestDoReq(t *testing.T) {
	type tt struct {
		trace        func(t *testing.T, success *bool) *ClientTrace
		method, path string
		opts         *Options
		client       *Client
		status       int
		err          string
	}

	tests := testy.NewTable()
	tests.Add("no method", tt{
		status: 500,
		err:    "chttp: method required",
	})
	tests.Add("invalid url", tt{
		method: "GET",
		path:   "%xx",
		client: newTestClient(nil, nil),
		status: http.StatusBadRequest,
		err:    `parse "?%xx"?: invalid URL escape "%xx"`,
	})
	tests.Add("network error", tt{
		method: "GET",
		path:   "foo",
		client: newTestClient(nil, errors.New("net error")),
		status: http.StatusBadGateway,
		err:    `Get "?http://example.com/foo"?: net error`,
	})
	tests.Add("error response", tt{
		method: "GET",
		path:   "foo",
		client: newTestClient(&http.Response{
			StatusCode: 400,
			Body:       Body(""),
		}, nil),
		// No error here
	})
	tests.Add("success", tt{
		method: "GET",
		path:   "foo",
		client: newTestClient(&http.Response{
			StatusCode: 200,
			Body:       Body(""),
		}, nil),
		// success!
	})
	tests.Add("body error", tt{
		method: "PUT",
		path:   "foo",
		client: newTestClient(nil, &kivik.Error{Status: http.StatusBadRequest, Message: "bad request"}),
		status: http.StatusBadRequest,
		err:    `Put "?http://example.com/foo"?: bad request`,
	})
	tests.Add("response trace", tt{
		trace: func(t *testing.T, success *bool) *ClientTrace {
			return &ClientTrace{
				HTTPResponse: func(r *http.Response) {
					*success = true
					expected := &http.Response{StatusCode: 200}
					if d := testy.DiffHTTPResponse(expected, r); d != nil {
						t.Error(d)
					}
				},
			}
		},
		method: "GET",
		path:   "foo",
		client: newTestClient(&http.Response{
			StatusCode: 200,
			Body:       Body(""),
		}, nil),
		// response body trace
	})
	tests.Add("response body trace", tt{
		trace: func(t *testing.T, success *bool) *ClientTrace {
			return &ClientTrace{
				HTTPResponseBody: func(r *http.Response) {
					*success = true
					expected := &http.Response{
						StatusCode: 200,
						Body:       Body("foo"),
					}
					if d := testy.DiffHTTPResponse(expected, r); d != nil {
						t.Error(d)
					}
				},
			}
		},
		method: "PUT",
		path:   "foo",
		client: newTestClient(&http.Response{
			StatusCode: 200,
			Body:       Body("foo"),
		}, nil),
		// response trace
	})
	tests.Add("request trace", tt{
		trace: func(t *testing.T, success *bool) *ClientTrace {
			return &ClientTrace{
				HTTPRequest: func(r *http.Request) {
					*success = true
					expected := httptest.NewRequest("PUT", "/foo", nil)
					expected.Header.Add("Accept", "application/json")
					expected.Header.Add("Content-Type", "application/json")
					expected.Header.Add("Content-Encoding", "gzip")
					expected.Header.Add("User-Agent", defaultUA)
					if d := testy.DiffHTTPRequest(expected, r); d != nil {
						t.Error(d)
					}
				},
			}
		},
		method: "PUT",
		path:   "/foo",
		client: newTestClient(&http.Response{
			StatusCode: 200,
			Body:       Body("foo"),
		}, nil),
		opts: &Options{
			Body: Body("bar"),
		},
		// request trace
	})
	tests.Add("request body trace", tt{
		trace: func(t *testing.T, success *bool) *ClientTrace {
			return &ClientTrace{
				HTTPRequestBody: func(r *http.Request) {
					*success = true
					body := io.NopCloser(bytes.NewReader([]byte{
						31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 74, 74, 44, 2,
						4, 0, 0, 255, 255, 170, 140, 255, 118, 3, 0, 0, 0,
					}))
					expected := httptest.NewRequest("PUT", "/foo", body)
					expected.Header.Add("Accept", "application/json")
					expected.Header.Add("Content-Type", "application/json")
					expected.Header.Add("Content-Encoding", "gzip")
					expected.Header.Add("User-Agent", defaultUA)
					expected.Header.Add("Content-Length", "27")
					if d := testy.DiffHTTPRequest(expected, r); d != nil {
						t.Error(d)
					}
				},
			}
		},
		method: "PUT",
		path:   "/foo",
		client: newTestClient(&http.Response{
			StatusCode: 200,
			Body:       Body("foo"),
		}, nil),
		opts: &Options{
			Body: Body("bar"),
		},
		// request body trace
	})
	tests.Add("couchdb mounted below root", tt{
		client: newCustomClient("http://foo.com/dbroot/", func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/dbroot/foo" {
				return nil, errors.Errorf("Unexpected path: %s", r.URL.Path)
			}
			return &http.Response{}, nil
		}),
		method: "GET",
		path:   "/foo",
	})
	tests.Add("user agent", tt{
		client: newCustomClient("http://foo.com/", func(r *http.Request) (*http.Response, error) {
			if ua := r.UserAgent(); ua != defaultUA {
				return nil, errors.Errorf("Unexpected User Agent: %s", ua)
			}
			return &http.Response{}, nil
		}),
		method: "GET",
		path:   "/foo",
	})
	tests.Add("gzipped request", tt{
		client: newCustomClient("http://foo.com/", func(r *http.Request) (*http.Response, error) {
			if ce := r.Header.Get("Content-Encoding"); ce != "gzip" {
				return nil, errors.Errorf("Unexpected Content-Encoding: %s", ce)
			}
			return &http.Response{}, nil
		}),
		method: "PUT",
		path:   "/foo",
		opts: &Options{
			Body: Body("raw body"),
		},
	})
	tests.Add("gzipped disabled", tt{
		client: newCustomClient("http://foo.com/", func(r *http.Request) (*http.Response, error) {
			if ce := r.Header.Get("Content-Encoding"); ce != "" {
				return nil, errors.Errorf("Unexpected Content-Encoding: %s", ce)
			}
			return &http.Response{}, nil
		}),
		method: "PUT",
		path:   "/foo",
		opts: &Options{
			Body:   Body("raw body"),
			NoGzip: true,
		},
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		ctx := context.Background()
		traceSuccess := true
		if tt.trace != nil {
			traceSuccess = false
			ctx = WithClientTrace(ctx, tt.trace(t, &traceSuccess))
		}
		res, err := tt.client.DoReq(ctx, tt.method, tt.path, tt.opts)
		statusErrorRE(t, tt.err, tt.status, err)
		defer res.Body.Close()
		_, _ = io.Copy(io.Discard, res.Body)
		if !traceSuccess {
			t.Error("Trace failed")
		}
	})
}

func TestDoError(t *testing.T) {
	tests := []struct {
		name         string
		method, path string
		opts         *Options
		client       *Client
		status       int
		err          string
	}{
		{
			name:   "no method",
			status: 500,
			err:    "chttp: method required",
		},
		{
			name:   "error response",
			method: "GET",
			path:   "foo",
			client: newTestClient(&http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       Body(""),
				Request:    &http.Request{Method: "GET"},
			}, nil),
			status: http.StatusBadRequest,
			err:    "Bad Request",
		},
		{
			name:   "success",
			method: "GET",
			path:   "foo",
			client: newTestClient(&http.Response{
				StatusCode: http.StatusOK,
				Body:       Body(""),
				Request:    &http.Request{Method: "GET"},
			}, nil),
			// No error
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.client.DoError(context.Background(), test.method, test.path, test.opts)
			testy.StatusError(t, test.err, test.status, err)
		})
	}
}

func TestNetError(t *testing.T) {
	tests := []struct {
		name  string
		input error

		status int
		err    string
	}{
		{
			name:   "nil",
			input:  nil,
			status: 0,
			err:    "",
		},
		{
			name: "timeout",
			input: func() error {
				s := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
					time.Sleep(1 * time.Second)
				}))
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
				defer cancel()
				req, err := http.NewRequest("GET", s.URL, nil)
				if err != nil {
					t.Fatal(err)
				}
				_, err = http.DefaultClient.Do(req.WithContext(ctx))
				return err
			}(),
			status: http.StatusBadGateway,
			err:    `Get "?http://127.0.0.1:\d+"?: context deadline exceeded`,
		},
		{
			name: "cannot resolve host",
			input: func() error {
				req, err := http.NewRequest("GET", "http://foo.com.invalid.hostname", nil)
				if err != nil {
					t.Fatal(err)
				}
				_, err = http.DefaultClient.Do(req)
				return err
			}(),
			status: http.StatusBadGateway,
			err:    ": no such host$",
		},
		{
			name: "connection refused",
			input: func() error {
				req, err := http.NewRequest("GET", "http://localhost:99", nil)
				if err != nil {
					t.Fatal(err)
				}
				_, err = http.DefaultClient.Do(req)
				return err
			}(),
			status: http.StatusBadGateway,
			err:    ": connection refused$",
		},
		{
			name: "too many redirects",
			input: func() error {
				var s *httptest.Server
				redirHandler := func(w http.ResponseWriter, r *http.Request) {
					http.Redirect(w, r, s.URL, 302)
				}
				s = httptest.NewServer(http.HandlerFunc(redirHandler))
				_, err := http.Get(s.URL)
				return err
			}(),
			status: http.StatusBadGateway,
			err:    `^Get "?http://127.0.0.1:\d+"?: stopped after 10 redirects$`,
		},
		{
			name: "url error",
			input: &url.Error{
				Op:  "Get",
				URL: "http://foo.com/",
				Err: errors.New("some error"),
			},
			status: http.StatusBadGateway,
			err:    `Get "?http://foo.com/"?: some error`,
		},
		{
			name: "url error with embedded status",
			input: &url.Error{
				Op:  "Get",
				URL: "http://foo.com/",
				Err: &kivik.Error{Status: http.StatusBadRequest, Message: "some error"},
			},
			status: http.StatusBadRequest,
			err:    `Get "?http://foo.com/"?: some error`,
		},
		{
			name:   "other error",
			input:  errors.New("other error"),
			status: http.StatusBadGateway,
			err:    "other error",
		},
		{
			name:   "other error with embedded status",
			input:  &kivik.Error{Status: http.StatusBadRequest, Message: "bad req"},
			status: http.StatusBadRequest,
			err:    "bad req",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := netError(test.input)
			statusErrorRE(t, test.err, test.status, err)
		})
	}
}

func TestUserAgent(t *testing.T) {
	tests := []struct {
		name     string
		ua       []string
		expected string
	}{
		{
			name: "defaults",
			expected: fmt.Sprintf("%s/%s (Language=%s; Platform=%s/%s)",
				UserAgent, Version, runtime.Version(), runtime.GOARCH, runtime.GOOS),
		},
		{
			name: "custom",
			ua:   []string{"Oinky/1.2.3"},
			expected: fmt.Sprintf("%s/%s (Language=%s; Platform=%s/%s) Oinky/1.2.3",
				UserAgent, Version, runtime.Version(), runtime.GOARCH, runtime.GOOS),
		},
		{
			name: "multiple",
			ua:   []string{"Oinky/1.2.3", "Moo/5.4.3"},
			expected: fmt.Sprintf("%s/%s (Language=%s; Platform=%s/%s) Oinky/1.2.3 Moo/5.4.3",
				UserAgent, Version, runtime.Version(), runtime.GOARCH, runtime.GOOS),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &Client{
				UserAgents: test.ua,
			}
			result := c.userAgent()
			if result != test.expected {
				t.Errorf("Unexpected user agent: %s", result)
			}
		})
	}
}

func Test_extractRev(t *testing.T) {
	type tt struct {
		resp *http.Response
		rev  string
		err  string
	}

	tests := testy.NewTable()
	tests.Add("HEAD request", tt{
		resp: &http.Response{
			Request: &http.Request{
				Method: http.MethodHead,
			},
		},
		rev: "",
		err: "unable to determine document revision",
	})
	tests.Add("empty body", tt{
		resp: &http.Response{
			Body: io.NopCloser(strings.NewReader("")),
		},
		rev: "",
		err: "unable to determine document revision: EOF",
	})
	tests.Add("invalid JSON", tt{
		resp: &http.Response{
			Body: io.NopCloser(strings.NewReader(`bogus`)),
		},
		err: `unable to determine document revision: invalid character 'b' looking for beginning of value`,
	})
	tests.Add("rev found", tt{
		resp: &http.Response{
			Body: io.NopCloser(strings.NewReader(`{"_rev":"1-xyz"}`)),
		},
		rev: "1-xyz",
	})
	tests.Add("rev found in middle", tt{
		resp: &http.Response{
			Body: io.NopCloser(strings.NewReader(`{
				"_id":"foo",
				"_rev":"1-xyz",
				"asdf":"qwerty",
				"number":12345
			}`)),
		},
		rev: "1-xyz",
	})
	tests.Add("rev not found middle", tt{
		resp: &http.Response{
			Body: io.NopCloser(strings.NewReader(`{
				"_id":"foo",
				"asdf":"qwerty",
				"number":12345
			}`)),
		},
		err: "unable to determine document revision: _rev key not found in response body",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		if tt.resp.Request == nil {
			tt.resp.Request = &http.Request{}
		}
		rev, err := extractRev(tt.resp)
		testy.Error(t, tt.err, err)
		if tt.rev != rev {
			t.Errorf("Expected %s, got %s", tt.rev, rev)
		}
		if d := testy.DiffJSON(testy.Snapshot(t), tt.resp.Body); d != nil {
			t.Error(d)
		}
	})
}

func Test_readRev(t *testing.T) {
	type tt struct {
		input string
		rev   string
		err   string
	}

	tests := testy.NewTable()
	tests.Add("empty body", tt{
		input: "",
		err:   "EOF",
	})
	tests.Add("invalid JSON", tt{
		input: "bogus",
		err:   `invalid character 'b' looking for beginning of value`,
	})
	tests.Add("non-object", tt{
		input: "[]",
		err:   `Expected '{' token, found "["`,
	})
	tests.Add("_rev missing", tt{
		input: "{}",
		err:   "_rev key not found in response body",
	})
	tests.Add("invalid key", tt{
		input: "{asdf",
		err:   `invalid character 'a'`,
	})
	tests.Add("invalid value", tt{
		input: `{"_rev":xyz}`,
		err:   `invalid character 'x' looking for beginning of value`,
	})
	tests.Add("non-string rev", tt{
		input: `{"_rev":[]}`,
		err:   `found "[" in place of _rev value`,
	})
	tests.Add("success", tt{
		input: `{"_rev":"1-xyz"}`,
		rev:   "1-xyz",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		rev, err := readRev(strings.NewReader(tt.input))
		testy.Error(t, tt.err, err)
		if rev != tt.rev {
			t.Errorf("Wanted %s, got %s", tt.rev, rev)
		}
	})
}
