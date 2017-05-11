// Package kt provides common utilities for Kivik tests.
package kt

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver/couchdb/chttp"
)

// Context is a collection of client connections with different security access.
type Context struct {
	// RW is true if we should run read-write tests.
	RW bool
	// Admin is a client connection with database admin privileges.
	Admin *kivik.Client
	// CHTTPAdmin is a chttp connection with admin privileges.
	CHTTPAdmin *chttp.Client
	// NoAuth isa client connection with no authentication.
	NoAuth *kivik.Client
	// CHTTPNoAuth is a chttp connection with no authentication.
	CHTTPNoAuth *chttp.Client
	// Config is the suite config
	Config SuiteConfig
	// T is the *testing.T value
	T *testing.T
}

// Child returns a shallow copy of itself with a new t.
func (c *Context) Child(t *testing.T) *Context {
	return &Context{
		RW:          c.RW,
		Admin:       c.Admin,
		CHTTPAdmin:  c.CHTTPAdmin,
		NoAuth:      c.NoAuth,
		CHTTPNoAuth: c.CHTTPNoAuth,
		Config:      c.Config,
		T:           t,
	}
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

// String returns a string from config.
func (c *Context) String(key string) string {
	return c.Config.String(c.T, key)
}

// MustString returns a string from config, or fails if the value is unset.
func (c *Context) MustString(key string) string {
	c.MustBeSet(key)
	return c.String(key)
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

// Interface returns the configuration value as an interface{}.
func (c *Context) Interface(key string) interface{} {
	return c.Config.get(name(c.T), key)
}

// Options returns an options map value.
func (c *Context) Options(key string) map[string]interface{} {
	i := c.Config.get(name(c.T), key)
	o, _ := i.(map[string]interface{})
	return o
}

// MustInterface returns an interface{} from the config, or fails if the value is unset.
func (c *Context) MustInterface(key string) interface{} {
	c.MustBeSet(key)
	return c.Interface(key)
}

// IsSet returns true if the value is set in the configuration.
func (c *Context) IsSet(key string) bool {
	return c.Interface(key) != nil
}

// Run wraps t.Run()
func (c *Context) Run(name string, fn testFunc) {
	c.T.Run(name, func(t *testing.T) {
		ctx := c.Child(t)
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
var rndMU = &sync.Mutex{}

func init() {
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// TestDBPrefix is used to prefix temporary database names during tests.
const TestDBPrefix = "kivik$"

// TestDB creates a test database and returns its name.
func (c *Context) TestDB() string {
	var dbname string
	err := Retry(func() error {
		dbname = c.TestDBName()
		return c.Admin.CreateDB(context.Background(), dbname, c.Options("db"))
	})
	if err != nil {
		c.Fatalf("Failed to create database: %s", err)
	}
	return dbname
}

// TestDBName generates a randomized string suitable for a database name for
// testing.
func (c *Context) TestDBName() string {
	return TestDBName(c.T)
}

// TestDBName generates a randomized string suitable for a database name for
// testing.
func TestDBName(t *testing.T) string {
	id := strings.ToLower(tName(t))
	id = id[strings.Index(id, "/")+1:]
	id = strings.Replace(id, "/", "_", -1) + "$"
	rndMU.Lock()
	defer rndMU.Unlock()
	dbname := fmt.Sprintf("%s%s%016x", TestDBPrefix, id, rnd.Int63())
	return dbname
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

// Retry will try an operation up to 3 times, in case of one of the following
// failures. All other failures are returned.
func Retry(fn func() error) error {
	var err error
	for i := 0; i < 3; i++ {
		err = fn()
		if err != nil {
			msg := strings.TrimSpace(err.Error())
			if strings.HasSuffix(msg, "io: read/write on closed pipe") ||
				strings.HasSuffix(msg, "write: broken pipe") ||
				strings.HasSuffix(msg, ": EOF") ||
				strings.HasSuffix(msg, ": http: server closed idle connection") ||
				strings.HasSuffix(msg, "read: connection reset by peer") {
				fmt.Printf("Retrying after error: %s\n", err)
				continue
			}
		}
		break
	}
	return err
}
