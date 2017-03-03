package test

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/flimzy/kivik"
)

// The available test suites
const (
	SuiteAuto        = "auto"
	SuitePouchLocal  = "pouch"
	SuitePouchRemote = "pouchRemote"
	SuiteCouch16     = "couch16"
	SuiteCouch20     = "couch20"
	SuiteCloudant    = "cloudant"
	SuiteKivikServer = "kivikServer"
	SuiteKivikMemory = "kivikMemory"
	SuiteKivikFS     = "kivikFilesystem"
)

// AllSuites is a list of all defined suites.
var AllSuites = []string{
	SuitePouchLocal,
	SuitePouchRemote,
	SuiteCouch16,
	SuiteCouch20,
	SuiteKivikMemory,
	SuiteKivikFS,
	SuiteCloudant,
	SuiteKivikServer,
}

var driverMap = map[string]string{
	SuitePouchLocal:  "pouch",
	SuitePouchRemote: "pouch",
	SuiteCouch16:     "couch",
	SuiteCouch20:     "couch",
	SuiteCloudant:    "couch",
	SuiteKivikServer: "couch",
	SuiteKivikMemory: "memory",
	SuiteKivikFS:     "fs",
}

var rnd *rand.Rand

func init() {
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func testDBName() string {
	return fmt.Sprintf("kivik$%016x", rnd.Int63())
}

// ListTests prints a list of available test suites to stdout.
func ListTests() {
	fmt.Printf("Available test suites:\n\tauto\n")
	for _, suite := range AllSuites {
		fmt.Printf("\t%s\n", suite)
	}
}

// Options are the options to run a test from the command line tool.
type Options struct {
	Driver  string
	DSN     string
	Verbose bool
	RW      bool
	Match   string
	Suites  []string
}

// RunTests runs the requested test suites against the requested driver and DSN.
func RunTests(opts Options) {
	flag.Set("test.run", opts.Match)
	if opts.Verbose {
		flag.Set("test.v", "true")
	}

	clients, err := connectClients(opts.Driver, opts.DSN)
	if err != nil {
		fmt.Printf("Failed to connect to %s (%s driver): %s\n", opts.DSN, opts.Driver, err)
		os.Exit(1)
	}
	testSuites := opts.Suites
	tests := make(map[string]struct{})
	for _, test := range testSuites {
		tests[test] = struct{}{}
	}
	if _, ok := tests[SuiteAuto]; ok {
		fmt.Printf("Detecting target service compatibility...\n")
		suites, err := detectCompatibility(clients.Admin)
		if err != nil {
			fmt.Printf("Unable to determine server suite compatibility: %s\n", err)
			os.Exit(1)
		}
		tests = make(map[string]struct{})
		for _, suite := range suites {
			tests[suite] = struct{}{}
		}
	}
	testSuites = make([]string, 0, len(tests))
	for test := range tests {
		testSuites = append(testSuites, test)
	}
	fmt.Printf("Running the following test suites: %s\n", strings.Join(testSuites, ", "))
	mainStart(clients, testSuites, opts.RW)
}

func detectCompatibility(client *kivik.Client) ([]string, error) {
	info, err := client.ServerInfo()
	if err != nil {
		return nil, err
	}
	switch info.Vendor() {
	case "PouchDB":
		return []string{SuitePouchLocal}, nil
	case "IBM Cloudant":
		return []string{SuiteCloudant}, nil
	case "The Apache Software Foundation":
		if info.Version() == "2.0" {
			return []string{SuiteCouch20}, nil
		}
		return []string{SuiteCouch16}, nil
	case "Kivik Memory Adaptor":
		return []string{SuiteKivikMemory}, nil
	}
	return []string{}, errors.New("Unable to automatically determine the proper test suite")
}

type testFunc func(*Clients, string, *testing.T)

// tests is a map of the format map[suite]map[name]testFunc
var tests = make(map[string]map[string]testFunc)

var rwtests = make(map[string]map[string]testFunc)

// RegisterTest registers a test to be run for the given test suite. rw should
// be true if the test writes to the database.
func RegisterTest(suite, name string, rw bool, fn testFunc) {
	if rw {
		if _, ok := rwtests[suite]; !ok {
			rwtests[suite] = make(map[string]testFunc)
		}
		rwtests[suite][name] = fn
		return
	}
	if _, ok := tests[suite]; !ok {
		tests[suite] = make(map[string]testFunc)
	}
	tests[suite][name] = fn
}

// RunSubtests executes the requested suites of tests against the client.
func RunSubtests(clients *Clients, rw bool, suite string, t *testing.T) {
	for name, fn := range tests[suite] {
		runSubtest(clients, name, suite, fn, t)
	}
	if rw {
		for name, fn := range rwtests[suite] {
			runSubtest(clients, name, suite, fn, t)
		}
	}
}

func runSubtest(clients *Clients, name, suite string, fn testFunc, t *testing.T) {
	t.Run(name, func(t *testing.T) {
		fn(clients, suite, t)
	})
}

func internalTest(clients *Clients, name, suite string, fn testFunc) testing.InternalTest {
	return testing.InternalTest{
		Name: name,
		F: func(t *testing.T) {
			fn(clients, suite, t)
		},
	}
}

func gatherTests(clients *Clients, suites []string, rw bool) []testing.InternalTest {
	internalTests := make([]testing.InternalTest, 0)
	for _, suite := range suites {
		for name, fn := range tests[suite] {
			internalTests = append(internalTests, internalTest(clients, name, suite, fn))
		}
		if rw {
			for name, fn := range rwtests[suite] {
				internalTests = append(internalTests, internalTest(clients, name, suite, fn))
			}
		}
	}
	return internalTests
}

// Clients is a collection of client connections with different security access.
type Clients struct {
	Admin  *kivik.Client
	NoAuth *kivik.Client
}

func connectClients(driverName, dsn string) (*Clients, error) {
	var noAuthDSN string
	if parsed, err := url.Parse(dsn); err == nil {
		if parsed.User == nil {
			return nil, errors.New("DSN does not contain authentication credentials")
		}
		parsed.User = nil
		noAuthDSN = parsed.String()
	}
	clients := &Clients{}
	fmt.Printf("Connecting to %s ...\n", dsn)
	if client, err := kivik.New(driverName, dsn); err == nil {
		clients.Admin = client
	} else {
		return nil, err
	}

	fmt.Printf("Connecting to %s ...\n", noAuthDSN)
	if client, err := kivik.New(driverName, noAuthDSN); err == nil {
		clients.NoAuth = client
	} else {
		return nil, err
	}

	return clients, nil
}
