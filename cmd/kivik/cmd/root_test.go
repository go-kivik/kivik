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
	"io"
	"net/http"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/cmd/kivik/log"
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
		Regexp:      regexp.MustCompile(`go\d\.\d+\.?[\da-z-]+`),
		Replacement: `goX.XX.X`,
	},
	{
		Regexp:      regexp.MustCompile(`dial tcp 127\.0\.0\.1:`),
		Replacement: "dial tcp [::1]:",
	},
	{
		Regexp:      regexp.MustCompile(`: (dial.*i/o timeout|context deadline exceeded) \(Client.Timeout`),
		Replacement: ": ... (Client.Timeout)",
	},
	{
		Regexp:      regexp.MustCompile(`User-Agent: Kivik/` + kivik.Version),
		Replacement: "User-Agent: Kivik/X.Y.Z",
	},
	{
		Regexp:      regexp.MustCompile(`kivik version ` + kivik.Version),
		Replacement: "kivik version X.Y.Z",
	},
	{
		Regexp:      regexp.MustCompile(`"version": "` + kivik.Version),
		Replacement: `"version": "X.Y.Z`,
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
		status = root.execute(t.Context())
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
		if !testy.ErrorMatchesRE(tt.err, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if err != nil {
			return
		}
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
