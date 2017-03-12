package serve

import (
	"net/http"
	"testing"
)

type redirTest struct {
	Name     string
	Input    string
	Expected string
	Err      string
}

/*

{Name: "GoodCredsJSONRedirRelative", Creds: true, Query: "next=/_session", Options: &chttp.Options{ContentType: "application/json",
    Body: bytes.NewBuffer([]byte(fmt.Sprintf(`{"name":"%s","password":"%s"}`, name, password))),
}},

*/

func TestRedirectURL(t *testing.T) {
	tests := []redirTest{
		{Name: "NoURL", Input: "-"},
		{Name: "EmptyValue", Input: "", Err: "400 redirection url must be relative to server root"},
		{Name: "Absolute", Input: "http://google.com/", Err: "400 redirection url must be relative to server root"},
		{Name: "HeaderInjection", Input: "next=/foo\nX-Injected: oink", Err: "400 redirection url must be relative to server root"},
		{Name: "InvalidURL", Input: "://google.com/", Err: "400 redirection url must be relative to server root"},
		{Name: "NoSlash", Input: "foobar", Err: "400 redirection url must be relative to server root"},
		{Name: "Relative", Input: "/_session", Expected: "/_session"},
		{Name: "InvalidRelative", Input: "/session%25%26%26", Err: "400 invalid redirection url"},
		{Name: "Schemaless", Input: "//evil.org", Err: "400 invalid redirection url"},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			url := "/"
			if test.Input != "-" {
				url += "?next=" + test.Input
			}
			r, _ := http.NewRequest("GET", url, nil)
			result, err := redirectURL(r)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if test.Err != errMsg {
				t.Errorf("Unexpected error result. Expected '%s', got '%s'", test.Err, errMsg)
			}
			if test.Expected != result {
				t.Errorf("Unexpected result. Expected '%s', got '%s'", test.Expected, result)
			}
		})
	}
}
