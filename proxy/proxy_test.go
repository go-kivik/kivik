package proxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/flimzy/diff"
)

func mustNew(t *testing.T, url string) *Proxy {
	p, err := New(url)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestNew(t *testing.T) {
	type newTest struct {
		Name  string
		URL   string
		Error string
	}
	tests := []newTest{
		{
			Name:  "NoURL",
			Error: "no URL specified",
		},
		{
			Name: "ValidURL",
			URL:  "http://foo.com/",
		},
		{
			Name:  "InvalidURL",
			URL:   "http://foo.com:port with spaces/",
			Error: `parse http://foo.com:port with spaces/: invalid character " " in host name`,
		},
		{
			Name:  "Auth",
			URL:   "http://foo:bar@foo.com/",
			Error: "proxy URL must not contain auth credentials",
		},
		{
			Name:  "Query",
			URL:   "http://foo.com?yes=no",
			Error: "proxy URL must not contain query parameters",
		},
	}
	for _, test := range tests {
		func(test newTest) {
			t.Run(test.Name, func(t *testing.T) {
				t.Parallel()
				_, err := New(test.URL)
				var msg string
				if err != nil {
					msg = err.Error()
				}
				if msg != test.Error {
					t.Errorf("Expected error: %s\n  Actual error: %s", test.Error, msg)
				}
			})
		}(test)
	}
}

func TestURL(t *testing.T) {
	type urlTest struct {
		Name       string
		ServerURL  string
		RequestURL string
		Expected   string
	}
	tests := []urlTest{
		{
			Name:       "Root",
			ServerURL:  "https://foobar.com:9000/foo",
			RequestURL: "http://foo.com:1234/",
			Expected:   "https://foobar.com:9000/foo/",
		},
		{
			Name:       "FullURL",
			ServerURL:  "https://foobar.com:9000/foo/",
			RequestURL: "http://foo.com:1234/_replicator",
			Expected:   "https://foobar.com:9000/foo/_replicator",
		},
		{
			Name:       "BasicAuth",
			ServerURL:  "https://foobar.com:9000/foo/",
			RequestURL: "http://bob:abc123@foo.com:1234/_replicator",
			Expected:   "https://bob:abc123@foobar.com:9000/foo/_replicator",
		},
		{
			Name:       "Query",
			ServerURL:  "https://foobar.com:9000/foo/",
			RequestURL: "http://bob:abc123@foo.com:1234/_replicator?yesterday=today",
			Expected:   "https://bob:abc123@foobar.com:9000/foo/_replicator?yesterday=today",
		},
	}
	for _, test := range tests {
		func(test urlTest) {
			t.Run(test.Name, func(t *testing.T) {
				t.Parallel()
				p, err := New(test.ServerURL)
				if err != nil {
					t.Fatalf("Proxy instantiation failed: %s", err)
				}
				reqURL, err := url.Parse(test.RequestURL)
				if err != nil {
					t.Fatalf("Failed to parse URL: %s", err)
				}
				result := p.url(reqURL)
				if result != test.Expected {
					t.Errorf("Expected: %s\n  Actual: %s", test.Expected, result)
				}
			})
		}(test)
	}
}

func serverDSN(t *testing.T) string {
	for _, env := range []string{"KIVIK_TEST_DSN_COUCH16", "KIVIK_TEST_DSN_COUCH20"} {
		if dsn := os.Getenv(env); dsn != "" {
			parsed, err := url.Parse(dsn)
			if err != nil {
				panic(err)
			}
			parsed.User = nil
			return parsed.String()
		}
	}
	t.Skip("No server specified in environment")
	return ""
}

func TestClient(t *testing.T) {
	type clientTest struct {
		Name     string
		Client   *http.Client
		Expected *http.Client
	}
	c := &http.Client{}
	tests := []clientTest{
		{
			Name:     "None",
			Expected: DefaultClient,
		},
		{
			Name:     "Explicit",
			Client:   c,
			Expected: c,
		},
	}
	for _, test := range tests {
		func(test clientTest) {
			t.Run(test.Name, func(t *testing.T) {
				p, err := New("http://foo.com")
				if err != nil {
					t.Fatal(err)
				}
				p.HTTPClient = test.Client
				result := p.client()
				if result != test.Expected {
					t.Error("Returned unexpected client")
				}
			})
		}(test)
	}
}

func TestProxy(t *testing.T) {
	type getTest struct {
		Name        string
		Method      string
		URL         string
		Body        io.Reader
		Status      int
		BodyMatch   string
		HeaderMatch map[string]string
	}
	tests := []getTest{
		{
			Name:   "InvalidMethod",
			Method: "CHICKEN",
			URL:    "http://foo.com/_utils",
			Status: http.StatusMethodNotAllowed,
		},
		{
			Name:        "UtilsRedirect",
			Method:      http.MethodGet,
			URL:         "http://foo.com/_utils",
			Status:      http.StatusMovedPermanently,
			HeaderMatch: map[string]string{"Location": "http://foo.com/_utils/"},
		},
		{
			Name:        "UtilsRedirectQuery",
			Method:      http.MethodGet,
			URL:         "http://foo.com/_utils?today=tomorrow",
			Status:      http.StatusMovedPermanently,
			HeaderMatch: map[string]string{"Location": "http://foo.com/_utils/?today=tomorrow"},
		},
		{
			Name:      "Utils",
			Method:    http.MethodGet,
			URL:       "http://foo.com/_utils/",
			Status:    http.StatusOK,
			BodyMatch: "Licensed under the Apache License",
		},
	}
	p, err := New(serverDSN(t))
	if err != nil {
		t.Fatalf("Failed to initialize proxy: %s", err)
	}
	p.StrictMethods = true
	for _, test := range tests {
		func(test getTest) {
			t.Run(test.Name, func(t *testing.T) {
				req := httptest.NewRequest(test.Method, test.URL, test.Body)
				w := httptest.NewRecorder()
				p.ServeHTTP(w, req)
				resp := w.Result()
				if resp.StatusCode != test.Status {
					t.Errorf("Expected status: %d/%s\n  Actual status: %d/%s", test.Status, http.StatusText(test.Status), resp.StatusCode, resp.Status)
				}
				body, _ := ioutil.ReadAll(resp.Body)
				if test.BodyMatch != "" && !bytes.Contains(body, []byte(test.BodyMatch)) {
					t.Errorf("Body does not contain '%s':\n%s", test.BodyMatch, body)
				}
				for header, value := range test.HeaderMatch {
					hv := resp.Header.Get(header)
					if hv != value {
						t.Errorf("Header %s: %s, expected %s", header, hv, value)
					}
				}
			})
		}(test)
	}
}

func TestProxyError(t *testing.T) {
	w := httptest.NewRecorder()
	ProxyError(w, errors.New("foo error"))
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	expected := "Proxy error: foo error"
	if string(body) != expected {
		t.Errorf("Expected: %s\n  Actual: %s\n", expected, body)
	}
}

func TestMethodAllowed(t *testing.T) {
	type maTest struct {
		Method   string
		Strict   bool
		Expected bool
	}
	tests := []maTest{
		{Method: "GET", Strict: true, Expected: true},
		{Method: "GET", Strict: false, Expected: true},
		{Method: "COPY", Strict: true, Expected: true},
		{Method: "COPY", Strict: false, Expected: true},
		{Method: "CHICKEN", Strict: true, Expected: false},
		{Method: "CHICKEN", Strict: false, Expected: true},
	}
	for _, test := range tests {
		func(test maTest) {
			t.Run(fmt.Sprintf("%s:%t", test.Method, test.Strict), func(t *testing.T) {
				t.Parallel()
				p := mustNew(t, "http://foo.com/")
				p.StrictMethods = test.Strict
				result := p.methodAllowed(test.Method)
				if result != test.Expected {
					t.Errorf("Expected: %t, Actual: %t", test.Expected, result)
				}
			})
		}(test)
	}
}

func TestCloneRequest(t *testing.T) {
	type cloneTest struct {
		Name     string
		Input    *http.Request
		Expected *http.Request
		Error    string
	}
	tests := []cloneTest{
		{
			Name:  "SimpleGet",
			Input: func() *http.Request { r, _ := http.NewRequest("GET", "http://zoo.com", nil); return r }(),
			Expected: func() *http.Request {
				r, _ := http.NewRequest("GET", "http://foo.com/", nil)
				return r.WithContext(context.Background())
			}(),
		},
		{
			Name:  "URLAuth",
			Input: func() *http.Request { r, _ := http.NewRequest("GET", "http://foo:bar@zoo.com", nil); return r }(),
			Expected: func() *http.Request {
				r, _ := http.NewRequest("GET", "http://foo:bar@foo.com/", nil)
				r = r.WithContext(context.Background())
				return r
			}(),
		},
		{
			Name: "BasicAuth",
			Input: func() *http.Request {
				r, _ := http.NewRequest("GET", "http://zoo.com", nil)
				r.SetBasicAuth("foo", "bar")
				return r
			}(),
			Expected: func() *http.Request {
				r, _ := http.NewRequest("GET", "http://foo.com/", nil)
				r = r.WithContext(context.Background())
				r.Header.Set("Authorization", "Basic Zm9vOmJhcg==")
				return r
			}(),
		},
	}
	p := mustNew(t, "http://foo.com/")
	for _, test := range tests {
		func(test cloneTest) {
			t.Run(test.Name, func(t *testing.T) {
				t.Parallel()
				result, err := p.cloneRequest(test.Input)
				var msg string
				if err != nil {
					msg = err.Error()
				}
				if msg != test.Error {
					t.Errorf("Expected error: %s\n  Actual error: %s", test.Error, msg)
				}
				if d := diff.Interface(test.Expected, result); d != "" {
					t.Error(d)
				}
			})
		}(test)
	}
}
