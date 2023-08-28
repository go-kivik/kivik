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
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"gitlab.com/flimzy/testy"
	"golang.org/x/net/publicsuffix"

	kivik "github.com/go-kivik/kivik/v4"
)

func TestCookieAuthAuthenticate(t *testing.T) {
	type cookieTest struct {
		dsn            string
		auth           *CookieAuth
		err            string
		status         int
		expectedCookie *http.Cookie
	}

	tests := testy.NewTable()
	tests.Add("success", func(t *testing.T) interface{} {
		var sessCounter int
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("Content-Type", "application/json")
			h.Set("Date", "Sat, 08 Sep 2018 15:49:29 GMT")
			h.Set("Server", "CouchDB/2.2.0 (Erlang OTP/19)")
			if r.URL.Path == "/_session" {
				sessCounter++
				if sessCounter > 1 {
					t.Fatal("Too many calls to /_session")
				}
				h.Set("Set-Cookie", "AuthSession=YWRtaW46NUI5M0VGODk6eLUGqXf0HRSEV9PPLaZX86sBYes; Version=1; Path=/; HttpOnly")
				w.WriteHeader(200)
				_, _ = w.Write([]byte(`{"ok":true,"name":"admin","roles":["_admin"]}`))
			} else {
				if cookie := r.Header.Get("Cookie"); cookie != "AuthSession=YWRtaW46NUI5M0VGODk6eLUGqXf0HRSEV9PPLaZX86sBYes" {
					t.Errorf("Expected cookie not found: %s", cookie)
				}
				w.WriteHeader(200)
				_, _ = w.Write([]byte(`{"ok":true}`))
			}
		}))
		return cookieTest{
			dsn:  s.URL,
			auth: &CookieAuth{Username: "foo", Password: "bar"},
			expectedCookie: &http.Cookie{
				Name:  kivik.SessionCookieName,
				Value: "YWRtaW46NUI5M0VGODk6eLUGqXf0HRSEV9PPLaZX86sBYes",
			},
		}
	})
	tests.Add("cookie not set", func(t *testing.T) interface{} {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("Content-Type", "application/json")
			h.Set("Date", "Sat, 08 Sep 2018 15:49:29 GMT")
			h.Set("Server", "CouchDB/2.2.0 (Erlang OTP/19)")
			w.WriteHeader(200)
		}))
		return cookieTest{
			dsn:  s.URL,
			auth: &CookieAuth{Username: "foo", Password: "bar"},
		}
	})

	tests.Run(t, func(t *testing.T, test cookieTest) {
		c, err := New(&http.Client{}, test.dsn, nil)
		if err != nil {
			t.Fatal(err)
		}
		if e := c.Auth(test.auth); e != nil {
			t.Fatal(e)
		}
		_, err = c.DoError(context.Background(), "GET", "/foo", nil)
		testy.StatusError(t, test.err, test.status, err)
		if d := testy.DiffInterface(test.expectedCookie, test.auth.Cookie()); d != nil {
			t.Error(d)
		}

		// Do it again; should be idempotent
		_, err = c.DoError(context.Background(), "GET", "/foo", nil)
		testy.StatusError(t, test.err, test.status, err)
		if d := testy.DiffInterface(test.expectedCookie, test.auth.Cookie()); d != nil {
			t.Error(d)
		}
	})
}

