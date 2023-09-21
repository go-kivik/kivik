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

package fs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-kivik/fsdb/v4/cdb"
	"github.com/go-kivik/fsdb/v4/filesystem"
	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

const dirMode = os.FileMode(0o700)

type fsDriver struct {
	fs filesystem.Filesystem
}

var _ driver.Driver = &fsDriver{}

// Identifying constants
const (
	Version = "0.0.1"
	Vendor  = "Kivik File System Adaptor"
)

func init() {
	kivik.Register("fs", &fsDriver{})
}

type client struct {
	version *driver.Version
	root    string
	fs      filesystem.Filesystem
}

var _ driver.Client = &client{}

func parseFileURL(dir string) (string, error) {
	parsed, err := url.Parse(dir)
	if parsed.Scheme != "" && parsed.Scheme != "file" {
		return "", statusError{status: http.StatusBadRequest, error: fmt.Errorf("Unsupported URL scheme '%s'. Wrong driver?", parsed.Scheme)}
	}
	if !strings.HasPrefix(dir, "file://") {
		return dir, nil
	}
	if err != nil {
		return "", err
	}
	return parsed.Path, nil
}

func (d *fsDriver) NewClient(dir string, _ driver.Options) (driver.Client, error) {
	path, err := parseFileURL(dir)
	if err != nil {
		return nil, err
	}
	fs := d.fs
	if fs == nil {
		fs = filesystem.Default()
	}
	return &client{
		version: &driver.Version{
			Version:     Version,
			Vendor:      Vendor,
			RawResponse: json.RawMessage(fmt.Sprintf(`{"version":"%s","vendor":{"name":"%s"}}`, Version, Vendor)),
		},
		fs:   fs,
		root: path,
	}, nil
}

// Version returns the configured server info.
func (c *client) Version(_ context.Context) (*driver.Version, error) {
	return c.version, nil
}

// Taken verbatim from http://docs.couchdb.org/en/2.0.0/api/database/common.html
var validDBNameRE = regexp.MustCompile("^[a-z_][a-z0-9_$()+/-]*$")

// AllDBs returns a list of all DBs present in the configured root dir.
func (c *client) AllDBs(context.Context, driver.Options) ([]string, error) {
	if c.root == "" {
		return nil, statusError{status: http.StatusBadRequest, error: errors.New("no root path provided")}
	}
	files, err := os.ReadDir(c.root)
	if err != nil {
		return nil, err
	}
	filenames := make([]string, 0, len(files))
	for _, file := range files {
		dbname := cdb.UnescapeID(file.Name())
		if !validDBNameRE.MatchString(dbname) {
			// FIXME #64: Add option to warn about non-matching files?
			continue
		}
		filenames = append(filenames, cdb.EscapeID(file.Name()))
	}
	return filenames, nil
}

// CreateDB creates a database
func (c *client) CreateDB(ctx context.Context, dbName string, options driver.Options) error {
	exists, err := c.DBExists(ctx, dbName, options)
	if err != nil {
		return err
	}
	if exists {
		return statusError{status: http.StatusPreconditionFailed, error: errors.New("database already exists")}
	}
	return os.Mkdir(filepath.Join(c.root, cdb.EscapeID(dbName)), dirMode)
}

// DBExistsreturns true if the database exists.
func (c *client) DBExists(_ context.Context, dbName string, _ driver.Options) (bool, error) {
	_, err := os.Stat(filepath.Join(c.root, cdb.EscapeID(dbName)))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// DestroyDB destroys the database
func (c *client) DestroyDB(ctx context.Context, dbName string, options driver.Options) error {
	exists, err := c.DBExists(ctx, dbName, options)
	if err != nil {
		return err
	}
	if !exists {
		return statusError{status: http.StatusNotFound, error: errors.New("database does not exist")}
	}
	// FIXME #65: Be safer here about unrecognized files
	return os.RemoveAll(filepath.Join(c.root, cdb.EscapeID(dbName)))
}

func (c *client) DB(dbName string, _ driver.Options) (driver.DB, error) {
	return c.newDB(dbName)
}

// dbPath returns the full DB path and the dbname.
func (c *client) dbPath(path string) (string, string, error) {
	// As a special case, skip validation on this one
	if c.root == "" && path == "." {
		return ".", ".", nil
	}
	dbname := path
	if c.root == "" {
		if strings.HasPrefix(path, "file://") {
			addr, err := url.Parse(path)
			if err != nil {
				return "", "", statusError{status: http.StatusBadRequest, error: err}
			}
			path = addr.Path
		}
		if strings.Contains(dbname, "/") {
			dbname = dbname[strings.LastIndex(dbname, "/")+1:]
		}
	} else {
		path = filepath.Join(c.root, dbname)
	}
	return path, dbname, nil
}

func (c *client) newDB(dbname string) (*db, error) {
	path, name, err := c.dbPath(dbname)
	if err != nil {
		return nil, err
	}
	return &db{
		client: c,
		dbPath: path,
		dbName: name,
		fs:     c.fs,
		cdb:    cdb.New(path, c.fs),
	}, nil
}
