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
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/couchdb"
	_ "github.com/go-kivik/kivik/v4/x/fsdb" // Filesystem driver

	"github.com/go-kivik/xkivik/v4/cmd/kivik/errors"
	"github.com/go-kivik/xkivik/v4/cmd/kivik/log"
)

const envPrefix = "KIVIK"

// Config is the full app configuration file.
type Config struct {
	Contexts       map[string]*Context `yaml:"contexts"`
	CurrentContext string              `yaml:"current-context"`
	log            log.Logger
	finalizer      func()
}

// Context represents a complete, or partial CouchDB DSN context.
type Context struct {
	Scheme   string `yaml:"scheme"`
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	DocID    string `yaml:"-"`
}

func (c *Context) String() string {
	return c.DSN()
}

func (c *Context) dsn() *url.URL {
	var user *url.Userinfo
	if c.User != "" || c.Password != "" {
		user = url.UserPassword(c.User, c.Password)
	}
	return &url.URL{
		Scheme: c.Scheme,
		Host:   c.Host,
		Path:   path.Join(c.Database, c.DocID),
		User:   user,
	}
}

func (c *Context) DSN() string {
	return c.dsn().String()
}

// ServerDSN returns just the server DSN, with no database or docid.
func (c *Context) ServerDSN() string {
	dsn := c.dsn()
	dsn.Path = ""
	return dsn.String()
}

func (c *Context) DB() (db string, err error) {
	_, db, doc, _ := expandDSN(c.dsn())
	if doc != "" {
		return "", errors.Code(errors.ErrUsage, "DSN expected to contain only the database")
	}
	return db, nil
}

func (c *Context) DBDoc() (db, doc string, err error) {
	addr := c.dsn()
	p := addr.Path
	addr.Path = ""
	if addr.String() == "" {
		return "", "", errors.Code(errors.ErrUsage, "document ID required")
	}
	if p != "" {
		db = strings.Trim(path.Dir(p), "/")
		doc = strings.Trim(path.Base(p), "/")
	}
	if strings.HasSuffix(db, "/_design") {
		db = strings.TrimSuffix(db, "/_design")
		doc = "_design/" + doc
	}
	if db == "" {
		db = doc
		doc = ""
	}
	return db, doc, nil
}

func (c *Context) DBDocFilename() (db, doc, filename string, err error) {
	_, db, doc, filename = expandDSN(c.dsn())
	if filename == "" {
		return "", "", "", errors.Code(errors.ErrUsage, "attachment filename required")
	}
	return db, doc, filename, nil
}

// UnmarshalYAML handles parsing of a Context from YAML input.
func (c *Context) UnmarshalYAML(v *yaml.Node) error {
	dsn := struct {
		DSN string `yaml:"dsn"`
	}{}
	if err := v.Decode(&dsn); err != nil {
		return err
	}
	if dsn.DSN == "" {
		type alias Context
		intl := alias{}
		err := v.Decode(&intl)
		*c = Context(intl)
		return err
	}
	uri, err := url.Parse(dsn.DSN)
	if err != nil {
		return err
	}
	var user, password string
	if u := uri.User; u != nil {
		user = u.Username()
		password, _ = u.Password()
	}
	*c = Context{
		Scheme:   uri.Scheme,
		Host:     uri.Host,
		User:     user,
		Password: password,
		Database: uri.Path,
	}
	return nil
}

// New returns an empty configuration object. Call Read() to populate it.
func New(finalizer func()) *Config {
	return &Config{
		Contexts:  make(map[string]*Context),
		finalizer: finalizer,
	}
}

// Read populates c with app configuration found in filename.
//
// - Reads from filename
// - If DSN env variable is set, it's added as context called 'ENV' and made current
func (c *Config) Read(filename string, lg log.Logger) error {
	c.log = lg
	if err := c.readYAML(filename); err != nil {
		return errors.WithCode(err, errors.ErrUsage)
	}
	if dsn := os.Getenv(envPrefix + "DSN"); dsn != "" {
		if err := c.setDefaultDSN(dsn); err != nil {
			return err
		}
		lg.Debug("set default DSN from environment")
	}
	return nil
}

