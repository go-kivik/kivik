package test

import (
	"errors"
	"fmt"
	"math/rand"
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
	SuitePouchRemote = "pouchremote"
	SuiteCouch16     = "couch16"
	SuiteCouch20     = "couch20"
	SuiteCloudant    = "cloudant"
	SuiteKivikServer = "kivikserver"
	SuiteKivikMemory = "kivikmemory"
	SuiteKivikFS     = "kivikfilesystem"
)

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

// AllSuites is a list of all defined suites.
var AllSuites = []string{SuitePouchLocal, SuitePouchRemote, SuiteCouch16, SuiteCouch20, SuiteKivikMemory, SuiteKivikFS, SuiteCloudant, SuiteKivikServer}

// ListTests prints a list of available test suites to stdout.
func ListTests() {
	fmt.Printf("Available test suites:\n\tauto\n")
	for _, suite := range AllSuites {
		fmt.Printf("\t%s\n", suite)
	}
}

// Tester mimics (or can be fulfilled by) a *testin.T struct
type Tester interface {
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

type tester struct {
	verbose bool
	failed  bool
}

func (t *tester) Errorf(format string, args ...interface{}) {
	t.failed = true
	fmt.Printf(format, args...)
}

func (t *tester) Fatalf(format string, args ...interface{}) {
	t.Errorf(format, args...)
	os.Exit(1)
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
	t := &tester{
		verbose: opts.Verbose,
	}
	mainTest(opts.Driver, opts.DSN, opts.RW, opts.Suites, t)
	if t.failed {
		fmt.Printf("FAILED\n")
		os.Exit(1)
	}
	fmt.Printf("PASSED\n")
	os.Exit(0)
}

func mainTest(driver, dsn string, rw bool, testSuites []string, t Tester) {
	fmt.Printf("Connecting to %s ...\n", dsn)
	client, err := kivik.New(driver, dsn)
	if err != nil {
		t.Fatalf("Failed to connect to %s (%s driver): %s\n", dsn, driver, err)
	}
	tests := make(map[string]struct{})
	for _, test := range testSuites {
		tests[test] = struct{}{}
	}
	if _, ok := tests[SuiteAuto]; ok {
		fmt.Printf("Detecting target service compatibility...\n")
		suites, err := detectCompatibility(client)
		if err != nil {
			t.Fatalf("Unable to determine server suite compatibility: %s\n", err)
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
	RunSubtests(client, rw, testSuites, t)
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

type testFunc func(*kivik.Client, string, FailFunc)

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

// FailFunc is passed to each test, and should be called whenever a test fails.
type FailFunc func(format string, args ...interface{})

// RunSubtests executes the requested suites of tests against the client.
func RunSubtests(client *kivik.Client, rw bool, suites []string, t Tester) {
	for _, suite := range suites {
		for name, fn := range tests[suite] {
			runTest(client, name, suite, fn, t)
		}
		if rw {
			for name, fn := range rwtests[suite] {
				runTest(client, name, suite, fn, t)
			}
		}
	}
}

func runTest(client *kivik.Client, name, suite string, fn testFunc, t Tester) {
	fail := func(format string, args ...interface{}) {
		format = fmt.Sprintf("[%s] %s: %s\n", suite, name, strings.TrimSpace(format))
		t.Errorf(format, args...)
	}
	fn(client, suite, fail)
}

func gatherTests(client *kivik.Client, suites []string) []testing.InternalTest {
	internalTests := make([]testing.InternalTest, 0)
	for _, suite := range suites {
		for name, fn := range tests[suite] {
			// if !runRE.MatchString(name) {
			// 	continue
			// }
			internalTests = append(internalTests, testing.InternalTest{
				Name: name,
				F: func(t *testing.T) {
					fail := func(format string, args ...interface{}) {
						format = fmt.Sprintf("[%s] %s: %s\n", suite, name, strings.TrimSpace(format))
						t.Errorf(format, args...)
					}
					fn(client, suite, fail)
				},
			})
		}
	}
	return internalTests
}
