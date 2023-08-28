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

package test

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-kivik/couchdb/v4/chttp"
	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.Register("Session", session)
}

func session(ctx *kt.Context) {
	chttpAdmin, err := chttp.New(&http.Client{}, ctx.Admin.DSN(), nil)
	if err != nil {
		ctx.Fatalf("chttp.Admin failed: %s", err)
	}
	chttpNoAuth, err := chttp.New(&http.Client{}, ctx.NoAuth.DSN(), nil)
	if err != nil {
		ctx.Fatalf("chttp.NoAuth failed: %s", err)
	}
	ctx.Run("Get", func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			testSession(ctx, chttpAdmin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			testSession(ctx, chttpNoAuth)
		})
	})
	ctx.Run("Post", func(ctx *kt.Context) {
		testCreateSession(ctx, chttpNoAuth)
	})
	ctx.Run("Delete", func(ctx *kt.Context) {
		testDeleteSession(ctx, chttpNoAuth)
	})
}

func testSession(ctx *kt.Context, client *chttp.Client) {
	ctx.Parallel()
	if client == nil {
		ctx.Skipf("No CHTTP client")
	}
	uCtx := struct {
		Info struct {
			AuthMethod   string   `json:"authenticated"`
			AuthDB       string   `json:"authentication_db"`
			AuthHandlers []string `json:"authentication_handlers"`
		} `json:"info"`
		OK      bool `json:"ok"`
		UserCtx struct {
			Name  string   `json:"name"`
			Roles []string `json:"roles"`
		} `json:"userCtx"`
	}{}
	err := client.DoJSON(context.Background(), http.MethodGet, "/_session", nil, &uCtx)
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	values := map[string]string{
		"info.authenticated":           uCtx.Info.AuthMethod,
		"info.authentication_db":       uCtx.Info.AuthDB,
		"info.authentication_handlers": strings.Join(uCtx.Info.AuthHandlers, ","),
		"ok":                           fmt.Sprintf("%t", uCtx.OK),
		"userCtx.roles":                strings.Join(uCtx.UserCtx.Roles, ","),
	}
	for key, actual := range values {
		expected := ctx.MustString(key)
		if actual != expected {
			ctx.Errorf("Unexpected value for `%s`. Expected '%s', actual '%s'", key, expected, actual)
		}
	}
	dsn, _ := url.Parse(client.DSN())
	var expected string
	if dsn.User != nil {
		expected = dsn.User.Username()
	}
	actual := uCtx.UserCtx.Name
	if actual != expected {
		ctx.Errorf("Unexpected value for `%s`. Expected '%s', actual '%s'", "userCtx.name", expected, actual)
	}
}

type sessionPostTest struct {
	Name    string
	Query   string
	Options *chttp.Options
	// True if the test requires valid credentials
	Creds bool
}