func (c *Config) readYAML(filename string) error {
	if filename == "" {
		c.log.Debug("no config file specified")
		return nil
	}
	f, err := os.Open(filename)
	if err != nil {
		c.log.Debugf("failed to read config: %s", err)
		if os.IsNotExist(err) {
			err = nil
		}
		return err
	}
	if err := yaml.NewDecoder(f).Decode(c); err != nil {
		c.log.Debugf("YAML parse error: %s", err)
		return err
	}
	c.log.Debugf("successfully read config file %q", filename)
	return nil
}

// CurrentCx returns the current context.
func (c *Config) CurrentCx() (*Context, error) {
	if c.CurrentContext == "" {
		if len(c.Contexts) == 1 {
			for _, cx := range c.Contexts {
				return cx, nil
			}
		}
		return nil, errors.Code(errors.ErrUsage, "no context specified")
	}
	cx, ok := c.Contexts[c.CurrentContext]
	if !ok {
		return nil, errors.Codef(errors.ErrUsage, "context %q not found", c.CurrentContext)
	}
	return cx, nil
}

// KivikClient returns a connection to Kivik.
func (c *Context) KivikClient(connTimeout, reqTimeout time.Duration) (*kivik.Client, error) {
	scheme, dsn, err := c.ClientInfo()
	if err != nil {
		return nil, err
	}

	switch scheme {
	case "file":
		return kivik.New("fs", dsn)
	case "http", "https", "couch", "couchs", "couchdb", "couchdbs":
		return kivik.New("couch", dsn, couchdb.OptionHTTPClient(&http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: connTimeout,
				}).DialContext,
			},
			Timeout: reqTimeout,
		}))
	}
	return nil, errors.Codef(errors.ErrUsage, "unsupported URL scheme: %s", scheme)
}

// ClientInfo returns the URL scheme, and DSN of the context.
func (c *Context) ClientInfo() (string, string, error) {
	dsn := c.ServerDSN()
	if dsn == "" {
		return "", "", errors.Code(errors.ErrUsage, "server hostname required")
	}
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	return scheme, dsn, nil
}

// ClientInfo returns the URL scheme, and DSN, for use by the root command to
// establish the kivik client connection.
func (c *Config) ClientInfo() (string, string, error) {
	cx, err := c.CurrentCx()
	if err != nil {
		return "", "", err
	}
	return cx.ClientInfo()
}

func (c *Config) URL() (*url.URL, error) {
	cx, err := c.CurrentCx()
	if err != nil {
		return nil, err
	}
	return cx.dsn(), nil
}

func (c *Config) DSN() (string, error) {
	cx, err := c.CurrentCx()
	if err != nil {
		return "", err
	}
	c.finalize()
	return cx.DSN(), nil
}

func (c *Config) finalize() {
	if c.finalizer != nil {
		c.finalizer()
	}
}

func (c *Config) Finalize() {
	c.finalize()
}

func (c *Config) ServerDSN() (string, error) {
	cx, err := c.CurrentCx()
	if err != nil {
		return "", err
	}
	dsn := cx.ServerDSN()
	if dsn == "" {
		return "", errors.Code(errors.ErrUsage, "server hostname required")
	}
	c.finalize()
	return dsn, nil
}

func (c *Config) HasAttachment() bool {
	cx, err := c.CurrentCx()
	if err != nil {
		return false
	}
	c.finalize()
	_, _, filename, err := cx.DBDocFilename()
	return err == nil && filename != ""
}

func (c *Config) HasDoc() bool {
	cx, err := c.CurrentCx()
	if err != nil {
		return false
	}
	c.finalize()
	_, doc, err := cx.DBDoc()
	return err == nil && doc != ""
}

func (c *Config) HasDB() bool {
	cx, err := c.CurrentCx()
	if err != nil {
		return false
	}
	c.finalize()
	db, _, err := cx.DBDoc()
	return err == nil && db != ""
}

func (c *Config) DB() (db string, err error) {
	cx, err := c.CurrentCx()
	if err != nil {
		return "", err
	}
	db, err = cx.DB()
	if err != nil {
		return "", err
	}
	c.finalize()
	return db, nil
}

func (c *Config) DBDoc() (db, doc string, err error) {
	cx, err := c.CurrentCx()
	if err != nil {
		return "", "", err
	}
	c.finalize()
	return cx.DBDoc()
}

func (c *Config) DBDocFilename() (db, doc, filename string, err error) {
	cx, err := c.CurrentCx()
	if err != nil {
		return "", "", "", err
	}
	c.finalize()
	return cx.DBDocFilename()
}

