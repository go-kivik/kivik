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

// Package kt provides common utilities for Kivik tests.
package kt

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"

	kivik "github.com/go-kivik/kivik/v4"
)

// Context is a collection of client connections with different security access.
type Context struct {
	*ContextCore
	// T is the *testing.T value
	T *testing.T
}

// Child returns a shallow copy of itself with a new t.
func (c *Context) Child(t *testing.T) *Context {
	t.Helper()
	return &Context{
		ContextCore: c.ContextCore,
		T:           t,
	}
}

// Skip will skip the currently running test if configuration dictates.
func (c *Context) Skip() {
	c.T.Helper()
	c.ContextCore.Skip(c.T)
}

// Skipf is a wrapper around t.Skipf()
func (c *Context) Skipf(format string, args ...any) {
	c.T.Helper()
	c.T.Skipf(format, args...)
}

// Logf is a wrapper around t.Logf()
func (c *Context) Logf(format string, args ...any) {
	c.T.Helper()
	c.T.Logf(format, args...)
}

// Fatalf is a wrapper around t.Fatalf()
func (c *Context) Fatalf(format string, args ...any) {
	c.T.Helper()
	c.T.Fatalf(format, args...)
}

// MustBeSet ends the test with a failure if the config key is not set.
func (c *Context) MustBeSet(key string) {
	c.T.Helper()
	c.ContextCore.MustBeSet(c.T, key)
}

// MustStringSlice returns a string slice, or fails if the value is unset.
func (c *Context) MustStringSlice(key string) []string {
	c.T.Helper()
	return c.ContextCore.MustStringSlice(c.T, key)
}

// MustBool returns a bool, or fails if the value is unset.
func (c *Context) MustBool(key string) bool {
	c.T.Helper()
	return c.ContextCore.MustBool(c.T, key)
}

// IntSlice returns a []int from config.
func (c *Context) IntSlice(key string) []int {
	c.T.Helper()
	return c.ContextCore.IntSlice(c.T, key)
}

// MustIntSlice returns a []int, or fails if the value is unset.
func (c *Context) MustIntSlice(key string) []int {
	c.T.Helper()
	return c.ContextCore.MustIntSlice(c.T, key)
}

// StringSlice returns a string slice from the config.
func (c *Context) StringSlice(key string) []string {
	c.T.Helper()
	return c.ContextCore.StringSlice(c.T, key)
}

// String returns a string from config.
func (c *Context) String(key string) string {
	c.T.Helper()
	return c.ContextCore.String(c.T, key)
}

// MustString returns a string from config, or fails if the value is unset.
func (c *Context) MustString(key string) string {
	c.T.Helper()
	return c.ContextCore.MustString(c.T, key)
}

// Int returns an int from the config.
func (c *Context) Int(key string) int {
	c.T.Helper()
	return c.ContextCore.Int(c.T, key)
}

// MustInt returns an int from the config, or fails if the value is unset.
func (c *Context) MustInt(key string) int {
	c.T.Helper()
	return c.ContextCore.MustInt(c.T, key)
}

// Bool returns a bool from the config.
func (c *Context) Bool(key string) bool {
	c.T.Helper()
	return c.ContextCore.Bool(c.T, key)
}

// Interface returns the configuration value as an any.
func (c *Context) Interface(key string) any {
	c.T.Helper()
	return c.ContextCore.Interface(c.T, key)
}

// Options returns an options map value.
func (c *Context) Options(key string) kivik.Option {
	c.T.Helper()
	return c.ContextCore.Options(c.T, key)
}

// MustInterface returns an any from the config, or fails if the value is unset.
func (c *Context) MustInterface(key string) any {
	c.T.Helper()
	return c.ContextCore.MustInterface(c.T, key)
}

// IsSet returns true if the value is set in the configuration.
func (c *Context) IsSet(key string) bool {
	c.T.Helper()
	return c.ContextCore.IsSet(c.T, key)
}

// Run wraps t.Run()
func (c *Context) Run(name string, fn testFunc) {
	c.T.Helper()
	c.T.Run(name, func(t *testing.T) {
		c.T.Helper()
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
}

// TestFunc is the signature for tests that receive *testing.T directly.
type TestFunc func(*testing.T, *ContextCore)

// RegisterV2 registers a test using the new signature.
func RegisterV2(name string, fn TestFunc) {
	Register(name, func(ctx *Context) {
		fn(ctx.T, ctx.ContextCore)
	})
}

// RunSubtests executes the requested suites of tests against the client.
func RunSubtests(ctx *Context) {
	for name, fn := range tests {
		ctx.Run(name, fn)
	}
}

var (
	rnd   *rand.Rand
	rndMU = &sync.Mutex{}
)

func init() {
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// TestDBPrefix is used to prefix temporary database names during tests.
const TestDBPrefix = "kivik$"

// TestDB creates a test database, registers a cleanup function to destroy it,
// and returns its name.
func (c *Context) TestDB() string {
	c.T.Helper()
	return c.ContextCore.TestDB(c.T)
}

// TestDBName generates a randomized string suitable for a database name for
// testing.
func (c *Context) TestDBName() string {
	return TestDBName(c.T)
}

var invalidDBCharsRE = regexp.MustCompile(`[^a-z0-9_$\(\)+/-]`)

// TestDBName generates a randomized string suitable for a database name for
// testing.
func TestDBName(t *testing.T) string {
	id := strings.ToLower(t.Name())
	id = invalidDBCharsRE.ReplaceAllString(id, "_")
	id = id[strings.Index(id, "/")+1:]
	id = strings.ReplaceAll(id, "/", "_") + "$"
	rndMU.Lock()
	dbname := fmt.Sprintf("%s%s%016x", TestDBPrefix, id, rnd.Int63())
	rndMU.Unlock()
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
func (c *Context) Errorf(format string, args ...any) {
	c.T.Helper()
	c.T.Errorf(format, args...)
}

// Parallel is a wrapper around t.Parallel()
func (c *Context) Parallel() {
	c.T.Parallel()
}

const maxRetries = 5

// Retry will try an operation up to maxRetries times, in case of one of the
// following failures. All other failures are returned.
func Retry(fn func() error) error {
	bo := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), maxRetries)
	var i int
	return backoff.Retry(func() error {
		err := fn()
		if err != nil {
			if shouldRetry(err) {
				i++
				return fmt.Errorf("attempt #%d failed: %w", i, err)
			}
			return backoff.Permanent(err)
		}
		return nil
	}, bo)
}

func shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	var statusErr interface {
		error
		HTTPStatus() int
	}
	if errors.As(err, &statusErr) {
		if status := statusErr.HTTPStatus(); status < http.StatusInternalServerError {
			return false
		}
	}
	var errno syscall.Errno
	if errors.As(err, &errno) {
		switch errno {
		case syscall.ECONNRESET, syscall.EPIPE:
			return true
		}
	}
	urlErr := new(url.Error)
	if errors.As(err, &urlErr) {
		// Seems string comparison is necessary in some cases.
		msg := strings.TrimSpace(urlErr.Error())
		return strings.HasSuffix(msg, ": connection reset by peer") || // Observed in Go 1.18
			strings.HasSuffix(msg, ": broken pipe") // Observed in Go 1.19 & 1.17
	}
	return false
}

// Body turns a string into a read closer, useful as a request or attachment
// body.
func Body(str string, args ...any) io.ReadCloser {
	return io.NopCloser(strings.NewReader(fmt.Sprintf(str, args...)))
}
