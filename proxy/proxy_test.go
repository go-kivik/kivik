package proxy

import (
	"net/url"
	"testing"
)

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
