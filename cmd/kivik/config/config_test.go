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

package config

import (
	"net/url"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"gitlab.com/flimzy/testy"
	"gopkg.in/yaml.v3"

	"github.com/go-kivik/xkivik/v4/cmd/kivik/log"
)

func Test_unmarshalContext(t *testing.T) {
	type tt struct {
		input string
		err   string
	}

	tests := testy.NewTable()
	tests.Add("invalid YAML", tt{
		input: "- [",
		err:   "yaml: line 1: did not find expected node content",
	})
	tests.Add("long context", tt{
		input: `
name: long
scheme: https
host: localhost:5984
user: admin
password: abc123
database: foo
`,
	})
	tests.Add("invalid DSN", tt{
		input: `
name: short
dsn: https://admin:%xxx@localhost:5984/somedb
`,
		err: `parse "https://admin:%xxx@localhost:5984/somedb": invalid URL escape "%xx"`,
	})
	tests.Add("full DSN", tt{
		input: `
name: short
dsn: https://admin:abc123@localhost:5984/somedb
`,
	})
	tests.Run(t, func(t *testing.T, tt tt) {
		cx := &Context{}
		err := yaml.Unmarshal([]byte(tt.input), cx)
		testy.Error(t, tt.err, err)
		if d := testy.DiffInterface(testy.Snapshot(t), cx); d != nil {
			t.Error(d)
		}
	})
}

func TestConfig_Read(t *testing.T) {
	type tt struct {
		filename string
		env      map[string]string
		err      string
	}

	tests := testy.NewTable()
	tests.Add("no config file", tt{})
	tests.Add("other read error", tt{
		filename: "foo\x00bar",
		err:      "open foo\x00bar: invalid argument",
	})
	tests.Add("not regular file", tt{
		filename: "./testdata",
		err:      "yaml: input error: read ./testdata: is a directory",
	})
	tests.Add("file not found", tt{
		filename: "not found",
	})
	tests.Add("invalid env", tt{
		env: map[string]string{
			"KIVIKDSN": "http://foo.com/%xxx",
		},
		err: `parse "http://foo.com/%xxx": invalid URL escape "%xx"`,
	})
	tests.Add("env only", tt{
		env: map[string]string{
			"KIVIKSN": "http://foo.com/",
		},
	})
	tests.Add("invalid yaml", tt{
		filename: "testdata/invalid.yaml",
		err:      `yaml: found unexpected end of stream`,
	})
	tests.Add("valid yaml", tt{
		filename: "testdata/valid.yaml",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		testEnv(t, tt.env)
		l := log.NewTest()
		cf := New(nil)
		err := cf.Read(tt.filename, l)
		testy.Error(t, tt.err, err)
		if d := testy.DiffInterface(testy.Snapshot(t), cf); d != nil {
			t.Error(d)
		}
		l.Check(t)
	})
}

func testEnv(t *testing.T, env map[string]string) {
	t.Helper()
	t.Cleanup(testy.RestoreEnv())
	os.Clearenv()
	if err := testy.SetEnv(env); err != nil {
		t.Fatal(err)
	}
}

func TestConfig_DSN(t *testing.T) {
	type tt struct {
		cf   *Config
		want string
		err  string
	}

	tests := testy.NewTable()
	tests.Add("no current context", tt{
		cf:  &Config{},
		err: "no context specified",
	})
	tests.Add("context not found", tt{
		cf:  &Config{CurrentContext: "xxx"},
		err: `context "xxx" not found`,
	})
	tests.Add("only one context, no default", tt{
		cf: &Config{
			Contexts: map[string]*Context{
				"foo": {
					Scheme:   "http",
					Host:     "localhost:5984",
					User:     "admin",
					Password: "abc123",
					Database: "_users",
				},
			},
		},
		want: "http://admin:abc123@localhost:5984/_users",
	})
	tests.Add("success", tt{
		cf: &Config{
			Contexts: map[string]*Context{
				"foo": {
					Scheme:   "http",
					Host:     "localhost:5984",
					User:     "admin",
					Password: "abc123",
					Database: "_users",
				},
			},
			CurrentContext: "foo",
		},
		want: "http://admin:abc123@localhost:5984/_users",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		got, err := tt.cf.DSN()
		testy.Error(t, tt.err, err)
		if got != tt.want {
			t.Errorf("Unexpected result: %s", got)
		}
	})
}

