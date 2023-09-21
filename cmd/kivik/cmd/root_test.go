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
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/xkivik/v4/cmd/kivik/errors"
	"github.com/go-kivik/xkivik/v4/cmd/kivik/log"
)

func gunzip(next testy.RequestValidator) testy.RequestValidator {
	return func(t *testing.T, r *http.Request) {
		t.Helper()
		if r.Header.Get("Content-Encoding") == "gzip" {
			r.Header.Del("Content-Encoding")
			gun, err := gzip.NewReader(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(gun)
			if err != nil {
				t.Fatal(err)
			}
			r.Body = io.NopCloser(bytes.NewReader(body))
		}
		next(t, r)
	}
}

func Test_root_RunE(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("unknown flag", cmdTest{
		args:   []string{"--bogus"},
		status: errors.ErrUsage,
	})
	tests.Add("unknown command", cmdTest{
		args:   []string{"bogus"},
		status: errors.ErrUsage,
	})
	tests.Add("Debug long", cmdTest{
		args:   []string{"--debug"},
		status: errors.ErrUsage,
	})
	tests.Add("Debug short", cmdTest{
		args:   []string{"--debug"},
		status: errors.ErrUsage,
	})
	tests.Add("context from config file", cmdTest{
		args: []string{"--debug", "--config", "./testdata/localhost.yaml"},
	})
	tests.Add("invalid timeout", cmdTest{
		args:   []string{"--request-timeout", "-78"},
		status: errors.ErrUsage,
	})
	tests.Add("timeout", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Body: io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
		}, func(*testing.T, *http.Request) {
			time.Sleep(time.Second)
		})

		return cmdTest{
			args:   []string{"--config", "./testdata/localhost.yaml", "ping", s.URL, "--request-timeout", "1ms"},
			status: errors.ErrUnavailable,
		}
	})
	tests.Add("retry", cmdTest{
		args:   []string{"--retry", "3", "ping", "http://localhost:5984"},
		status: errors.ErrUnavailable,
	})
	tests.Add("retry delay invalid", cmdTest{
		args:   []string{"--retry", "3", "--retry-delay", "oink", "ping", "http://localhost:5984"},
		status: errors.ErrUsage,
	})
	tests.Add("retry delay", cmdTest{
		args:   []string{"--retry", "3", "--retry-delay", "15ms", "ping", "http://localhost:5984"},
		status: errors.ErrUnavailable,
	})
	tests.Add("disable retry delay", cmdTest{
		args:   []string{"--retry", "3", "--retry-delay", "0", "ping", "http://localhost:5984"},
		status: errors.ErrUnavailable,
	})
	tests.Add("connect timeout invalid", cmdTest{
		args:   []string{"--connect-timeout", "oink", "ping", "http://localhost:5984"},
		status: errors.ErrUsage,
	})
	tests.Add("retry max time", cmdTest{
		args:   []string{"--retry", "100", "--retry-delay", "40ms", "--retry-timeout", "100ms", "ping", "http://localhost:5984"},
		status: errors.ErrUnavailable,
	})
	tests.Add("options", cmdTest{
		args:   []string{"--debug", "-O", "foo=bar", "--option", "bar=baz", "ping", "http://localhost:5984/"},
		status: errors.ErrUnavailable,
	})
	tests.Add("options from url", cmdTest{
		args:   []string{"--debug", "-O", "foo=bar", "--option", "bar=baz", "ping", "http://localhost:5984/?baz=qux"},
		status: errors.ErrUnavailable,
	})
	tests.Add("ping file url", cmdTest{
		args: []string{"--debug", "ping", "./testdata"},
	})
	tests.Add("view cleanup", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Body: io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
		}, func(*testing.T, *http.Request) {
			time.Sleep(time.Second)
		})

		return cmdTest{
			args: []string{"view-cleanup", s.URL + "/foo/_view_cleanup"},
		}
	})
	tests.Add("flush", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Body: io.NopCloser(strings.NewReader(`{"ok":true}`)),
		}, func(t *testing.T, req *http.Request) {
			if req.Method != http.MethodPost {
				t.Errorf("Unexpected method: %v", req.Method)
			}
			if req.URL.Path != "/foo/_ensure_full_commit" {
				t.Errorf("Unexpected path: %s", req.URL.Path)
			}
		})

		return cmdTest{
			args: []string{"flush", s.URL + "/foo"},
		}
	})
	tests.Add("compact", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Body: io.NopCloser(strings.NewReader(`{"ok":true}`)),
		}, func(t *testing.T, req *http.Request) {
			if req.Method != http.MethodPost {
				t.Errorf("Unexpected method: %v", req.Method)
			}
			if req.URL.Path != "/asdf/_compact" {
				t.Errorf("Unexpected path: %s", req.URL.Path)
			}
		})

		return cmdTest{
			args: []string{"compact", s.URL + "/asdf"},
		}
	})
	tests.Add("compact views", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Body: io.NopCloser(strings.NewReader(`{"ok":true}`)),
		}, func(t *testing.T, req *http.Request) {
			if req.Method != http.MethodPost {
				t.Errorf("Unexpected method: %v", req.Method)
			}
			if req.URL.Path != "/asdf/_compact/quack" {
				t.Errorf("Unexpected path: %s", req.URL.Path)
			}
		})

		return cmdTest{
			args: []string{"compact-views", s.URL + "/asdf/quack"},
		}
	})
	tests.Add("purge", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Body: io.NopCloser(strings.NewReader(`{"ok":true}`)),
		}, gunzip(func(t *testing.T, req *http.Request) {
			if req.Method != http.MethodPost {
				t.Errorf("Unexpected method: %v", req.Method)
			}
			if req.URL.Path != "/asdf/_purge" {
				t.Errorf("Unexpected path: %s", req.URL.Path)
			}
			if d := testy.DiffAsJSON(testy.Snapshot(t), req.Body); d != nil {
				t.Error(d)
			}
		}))

		return cmdTest{
			args: []string{"purge", s.URL + "/asdf/quack", "--revs", "1-xyz"},
		}
	})
	tests.Add("purge --data", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Body: io.NopCloser(strings.NewReader(`{"ok":true}`)),
		}, gunzip(func(t *testing.T, req *http.Request) {
			if req.Method != http.MethodPost {
				t.Errorf("Unexpected method: %v", req.Method)
			}
			if req.URL.Path != "/asdf/_purge" {
				t.Errorf("Unexpected path: %s", req.URL.Path)
			}
			if d := testy.DiffAsJSON(testy.Snapshot(t), req.Body); d != nil {
				t.Error(d)
			}
		}))

		return cmdTest{
			args: []string{"purge", s.URL + "/asdf", "--data", `{"foo":["1-xx","2-yy"]}`},
		}
	})

	res := []testy.Replacement{
		{
			Regexp:      regexp.MustCompile(`time: invalid duration oink`),
			Replacement: `time: invalid duration "oink"`,
		},
		{
			Regexp:      regexp.MustCompile(`Will retry in \d\.\d+s`),
			Replacement: `Will retry in X.XXs`,
		},
	}
	tests.Run(t, func(t *testing.T, tt cmdTest) {
		tt.Test(t, res...)
	})
}

