package test

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/flimzy/kivik"

	// Tests
	_ "github.com/flimzy/kivik/test/client"
	"github.com/flimzy/kivik/test/kt"
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
	Cleanup bool
}

// CleanupTests attempts to clean up any stray test databases created by a
// previous test run.
func CleanupTests(driver, dsn string, verbose bool) error {
	client, err := kivik.New(driver, dsn)
	if err != nil {
		return err
	}
	count, err := doCleanup(client, verbose)
	if verbose {
		fmt.Printf("Deleted %d test databases\n", count)
	}
	return err
}

func doCleanup(client *kivik.Client, verbose bool) (int, error) {
	allDBs, err := client.AllDBs()
	if err != nil {
		return 0, err
	}
	var count int
	for _, dbName := range allDBs {
		if strings.HasPrefix(dbName, kt.TestDBPrefix) {
			if verbose {
				fmt.Printf("\t--- Deleting %s\n", dbName)
			}
			err := client.DestroyDB(dbName)
			if err != nil {
				return count, err
			}
			count++
		}
	}
	return count, nil
}

// RunTests runs the requested test suites against the requested driver and DSN.
func RunTests(opts Options) {
	if opts.Cleanup {
		err := CleanupTests(opts.Driver, opts.DSN, opts.Verbose)
		if err != nil {
			fmt.Printf("Cleanup failed: %s\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	flag.Set("test.run", opts.Match)
	if opts.Verbose {
		flag.Set("test.v", "true")
	}
	tests := []testing.InternalTest{
		testing.InternalTest{
			Name: "MainTest",
			F: func(t *testing.T) {
				Test(opts.Driver, opts.DSN, opts.Suites, opts.RW, t)
			},
		},
	}

	mainStart(tests)
}

// Test is the main test entry point when running tests through the command line
// tool.
func Test(driver, dsn string, testSuites []string, rw bool, t *testing.T) {
	clients, err := connectClients(driver, dsn, t)
	if err != nil {
		t.Fatalf("Failed to connect to %s (%s driver): %s\n", dsn, driver, err)
	}
	clients.RW = rw
	tests := make(map[string]struct{})
	for _, test := range testSuites {
		tests[test] = struct{}{}
	}
	if _, ok := tests[SuiteAuto]; ok {
		t.Log("Detecting target service compatibility...")
		suites, err := detectCompatibility(clients.Admin)
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
	t.Logf("Running the following test suites: %s\n", strings.Join(testSuites, ", "))
	for _, suite := range testSuites {
		runTests(clients, suite, t)
	}
}

func runTests(clients *kt.Clients, suite string, t *testing.T) {
	conf, ok := suites[suite]
	if !ok {
		t.Skipf("No configuration found for suite '%s'", suite)
	}
	kt.RunSubtests(clients, conf, t)
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

func connectClients(driverName, dsn string, t *testing.T) (*kt.Clients, error) {
	var noAuthDSN string
	if parsed, err := url.Parse(dsn); err == nil {
		if parsed.User == nil {
			return nil, errors.New("DSN does not contain authentication credentials")
		}
		parsed.User = nil
		noAuthDSN = parsed.String()
	}
	clients := &kt.Clients{}
	t.Logf("Connecting to %s ...\n", dsn)
	if client, err := kivik.New(driverName, dsn); err == nil {
		clients.Admin = client
	} else {
		return nil, err
	}

	t.Logf("Connecting to %s ...\n", noAuthDSN)
	if client, err := kivik.New(driverName, noAuthDSN); err == nil {
		clients.NoAuth = client
	} else {
		return nil, err
	}

	return clients, nil
}
