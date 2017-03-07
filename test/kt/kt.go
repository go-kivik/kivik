// Package kt provides common utilities for Kivik tests.
package kt

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/flimzy/kivik"
)

// Context is a collection of client connections with different security access.
type Context struct {
	// RW is true if we should run read-write tests.
	RW bool
	// Admin is a client connection with database admin priveleges.
	Admin *kivik.Client
	// NoAuth isa client connection with no authentication.
	NoAuth *kivik.Client
	// Config is the suite config
	Config SuiteConfig
	// T is the *testing.T value
	T *testing.T
}

// Skip will skip the currently running test if configuration dictates.
func (c *Context) Skip() {
	if c.Config.Bool(c.T, "skip") {
		c.T.Skip("Test skipped by suite configuration")
	}
}

// Skipf is a wrapper around t.Skipf()
func (c *Context) Skipf(format string, args ...interface{}) {
	c.T.Skipf(format, args...)
}

// Logf is a wrapper around t.Logf()
func (c *Context) Logf(format string, args ...interface{}) {
	c.T.Logf(format, args...)
}

//Fatalf is a wrapper around t.Fatalf()
func (c *Context) Fatalf(format string, args ...interface{}) {
	c.T.Fatalf(format, args...)
}

// MustBeSet ends the test with a failure if the config key is not set.
func (c *Context) MustBeSet(key string) {
	if !c.IsSet(key) {
		c.T.Fatalf("'%s' not set. Please configure this test.", key)
	}
}

// MustStringSlice returns a string slice, or fails if the value is unset.
func (c *Context) MustStringSlice(key string) []string {
	c.MustBeSet(key)
	return c.StringSlice(key)
}

// MustBool returns a bool, or fails if the value is unset.
func (c *Context) MustBool(key string) bool {
	c.MustBeSet(key)
	return c.Bool(key)
}

// IntSlice returns a []int from config.
func (c *Context) IntSlice(key string) []int {
	v, _ := c.Config.Interface(c.T, key).([]int)
	return v
}

// MustIntSlice returns a []int, or fails if the value is unset.
func (c *Context) MustIntSlice(key string) []int {
	c.MustBeSet(key)
	return c.IntSlice(key)
}

// StringSlice returns a string slice from the config.
func (c *Context) StringSlice(key string) []string {
	return c.Config.StringSlice(c.T, key)
}

func (c *Context) String(key string) string {
	return c.Config.String(c.T, key)
}

// Int returns an int from the config.
func (c *Context) Int(key string) int {
	return c.Config.Int(c.T, key)
}

// MustInt returns an int from the config, or fails if the value is unset.
func (c *Context) MustInt(key string) int {
	c.MustBeSet(key)
	return c.Int(key)
}

// Bool returns a bool from the config.
func (c *Context) Bool(key string) bool {
	return c.Config.Bool(c.T, key)
}

// IsSet returns true if the value is set in the configuration.
func (c *Context) IsSet(key string) bool {
	return c.Config.Interface(c.T, key) != nil
}

// Run wraps t.Run()
func (c *Context) Run(name string, fn testFunc) {
	c.T.Run(name, func(t *testing.T) {
		ctx := &Context{
			RW:     c.RW,
			Admin:  c.Admin,
			NoAuth: c.NoAuth,
			Config: c.Config,
			T:      t,
		}
		ctx.Skip()
		fn(ctx)
	})
}

type testFunc func(*Context)

// tests is a map of the format map[suite]map[name]testFunc
var tests = make(map[string]testFunc)

// Register registers a test to be run for the given test suite. rw should
// be true if the test writes to the database.
func Register(name string, fn testFunc) {
	tests[name] = fn
	return
}

// RunSubtests executes the requested suites of tests against the client.
func RunSubtests(ctx *Context) {
	for name, fn := range tests {
		ctx.Run(name, fn)
	}
}

var rnd *rand.Rand

func init() {
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// TestDBPrefix is used to prefix temporary database names during tests.
const TestDBPrefix = "kivik$"

// Go 1.8+ supports this interface on the *testing.T type, but we need to check
// for the sake of earlier versions.
type namer interface {
	Name() string
}

// TestDBName generates a randomized string suitable for a database name for
// testing.
func (c *Context) TestDBName() string {
	var id string

	// All this non-sense to support Go < 1.8, which doesn't support t.Name()
	var ti interface{} = c.T
	if n, ok := ti.(namer); ok {
		id = strings.ToLower(n.Name())
		id = id[strings.Index(id, "/")+1:]
		id = strings.Replace(id, "/", "_", -1) + "$"
	}
	return fmt.Sprintf("%s%s%016x", TestDBPrefix, id, rnd.Int63())
}

// RunAdmin runs the test function iff c.Admin is set.
func (c *Context) RunAdmin(fn testFunc) {
	if c.Admin != nil {
		c.Run("Admin", fn)
	}
}

// RunNoAuth runs the test function iff c.NoAuth is set.
func (c *Context) RunNoAuth(fn testFunc) {
	if c.NoAuth != nil {
		c.Run("NoAuth", fn)
	}
}

// RunRW runs the test function iff c.RW is true.
func (c *Context) RunRW(fn testFunc) {
	if c.RW {
		c.Run("RW", fn)
	}
}

// RunRO runs the test function iff c.RW is false. Note that unlike RunRW, this
// does not start a new subtest. This should usually be run in conjunction with
// RunRW, to run only RO or RW tests, in situations where running both would be
// redundant.
func (c *Context) RunRO(fn testFunc) {
	if !c.RW {
		fn(c)
	}
}

// Errorf is a wrapper around t.Errorf()
func (c *Context) Errorf(format string, args ...interface{}) {
	c.T.Errorf(format, args...)
}

// Parallel is a wrapper around t.Parallel()
func (c *Context) Parallel() {
	c.T.Parallel()
}
