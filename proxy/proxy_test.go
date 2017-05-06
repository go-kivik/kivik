package proxy

import "testing"

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
