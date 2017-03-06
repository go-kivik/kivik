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

// Clients is a collection of client connections with different security access.
type Clients struct {
	// RW is true if we should run read-write tests.
	RW bool
	// Admin is a client connection with database admin priveleges.
	Admin *kivik.Client
	// NoAuth isa client connection with no authentication.
	NoAuth *kivik.Client
}

type testFunc func(*Clients, SuiteConfig, *testing.T)

// tests is a map of the format map[suite]map[name]testFunc
var tests = make(map[string]testFunc)

// Register registers a test to be run for the given test suite. rw should
// be true if the test writes to the database.
func Register(name string, fn testFunc) {
	tests[name] = fn
	return
}

// RunSubtests executes the requested suites of tests against the client.
func RunSubtests(clients *Clients, conf SuiteConfig, t *testing.T) {
	for name, fn := range tests {
		t.Run(name, func(t *testing.T) {
			conf.Skip(t)
			fn(clients, conf, t)
		})
	}
}

var rnd *rand.Rand

func init() {
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// TestDBPrefix is used to prefix temporary database names during tests.
const TestDBPrefix = "kivik$"

// TestDBName generates a randomized string suitable for a database name for
// testing.
func TestDBName(t *testing.T) string {
	id := strings.ToLower(t.Name())
	id = id[strings.Index(id, "/")+1:]
	id = strings.Replace(id, "/", "_", -1)
	return fmt.Sprintf("%s%s$%016x", TestDBPrefix, id, rnd.Int63())
}

// RunAdmin runs the test function iff c.Admin is set.
func (c *Clients) RunAdmin(t *testing.T, fn func(*testing.T)) {
	if c.Admin != nil {
		t.Run("Admin", fn)
	}
}

// RunNoAuth runs the test function iff c.NoAuth is set.
func (c *Clients) RunNoAuth(t *testing.T, fn func(*testing.T)) {
	if c.NoAuth != nil {
		t.Run("NoAuth", fn)
	}
}

// RunRW runs the test function iff c.RW is true.
func (c *Clients) RunRW(t *testing.T, fn func(*testing.T)) {
	if c.RW {
		t.Run("RW", fn)
	}
}