func TestCookie(t *testing.T) {
	tests := []struct {
		name     string
		auth     *CookieAuth
		expected *http.Cookie
	}{
		{
			name:     "No cookie jar",
			auth:     &CookieAuth{},
			expected: nil,
		},
		{
			name:     "No dsn",
			auth:     &CookieAuth{},
			expected: nil,
		},
		{
			name:     "no cookies",
			auth:     &CookieAuth{},
			expected: nil,
		},
		{
			name: "cookie found",
			auth: func() *CookieAuth {
				dsn, err := url.Parse("http://example.com/")
				if err != nil {
					t.Fatal(err)
				}
				jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
				if err != nil {
					t.Fatal(err)
				}
				jar.SetCookies(dsn, []*http.Cookie{
					{Name: kivik.SessionCookieName, Value: "foo"},
					{Name: "other", Value: "bar"},
				})
				return &CookieAuth{
					client: &Client{
						dsn: dsn,
						Client: &http.Client{
							Jar: jar,
						},
					},
				}
			}(),
			expected: &http.Cookie{Name: kivik.SessionCookieName, Value: "foo"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.auth.Cookie()
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

type dummyJar []*http.Cookie

var _ http.CookieJar = &dummyJar{}

func (j dummyJar) Cookies(_ *url.URL) []*http.Cookie {
	return []*http.Cookie(j)
}

func (j *dummyJar) SetCookies(_ *url.URL, cookies []*http.Cookie) {
	*j = cookies
}

func Test_shouldAuth(t *testing.T) {
	type tt struct {
		a    *CookieAuth
		req  *http.Request
		want bool
	}

	tests := testy.NewTable()
	tests.Add("no session", tt{
		a:    &CookieAuth{},
		req:  httptest.NewRequest("GET", "/", nil),
		want: true,
	})
	tests.Add("authed request", func() interface{} {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: kivik.SessionCookieName})
		return tt{
			a:    &CookieAuth{},
			req:  req,
			want: false,
		}
	})
	tests.Add("valid session", func() interface{} {
		c, _ := New(&http.Client{}, "http://example.com/", nil)
		c.Jar = &dummyJar{&http.Cookie{
			Name:    kivik.SessionCookieName,
			Expires: time.Now().Add(20 * time.Minute),
		}}
		a := &CookieAuth{client: c}

		return tt{
			a:    a,
			req:  httptest.NewRequest("GET", "/", nil),
			want: false,
		}
	})
	tests.Add("expired session", func() interface{} {
		c, _ := New(&http.Client{}, "http://example.com/", nil)
		c.Jar = &dummyJar{&http.Cookie{
			Name:    kivik.SessionCookieName,
			Expires: time.Now().Add(-20 * time.Second),
		}}
		a := &CookieAuth{client: c}

		return tt{
			a:    a,
			req:  httptest.NewRequest("GET", "/", nil),
			want: true,
		}
	})
	tests.Add("no expiry time", func() interface{} {
		c, _ := New(&http.Client{}, "http://example.com/", nil)
		c.Jar = &dummyJar{&http.Cookie{
			Name: kivik.SessionCookieName,
		}}
		a := &CookieAuth{client: c}

		return tt{
			a:    a,
			req:  httptest.NewRequest("GET", "/", nil),
			want: false,
		}
	})
	tests.Add("about to expire", func() interface{} {
		c, _ := New(&http.Client{}, "http://example.com/", nil)
		c.Jar = &dummyJar{&http.Cookie{
			Name:    kivik.SessionCookieName,
			Expires: time.Now().Add(20 * time.Second),
		}}
		a := &CookieAuth{client: c}

		return tt{
			a:    a,
			req:  httptest.NewRequest("GET", "/", nil),
			want: true,
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		got := tt.a.shouldAuth(tt.req)
		if got != tt.want {
			t.Errorf("Want %t, got %t", tt.want, got)
		}
	})
}

func Test401Response(t *testing.T) {
	var sessCounter, getCounter int
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Type", "application/json")
		h.Set("Date", "Sat, 08 Sep 2018 15:49:29 GMT")
		h.Set("Server", "CouchDB/2.2.0 (Erlang OTP/19)")
		if r.URL.Path == "/_session" {
			sessCounter++
			if sessCounter > 2 {
				t.Fatal("Too many calls to /_session")
			}
			var cookie string
			if sessCounter == 1 {
				// set another cookie at the start too
				h.Add("Set-Cookie", "Other=foo; Version=1; Path=/; HttpOnly")
				cookie = "First"
			} else {
				cookie = "Second"
			}
			h.Add("Set-Cookie", "AuthSession="+cookie+"; Version=1; Path=/; HttpOnly")
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{"ok":true,"name":"admin","roles":["_admin"]}`))
		} else {
			getCounter++
			cookie := r.Header.Get("Cookie")
			if !(strings.Contains(cookie, "AuthSession=")) {
				t.Errorf("Expected cookie not found: %s", cookie)
			}
			// because of the way the request is baked before the auth loop
			// cookies other than the auth cookie set when calling _session won't
			// get applied to requests until after that first request.
			if getCounter > 1 && !strings.Contains(cookie, "Other=foo") {
				t.Errorf("Expected cookie not found: %s", cookie)
			}
			if getCounter == 2 {
				w.WriteHeader(401)
				_, _ = w.Write([]byte(`{"error":"unauthorized","reason":"You are not authorized to access this db."}`))
				return
			}
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{"ok":true}`))
		}
	}))

	c, err := New(&http.Client{}, s.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	auth := &CookieAuth{Username: "foo", Password: "bar"}
	if e := c.Auth(auth); e != nil {
		t.Fatal(e)
	}

	expectedCookie := &http.Cookie{
		Name:  kivik.SessionCookieName,
		Value: "First",
	}
	newCookie := &http.Cookie{
		Name:  kivik.SessionCookieName,
		Value: "Second",
	}

	_, err = c.DoError(context.Background(), "GET", "/foo", nil)
	testy.StatusError(t, "", 0, err)
	if d := testy.DiffInterface(expectedCookie, auth.Cookie()); d != nil {
		t.Error(d)
	}

	_, err = c.DoError(context.Background(), "GET", "/foo", nil)

	// this causes a skip so this won't work for us.
	// testy.StatusError(t, "Unauthorized: You are not authorized to access this db.", 401, err)
	if !testy.ErrorMatches("Unauthorized: You are not authorized to access this db.", err) {
		t.Fatalf("Unexpected error: %s", err)
	}
	if status := testy.StatusCode(err); status != http.StatusUnauthorized {
		t.Errorf("Unexpected status code: %d", status)
	}

	var noCookie *http.Cookie
	if d := testy.DiffInterface(noCookie, auth.Cookie()); d != nil {
		t.Error(d)
	}

	_, err = c.DoError(context.Background(), "GET", "/foo", nil)
	testy.StatusError(t, "", 0, err)
	if d := testy.DiffInterface(newCookie, auth.Cookie()); d != nil {
		t.Error(d)
	}
}
