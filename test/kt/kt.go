// Package kt provides common utilities for Kivik tests.
package kt

import (
	"testing"

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