func TestConfigArgs(t *testing.T) {
	type tt struct {
		c    *Config
		args []string
		err  string
	}

	tests := testy.NewTable()
	tests.Add("no arguments", tt{
		c: &Config{},
	})
	tests.Add("invalid dsn", tt{
		c:    &Config{},
		args: []string{"http://localhost:5984/%xxx"},
		err:  `parse "http://localhost:5984/%xxx": invalid URL escape "%xx"`,
	})
	tests.Add("full dsn in args", tt{
		c:    &Config{},
		args: []string{"http://localhost:5984/foo/bar"},
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		lg := log.NewTest()
		c := tt.c
		c.log = lg
		if c.Contexts == nil {
			c.Contexts = make(map[string]*Context)
		}
		cmd := &cobra.Command{}
		err := tt.c.Args(cmd, tt.args)
		testy.Error(t, tt.err, err)
		if d := testy.DiffInterface(testy.Snapshot(t), c); d != nil {
			t.Error(d)
		}
		lg.Check(t)
	})
}

func TestConfig_SetURL(t *testing.T) {
	type tt struct {
		wd  string
		cf  *Config
		url string
		err string
	}

	tests := testy.NewTable()
	tests.Add("empty url", tt{
		cf:  New(nil),
		url: "",
	})
	tests.Add("full dsn, empty config", tt{
		cf:  New(nil),
		url: "http://admin:abc123@localhost:5984/foo/bar",
	})
	tests.Add("db/doc, empty config", tt{
		cf:  New(nil),
		url: "foo/bar",
	})
	tests.Add("doc only, empty config", tt{
		cf:  New(nil),
		url: "bar",
	})
	tests.Add("db/doc, with config", tt{
		cf: &Config{
			Contexts: map[string]*Context{
				"foo": {
					Scheme:   "http",
					Host:     "localhost:5984",
					User:     "admin",
					Password: "abc123",
					Database: "_users",
				},
			},
			CurrentContext: "foo",
		},
		url: "foo/bar",
	})
	tests.Add("doc, with config", tt{
		cf: &Config{
			Contexts: map[string]*Context{
				"foo": {
					Scheme:   "http",
					Host:     "localhost:5984",
					User:     "admin",
					Password: "abc123",
					Database: "_users",
				},
			},
			CurrentContext: "foo",
		},
		url: "bar",
	})
	tests.Add("absolute path, empty config", tt{
		cf:  New(nil),
		url: "/some/path",
	})
	tests.Add("relative path, empty config", tt{
		wd:  "/tmp",
		cf:  New(nil),
		url: "./some/path",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		var cwd string
		if tt.wd != "" {
			var err error
			cwd, err = os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Chdir(tt.wd); err != nil {
				t.Fatal(err)
			}
		}
		tl := log.NewTest()
		tt.cf.log = tl
		opts, err := tt.cf.SetURL(tt.url)
		if cwd != "" {
			_ = os.Chdir(cwd)
		}
		testy.Error(t, tt.err, err)
		tl.Check(t)
		tt.cf.log = nil
		if d := testy.DiffInterface(testy.Snapshot(t), tt.cf); d != nil {
			t.Error(d)
		}
		if d := testy.DiffInterface(testy.Snapshot(t, "opts"), opts); d != nil {
			t.Errorf("options: %s", d)
		}
	})
}

func Test_expandDSN(t *testing.T) {
	type tt struct {
		input                  string
		dsn, db, doc, filename string
	}

	tests := testy.NewTable()
	tests.Add("db", tt{
		input: "db",
		db:    "db",
	})
	tests.Add("db/doc", tt{
		input: "db/doc",
		db:    "db",
		doc:   "doc",
	})
	tests.Add("db/doc/filename", tt{
		input:    "db/doc/filename.txt",
		db:       "db",
		doc:      "doc",
		filename: "filename.txt",
	})
	tests.Add("full url", tt{
		input:    "http://foo.com/db/doc/filename.txt",
		dsn:      "http://foo.com/",
		db:       "db",
		doc:      "doc",
		filename: "filename.txt",
	})
	tests.Add("subdir-hosted couchdb, full url", tt{
		input:    "http://foo.com/couchdb/db/doc/filename.txt",
		dsn:      "http://foo.com/couchdb/",
		db:       "db",
		doc:      "doc",
		filename: "filename.txt",
	})
	tests.Add("subdir-hosted couchdb, db only", tt{
		input: "http://foo.com/couchdb//db",
		dsn:   "http://foo.com/couchdb/",
		db:    "db",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		addr, err := url.Parse(tt.input)
		if err != nil {
			t.Fatal(err)
		}
		dsn, db, doc, filename := expandDSN(addr)
		if dsn != tt.dsn || db != tt.db || doc != tt.doc || filename != tt.filename {
			t.Errorf("Unexpected output: dsn:%s, db:%s, doc:%s, filename:%s",
				dsn, db, doc, filename)
		}
	})
}
