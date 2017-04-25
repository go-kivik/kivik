// Package fs provides a filesystem-backed Kivik driver.
package fs

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/common"
	"github.com/flimzy/kivik/errors"
)

const dirMode = os.FileMode(0700)
const fileMode = os.FileMode(0600)

type fsDriver struct{}

var _ driver.Driver = &fsDriver{}

// Identifying constants
const (
	Version = "0.0.1"
	Vendor  = "Kivik Memory Adaptor"
)

func init() {
	kivik.Register("fs", &fsDriver{})
}

func (d *fsDriver) NewClient(_ context.Context, dir string) (driver.Client, error) {
	if err := validateRootDir(dir); err != nil {
		if os.IsPermission(errors.Cause(err)) {
			return nil, errors.Status(kivik.StatusUnauthorized, "access denied")
		}
		return nil, err
	}
	return &client{
		Client: common.NewClient(Version, Vendor),
		root:   dir,
	}, nil
}

func validateRootDir(dir string) error {
	// See if the target path exists, and is a directory
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, dirMode); err != nil {
			return errors.Wrapf(err, "failed to create dir '%s'", dir)
		}
		if _, err = os.Create(dir + "/.kivik"); err != nil {
			return errors.Wrapf(err, "failed to create file '%s/.kivik'", dir)
		}
		if err = os.Chmod(dir+"/.kivik", fileMode); err != nil {
			return errors.Wrapf(err, "failed to set mode %s on '%s/.kivik'", fileMode, dir)
		}
		return nil
	}
	// Ensure this is a .kivik directory
	if _, err := os.Stat(dir + "/.kivik"); os.IsNotExist(err) {
		return fmt.Errorf("kivik: '%s' is not a kivik data store (.kivik file missing)", dir)
	}
	// Ensure we have write access
	tmpF, err := ioutil.TempFile(dir, ".kivik-test")
	if err != nil {
		return errors.Wrapf(err, "failed to write to '%s'", dir)
	}
	_ = os.Remove(tmpF.Name())
	return nil
}

type client struct {
	*common.Client
	root string
}

var _ driver.Client = &client{}

// Taken verbatim from http://docs.couchdb.org/en/2.0.0/api/database/common.html
var validDBNameRE = regexp.MustCompile("^[a-z][a-z0-9_$()+/-]*$")

// AllDBs returns a list of all DBs present in the configured root dir.
func (c *client) AllDBs(_ context.Context, _ map[string]interface{}) ([]string, error) {
	files, err := ioutil.ReadDir(c.root)
	if err != nil {
		return nil, errors.WrapStatus(kivik.StatusInternalServerError, err)
	}
	filenames := make([]string, 0, len(files))
	for _, file := range files {
		if file.Name()[0] == '.' {
			// As a special case, we skip over dot files
			continue
		}
		if !validDBNameRE.MatchString(file.Name()) {
			// Warn about bad filenames
			fmt.Printf("kivik: Filename does not conform to database name standards: %s/%s\n", c.root, file.Name())
			continue
		}
		filenames = append(filenames, file.Name())
	}
	return filenames, nil
}

// CreateDB creates a database
func (c *client) CreateDB(ctx context.Context, dbName string, options map[string]interface{}) error {
	exists, err := c.DBExists(ctx, dbName, options)
	if err != nil {
		return err
	}
	if exists {
		return errors.Status(kivik.StatusPreconditionFailed, "database already exists")
	}
	if err := os.Mkdir(c.root+"/"+dbName, dirMode); err != nil {
		return errors.WrapStatus(kivik.StatusInternalServerError, err)
	}
	return nil
}

// DBExistsreturns true if the database exists.
func (c *client) DBExists(_ context.Context, dbName string, _ map[string]interface{}) (bool, error) {
	_, err := os.Stat(c.root + "/" + dbName)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, errors.WrapStatus(kivik.StatusInternalServerError, err)
}

// DestroyDB destroys the database
func (c *client) DestroyDB(ctx context.Context, dbName string, options map[string]interface{}) error {
	exists, err := c.DBExists(ctx, dbName, options)
	if err != nil {
		return err
	}
	if !exists {
		return errors.Status(kivik.StatusNotFound, "database does not exist")
	}
	if err = os.RemoveAll(c.root + "/" + dbName); err != nil {
		return errors.WrapStatus(kivik.StatusInternalServerError, err)
	}
	return nil
}

func (c *client) DB(_ context.Context, dbName string, options map[string]interface{}) (driver.DB, error) {
	return &db{
		client: c,
		dbName: dbName,
	}, nil
}