// Config sets config from the cobra command.
func (c *Config) Args(_ *cobra.Command, args []string) error {
	if len(args) > 0 {
		if err := c.setDefaultDSN(args[0]); err != nil {
			return err
		}
		c.log.Debug("set default DSN from command line arguments")
	}
	return nil
}

// setDefaultDSN sets the default DSN. It's meant to be used when setting from
// the environment, or CLI.
func (c *Config) setDefaultDSN(dsn string) error {
	cx, _, err := ContextFromDSN(dsn)
	if err != nil {
		return err
	}
	c.Contexts["*"] = cx
	c.CurrentContext = "*"
	return nil
}

// ContextFromDSN parses a DSN into a context object, and a map of kivik
// parameters read from the url query parameters.
func ContextFromDSN(dsn string) (*Context, map[string]string, error) {
	if len(dsn) > 0 {
		switch dsn[0] {
		case '/':
			dsn = "file://" + dsn
		case '.':
			cwd, err := os.Getwd()
			if err != nil {
				return nil, nil, err
			}
			dsn = "file://" + filepath.Join(cwd, dsn[1:])
		}
	}
	uri, err := url.Parse(dsn)
	if err != nil {
		return nil, nil, errors.WithCode(err, errors.ErrUsage)
	}
	var user, password string
	if u := uri.User; u != nil {
		user = u.Username()
		password, _ = u.Password()
	}
	db := uri.Path
	var docid string
	if strings.Contains(db, "/") {
		docid = path.Base(db)
		db = path.Dir(db)
	}
	// If we have no hostname, and no docid, that means we really got only a
	// docid, so we need to adjust...
	if uri.Host == "" && docid == "" {
		docid = db
		db = ""
	}
	return &Context{
		Scheme:   uri.Scheme,
		Host:     uri.Host,
		User:     user,
		Password: password,
		Database: db,
		DocID:    docid,
	}, query2options(uri.Query()), nil
}

// SetURL sets the current context based on a URL argument passed on the
// command line.
//
// Supported formats and examples:
//
// - Full DSN           -- http://localhost:5984/database/docid
// - Path only          -- database/docid
// - Doc ID only        -- docid
// - Absolute file path -- /absolute/path
// - Relative file path -- ./relative/path
func (c *Config) SetURL(dsn string) (map[string]string, error) {
	if dsn == "" {
		return nil, nil
	}
	cx, opts, err := ContextFromDSN(dsn)
	if err != nil {
		return nil, err
	}
	curCx, _ := c.CurrentCx()
	if cx.Host == "" && curCx != nil {
		c.log.Debugf("Incomplete DSN provided: %q, merging with current context: %q", dsn, curCx)
		cx.Scheme = curCx.Scheme
		cx.Host = curCx.Host
		cx.User = curCx.User
		cx.Password = curCx.Password
		if cx.Database == "" {
			cx.Database = curCx.Database
		}
	}
	c.Contexts["*"] = cx
	c.CurrentContext = "*"
	return opts, nil
}

func query2options(q url.Values) map[string]string {
	opts := make(map[string]string, len(q))
	for k := range q {
		opts[k] = q.Get(k)
	}
	return opts
}

func expandDSN(addr *url.URL) (dsn, db, doc, filename string) {
	parts := splitPath(addr.Path)
	dsn, db, doc, filename = parts[0], parts[1], parts[2], parts[3]
	if addr.Host != "" {
		addr.Path = strings.Trim(dsn, "/") + "/"
		dsn = addr.String()
	}
	return dsn, db, doc, filename
}

// splitPath splits p into a maximum of 4 elements, representing the couchdb
// root, database name, doc id, and attachment filename. A double slash (//) can
// be used to specify the end of the root element, when disambiguation is
// required.
func splitPath(p string) []string {
	if parts := strings.SplitN(p, "//", 2); len(parts) == 2 { // nolint:gomnd
		result := splitPath(parts[1])
		result[0] = parts[0] + "/" + result[0]
		return result
	}
	const maxElements = 4
	result := make([]string, 0, maxElements-1)
	for len(result) < maxElements-1 {
		base := path.Base(p)
		p = path.Dir(p)
		result = append([]string{base}, result...)
		if p == "." || p == "/" {
			p = ""
			break
		}
	}
	result = append([]string{p}, result...)
	return append(result, []string{"", "", ""}...)
}
