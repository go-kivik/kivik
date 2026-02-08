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

package kt

import (
	"context"
	"fmt"
	"testing"

	kivik "github.com/go-kivik/kivik/v4"
)

// Context holds the test suite's client connections and configuration.
type Context struct {
	// RW is true if we should run read-write tests.
	RW bool
	// Admin is a client connection with database admin privileges.
	Admin *kivik.Client
	// NoAuth is a client connection with no authentication.
	NoAuth *kivik.Client
	// Config is the suite config.
	Config SuiteConfig
}

// Skip will skip the currently running test if configuration dictates.
func (c *Context) Skip(t *testing.T) {
	t.Helper()
	if c.Config.Bool(t, "skip") {
		t.Skip("Test skipped by suite configuration")
	}
}

// MustBeSet ends the test with a failure if the config key is not set.
func (c *Context) MustBeSet(t *testing.T, key string) {
	t.Helper()
	if !c.IsSet(t, key) {
		t.Fatalf("'%s' not set. Please configure this test.", key)
	}
}

// MustStringSlice returns a string slice, or fails if the value is unset.
func (c *Context) MustStringSlice(t *testing.T, key string) []string {
	t.Helper()
	c.MustBeSet(t, key)
	return c.StringSlice(t, key)
}

// MustBool returns a bool, or fails if the value is unset.
func (c *Context) MustBool(t *testing.T, key string) bool {
	t.Helper()
	c.MustBeSet(t, key)
	return c.Bool(t, key)
}

// IntSlice returns a []int from config.
func (c *Context) IntSlice(t *testing.T, key string) []int {
	t.Helper()
	v, _ := c.Config.Interface(t, key).([]int)
	return v
}

// MustIntSlice returns a []int, or fails if the value is unset.
func (c *Context) MustIntSlice(t *testing.T, key string) []int {
	t.Helper()
	c.MustBeSet(t, key)
	return c.IntSlice(t, key)
}

// StringSlice returns a string slice from the config.
func (c *Context) StringSlice(t *testing.T, key string) []string {
	t.Helper()
	return c.Config.StringSlice(t, key)
}

// String returns a string from config.
func (c *Context) String(t *testing.T, key string) string {
	t.Helper()
	return c.Config.String(t, key)
}

// MustString returns a string from config, or fails if the value is unset.
func (c *Context) MustString(t *testing.T, key string) string {
	t.Helper()
	c.MustBeSet(t, key)
	return c.String(t, key)
}

// Int returns an int from the config.
func (c *Context) Int(t *testing.T, key string) int {
	t.Helper()
	return c.Config.Int(t, key)
}

// MustInt returns an int from the config, or fails if the value is unset.
func (c *Context) MustInt(t *testing.T, key string) int {
	t.Helper()
	c.MustBeSet(t, key)
	return c.Int(t, key)
}

// Bool returns a bool from the config.
func (c *Context) Bool(t *testing.T, key string) bool {
	t.Helper()
	return c.Config.Bool(t, key)
}

// Interface returns the configuration value as an any.
func (c *Context) Interface(t *testing.T, key string) any {
	t.Helper()
	return c.Config.get(name(t), key)
}

// Options returns an options map value.
func (c *Context) Options(t *testing.T, key string) kivik.Option {
	t.Helper()
	testName := name(t)
	i := c.Config.get(testName, key)
	if i == nil {
		return nil
	}
	o, ok := i.(kivik.Option)
	if !ok {
		panic(fmt.Sprintf("Options field %s/%s of unsupported type: %T", testName, key, i))
	}
	return o
}

// MustInterface returns an any from the config, or fails if the value is unset.
func (c *Context) MustInterface(t *testing.T, key string) any {
	t.Helper()
	c.MustBeSet(t, key)
	return c.Interface(t, key)
}

// IsSet returns true if the value is set in the configuration.
func (c *Context) IsSet(t *testing.T, key string) bool {
	t.Helper()
	return c.Interface(t, key) != nil
}

// Run wraps t.Run().
func (c *Context) Run(t *testing.T, name string, fn func(*testing.T)) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		t.Helper()
		c.Skip(t)
		fn(t)
	})
}

// RunAdmin runs the test function iff c.Admin is set.
func (c *Context) RunAdmin(t *testing.T, fn func(*testing.T)) {
	t.Helper()
	if c.Admin != nil {
		c.Run(t, "Admin", fn)
	}
}

// RunNoAuth runs the test function iff c.NoAuth is set.
func (c *Context) RunNoAuth(t *testing.T, fn func(*testing.T)) {
	t.Helper()
	if c.NoAuth != nil {
		c.Run(t, "NoAuth", fn)
	}
}

// RunRW runs the test function iff c.RW is true.
func (c *Context) RunRW(t *testing.T, fn func(*testing.T)) {
	t.Helper()
	if c.RW {
		c.Run(t, "RW", fn)
	}
}

// RunRO runs the test function iff c.RW is false. Note that unlike RunRW, this
// does not start a new subtest. This should usually be run in conjunction with
// RunRW, to run only RO or RW tests, in situations where running both would be
// redundant.
func (c *Context) RunRO(t *testing.T, fn func(*testing.T)) {
	t.Helper()
	if !c.RW {
		fn(t)
	}
}

// TestDB creates a test database, registers a cleanup function to destroy it,
// and returns its name.
func (c *Context) TestDB(t *testing.T) string {
	t.Helper()
	dbname := TestDBName(t)
	err := Retry(func() error {
		return c.Admin.CreateDB(context.Background(), dbname, c.Options(t, "db"))
	})
	if err != nil {
		t.Fatalf("Failed to create database %q: %s", dbname, err)
	}
	t.Cleanup(func() { c.DestroyDB(t, dbname) })
	return dbname
}

// DestroyDB cleans up the specified DB after tests run.
func (c *Context) DestroyDB(t *testing.T, name string) {
	t.Helper()
	Retry(func() error { //nolint:errcheck
		return c.Admin.DestroyDB(context.Background(), name, c.Options(t, "db"))
	})
}