func testCreateSession(ctx *kt.Context, client *chttp.Client) {
	if client == nil {
		ctx.Skipf("No CHTTP client")
	}
	// Re-create client, so we can override defaults
	client, _ = chttp.New(&http.Client{}, client.DSN(), nil)
	// Don't follow redirect
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	var name, password string
	if ctx.Admin != nil {
		if dsn, _ := url.Parse(ctx.Admin.DSN()); dsn.User != nil {
			name = dsn.User.Username()
			password, _ = dsn.User.Password()
		}
	}
	tests := []sessionPostTest{
		{Name: "EmptyJSON", Options: &chttp.Options{ContentType: "application/json"}},
		{Name: "BadJSON", Options: &chttp.Options{
			ContentType: "application/json",
			Body:        kt.Body("oink"),
		}},
		{Name: "BogusTypeJSON", Creds: true, Options: &chttp.Options{
			ContentType: "image/gif",
			Body:        kt.Body(`{"name":"%s","password":"%s"}`, name, password),
		}},
		{Name: "BogusTypeForm", Creds: true, Options: &chttp.Options{
			ContentType: "image/gif",
			Body:        kt.Body(`name=%s&password=%s`, name, password),
		}},
		{Name: "EmptyForm", Options: &chttp.Options{ContentType: "application/x-www-form-urlencoded"}},
		{Name: "BadForm", Options: &chttp.Options{
			ContentType: "application/x-www-form-urlencoded",
			Body:        kt.Body("o\\ink"),
		}},
		{Name: "MeaninglessJSON", Options: &chttp.Options{
			ContentType: "application/json",
			Body:        kt.Body(`{"ok":true}`),
		}},
		{Name: "MeaninglessForm", Options: &chttp.Options{
			ContentType: "application/x-www-form-urlencoded",
			Body:        kt.Body("ok=true"),
		}},
		{Name: "GoodJSON", Options: &chttp.Options{
			ContentType: "application/json",
			Body:        kt.Body(`{"name":"bob","password":"abc123"}`),
		}},
		{Name: "BadQueryParam", Query: "foobarbaz!", Options: &chttp.Options{
			ContentType: "application/json",
			Body:        kt.Body(`{"name":"bob","password":"abc123"}`),
		}},
		{Name: "GoodCredsJSON", Creds: true, Options: &chttp.Options{
			ContentType: "application/json",
			Body:        kt.Body(fmt.Sprintf(`{"name":"%s","password":"%s"}`, name, password)),
		}},
		{Name: "GoodCredsForm", Creds: true, Options: &chttp.Options{
			ContentType: "application/x-www-form-urlencoded",
			Body:        kt.Body(fmt.Sprintf(`name=%s&password=%s`, name, password)),
		}},
		{Name: "BadCredsJSON", Creds: true, Options: &chttp.Options{
			ContentType: "application/json",
			Body:        kt.Body(fmt.Sprintf(`{"name":"%s","password":"%sxxx"}`, name, password)),
		}},
		{Name: "BadCredsForm", Creds: true, Options: &chttp.Options{
			ContentType: "application/x-www-form-urlencoded",
			Body:        kt.Body(`name=%s&password=%sxxx`, name, password),
		}},
		{Name: "GoodCredsJSONRedirEmpty", Creds: true, Query: "next=", Options: &chttp.Options{
			ContentType: "application/json",
			Body:        kt.Body(`{"name":"%s","password":"%s"}`, name, password),
		}},
		{Name: "GoodCredsJSONRedirRelative", Creds: true, Query: "next=/_session", Options: &chttp.Options{
			ContentType: "application/json",
			Body:        kt.Body(`{"name":"%s","password":"%s"}`, name, password),
		}},
		{Name: "GoodCredsJSONRedirSchemaless", Creds: true, Query: "next=//_session", Options: &chttp.Options{
			ContentType: "application/json",
			Body:        kt.Body(`{"name":"%s","password":"%s"}`, name, password),
		}},
		{Name: "GoodCredsJSONRedirRelativeNoSlash", Creds: true, Query: "next=foobar", Options: &chttp.Options{
			ContentType: "application/json",
			Body:        kt.Body(`{"name":"%s","password":"%s"}`, name, password),
		}},
		{Name: "GoodCredsJSONRemoteRedirAbsolute", Creds: true, Query: "next=http://google.com/", Options: &chttp.Options{
			ContentType: "application/json",
			Body:        kt.Body(`{"name":"%s","password":"%s"}`, name, password),
		}},
		{Name: "GoodCredsJSONRemoteRedirInvalidURL", Creds: true, Query: "next=/session%25%26%26", Options: &chttp.Options{
			ContentType: "application/json",
			Body:        kt.Body(`{"name":"%s","password":"%s"}`, name, password),
		}},
		{Name: "GoodCredsJSONRemoteRedirHeaderInjection", Creds: true, Query: "next=/foo\nX-Injected: oink", Options: &chttp.Options{
			ContentType: "application/json",
			Body:        kt.Body(`{"name":"%s","password":"%s"}`, name, password),
		}},
		{Name: "AcceptPlain", Creds: true, Options: &chttp.Options{
			ContentType: "application/json", Accept: "text/plain",
			Body: kt.Body(`{"name":"%s","password":"%s"}`, name, password),
		}},
		{Name: "AcceptImage", Creds: true, Options: &chttp.Options{
			ContentType: "application/json", Accept: "image/gif",
			Body: kt.Body(`{"name":"%s","password":"%s"}`, name, password),
		}},
	}
	for _, postTest := range tests {
		func(test sessionPostTest) {
			ctx.Run(test.Name, func(ctx *kt.Context) {
				if test.Creds && name == "" {
					ctx.Skipf("Credentials required but missing, skipping test.")
				}
				ctx.Parallel()
				reqURL := "/_session"
				if test.Query != "" {
					reqURL += "?" + test.Query
				}
				r, err := client.DoReq(context.Background(), http.MethodPost, reqURL, test.Options)
				if err == nil {
					err = chttp.ResponseError(r)
				}
				if !ctx.IsExpectedSuccess(err) {
					return
				}
				defer r.Body.Close() // nolint: errcheck
				if _, ok := r.Header["Cache-Control"]; !ok {
					ctx.Errorf("No Cache-Control set in response.")
				} else {
					cc := r.Header.Get("Cache-Control")
					if strings.ToLower(cc) != "must-revalidate" {
						ctx.Errorf("Expected Cache-Control: must-revalidate, but got'%s", cc)
					}
				}
				if strings.HasPrefix(test.Query, "next=") {
					if r.StatusCode != http.StatusFound {
						ctx.Errorf("Expected redirect")
					} else {
						q, _ := url.ParseQuery(test.Query)
						loc := r.Header.Get("Location")
						next := q.Get("next")
						if !strings.HasSuffix(loc, next) {
							ctx.Errorf("Expected Location: ...%s, got: %s", next, loc)
						}
					}
				}
				cookies := r.Cookies()
				if len(cookies) != 1 {
					ctx.Errorf("Expected 1 cookie, got %d", len(cookies))
				}
				if cookies[0].Name != kivik.SessionCookieName {
					ctx.Errorf("Server set cookie '%s', expected '%s'", cookies[0].Name, kivik.SessionCookieName)
				}
				if !cookies[0].HttpOnly {
					ctx.Errorf("Cookie is not set HttpOnly")
				}
				if cookies[0].Path != "/" {
					ctx.Errorf("Unexpected cookie path. Got '%s', expected '/'", cookies[0].Path)
				}
				val, err := base64.RawURLEncoding.DecodeString(cookies[0].Value)
				if err != nil {
					ctx.Fatalf("Failed to decode cookie value: %s", err)
				}
				parts := strings.SplitN(string(val), ":", 3) // nolint:gomnd
				if parts[0] != name {
					ctx.Errorf("Cookie does not match username. Want '%s', got '%s'", name, parts[0])
				}
				if _, err := hex.DecodeString(parts[1]); err != nil {
					ctx.Errorf("Failed to decode cookie timestamp: %s", err)
				}
				response := struct {
					OK    bool     `json:"ok"`
					Name  *string  `json:"name"`
					Roles []string `json:"roles"`
				}{}
				if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
					ctx.Fatalf("Failed to decode response: %s", err)
				}
				if !response.OK {
					ctx.Errorf("Expected OK response")
				}
				if response.Name != nil && *response.Name != name {
					ctx.Errorf("Unexpected name in response. Expected '%s', got '%s'", name, *response.Name)
				}
			})
		}(postTest)
	}
}