type cmdTest struct {
	args   []string
	stdin  string
	status int
}

var standardReplacements = []testy.Replacement{
	{
		Regexp:      regexp.MustCompile(`http://127\.0\.0\.1:\d+/`),
		Replacement: "http://127.0.0.1:XXX/",
	},
	{
		Regexp:      regexp.MustCompile(`Date: .*`),
		Replacement: `Date: XXX`,
	},
	{
		Regexp:      regexp.MustCompile(`Host: \S*`),
		Replacement: `Host: XXX`,
	},
	{
		Regexp:      regexp.MustCompile(`go\d\.\d+\.\d+`),
		Replacement: `goX.XX.X`,
	},
}

func (tt *cmdTest) Test(t *testing.T, re ...testy.Replacement) {
	t.Helper()
	lg := log.New()
	root := rootCmd(lg)
	root.resolveHome = func(i string) string { return i }

	root.cmd.SetArgs(tt.args)
	var status int
	stdout, stderr := testy.RedirIO(strings.NewReader(tt.stdin), func() {
		status = root.execute(context.Background())
	})
	repl := append(standardReplacements, re...) //nolint:gocritic
	if d := testy.DiffText(testy.Snapshot(t, "_stdout"), stdout, repl...); d != nil {
		t.Errorf("STDOUT: %s", d)
	}
	if d := testy.DiffText(testy.Snapshot(t, "_stderr"), stderr, repl...); d != nil {
		t.Errorf("STDERR: %s", d)
	}
	if tt.status != status {
		t.Errorf("Unexpected exit status. Want %d, got %d", tt.status, status)
	}
}

func Test_parseTimeout(t *testing.T) {
	type tt struct {
		input string
		want  string
		err   string
	}

	tests := testy.NewTable()
	tests.Add("empty", tt{
		want: "0s",
	})
	tests.Add("invalid", tt{
		input: "bogus",
		err:   `time: invalid duration "?bogus"?`,
	})
	tests.Add("ms", tt{
		input: "100ms",
		want:  "100ms",
	})
	tests.Add("default to seconds", tt{
		input: "15",
		want:  "15s",
	})
	tests.Add("negative", tt{
		input: "-1.5s",
		err:   "negative timeout not permitted",
	})
	tests.Add("negative seconds", tt{
		input: "-1.5",
		err:   "negative timeout not permitted",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		got, err := parseDuration(tt.input)
		testy.ErrorRE(t, tt.err, err)
		if got.String() != tt.want {
			t.Errorf("Want: %s\n Got: %s", tt.want, got)
		}
	})
}

func Test_fmtDuration(t *testing.T) {
	type tt struct {
		d    time.Duration
		want string
	}

	tests := testy.NewTable()
	tests.Add("1.8s", tt{
		d:    1800 * time.Millisecond,
		want: "1.80s",
	})
	tests.Add("3m2s", tt{
		d:    182 * time.Second,
		want: "3m2s",
	})
	tests.Add("3m", tt{
		d:    3 * time.Minute,
		want: "3m0s",
	})
	tests.Add("1h3m4s", tt{
		d:    63*time.Minute + 4*time.Second,
		want: "1h3m",
	})
	tests.Add("3d1h3m4s", tt{
		d:    3*24*time.Hour + 63*time.Minute + 4*time.Second,
		want: "3d1h3m",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		got := fmtDuration(tt.d)
		if got != tt.want {
			t.Errorf("Want: %s\n Got: %s", tt.want, got)
		}
	})
}

func Test_resolveHome(t *testing.T) {
	t.Run("~ path", func(t *testing.T) {
		usr, _ := user.Current()
		want := filepath.Join(usr.HomeDir, "foo")
		got := resolveHome("~/foo")
		if got != want {
			t.Errorf("Unexpected result: %s", got)
		}
	})
	t.Run("no ~ in path", func(t *testing.T) {
		want := "asdf/foo"
		got := resolveHome(want)
		if got != want {
			t.Errorf("Unexpected result: %s", got)
		}
	})
}