type deleteSessionTest struct {
	Name   string
	Creds  bool
	Cookie *http.Cookie
}

func testDeleteSession(ctx *kt.Context, client *chttp.Client) {
	ctx.Parallel()
	if client == nil {
		ctx.Skipf("No CHTTP client")
	}
	// Re-create client, so we can override defaults
	client, _ = chttp.New(&http.Client{}, client.DSN(), nil)
	// Don't save sessions
	client.Jar = nil
	var cookie *http.Cookie
	if ctx.Admin != nil {
		if dsn, _ := url.Parse(ctx.Admin.DSN()); dsn.User != nil {
			name := dsn.User.Username()
			password, _ := dsn.User.Password()
			r, err := client.DoReq(context.Background(), http.MethodPost, "/_session", &chttp.Options{
				Body: kt.Body(`{"name":"%s","password":"%s"}`, name, password),
			})
			if err != nil {
				ctx.Errorf("Failed to establish session: %s", err)
				return
			}
			for _, c := range r.Cookies() {
				if c.Name == kivik.SessionCookieName {
					cookie = c
					break
				}
			}
		}
	}
	tests := []deleteSessionTest{
		{Name: "ValidSession", Creds: true, Cookie: cookie},
		{Name: "NoSession"},
		// {Name: "InvalidSession"},
		// {Name: "ExpiredSession"},
	}
	for _, test := range tests {
		func(test deleteSessionTest) {
			ctx.Run(test.Name, func(ctx *kt.Context) {
				if test.Creds && cookie == nil {
					ctx.Skipf("Credentials required but missing, skipping test.")
				}
				response := struct {
					OK bool `json:"ok"`
				}{}
				req, err := client.NewRequest(context.Background(), http.MethodDelete, "/_session", nil, nil)
				if err != nil {
					ctx.Fatalf("Failed to create request: %s", err)
				}
				if test.Cookie != nil {
					req.AddCookie(test.Cookie)
				}
				r, err := client.Do(req)
				if err == nil {
					err = chttp.ResponseError(r)
				}
				if err == nil {
					defer r.Body.Close() // nolint: errcheck
					err = json.NewDecoder(r.Body).Decode(&response)
				}
				if !ctx.IsExpectedSuccess(err) {
					return
				}
				if _, ok := r.Header["Cache-Control"]; !ok {
					ctx.Errorf("No Cache-Control set in response.")
				} else {
					cc := r.Header.Get("Cache-Control")
					if strings.ToLower(cc) != "must-revalidate" {
						ctx.Errorf("Expected Cache-Control: must-revalidate, but got'%s", cc)
					}
				}
				for _, c := range r.Cookies() {
					if c.Name == kivik.SessionCookieName {
						if c.Value != "" {
							ctx.Errorf("Expected empty cookie value, got '%s'", c.Value)
						}
						break
					}
				}
			})
		}(test)
	}
}
