package test

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/flimzy/kivik"
	"github.com/pkg/errors"
)

// The available test suites
const (
	SuiteMinimal     = "minimal"
	SuiteAuto        = "auto"
	SuitePouch       = "pouch"
	SuiteCouch       = "couch"
	SuiteCouch20     = "couch2.0"
	SuiteKivikMemory = "kivikmemory"
	SuiteCloudant    = "cloudant"
)

// AllSuites is a list of all defined suites.
var AllSuites = []string{SuiteMinimal, SuitePouch, SuiteCouch, SuiteCouch20, SuiteKivikMemory, SuiteCloudant}

// RunTests runs the requested test suites against the requested driver and DSN.
func RunTests(driver, dsn string, testSuites []string, run string) error {
	runRE, err := regexp.Compile(run)
	if err != nil {
		return err
	}
	fmt.Printf("Connecting to %s ...\n", dsn)
	client, err := kivik.New(driver, dsn)
	if err != nil {
		return err
	}
	tests := make(map[string]struct{})
	for _, test := range testSuites {
		tests[test] = struct{}{}
	}
	if _, ok := tests[SuiteAuto]; ok {
		fmt.Printf("Detecting target service compatibility...\n")
		t, err := detectCompatibility(client)
		if err != nil {
			return errors.Wrap(err, "failed to determine server compatibility")
		}
		tests = make(map[string]struct{})
		for _, test := range t {
			tests[test] = struct{}{}
		}
	}
	testSuites = make([]string, 0, len(tests))
	for test := range tests {
		testSuites = append(testSuites, test)
	}
	fmt.Printf("Going to run the following test suites: %s\n", strings.Join(testSuites, ", "))
	if result := testMain(client, testSuites, runRE); result > 0 {
		fmt.Println("FAILED")
		os.Exit(result)
	}
	return nil
}

func detectCompatibility(client *kivik.Client) ([]string, error) {
	info, err := client.ServerInfo()
	if err != nil {
		return nil, err
	}
	switch info.Vendor() {
	case "PouchDB":
		return []string{SuitePouch}, nil
	case "IBM Cloudant":
		return []string{SuiteCloudant}, nil
	case "The Apache Software Foundation":
		if info.Version() == "2.0" {
			return []string{SuiteCouch20}, nil
		}
		return []string{SuiteCouch}, nil
	case "Kivik Memory Adaptor":
		return []string{SuiteKivikMemory}, nil
	}
	return []string{SuiteMinimal}, nil
}

type testFunc func(*kivik.Client, string, FailFunc)

// tests is a map of the format map[suite]map[name]testFunc
var tests = make(map[string]map[string]testFunc)

// RegisterTest registers a test to be run for the given test suite.
func RegisterTest(suite, name string, fn testFunc) {
	if _, ok := tests[suite]; !ok {
		tests[suite] = make(map[string]testFunc)
	}
	tests[suite][name] = fn
}

// FailFunc is passed to each test, and should be called whenever a test fails.
type FailFunc func(format string, args ...interface{})

func testMain(client *kivik.Client, suites []string, runRE *regexp.Regexp) int {
	var failed bool
	for _, suite := range suites {
		for name, fn := range tests[suite] {
			if !runRE.MatchString(name) {
				continue
			}
			fail := func(format string, args ...interface{}) {
				format = fmt.Sprintf("[%s] %s: %s\n", suite, name, strings.TrimSpace(format))
				fmt.Printf(format, args...)
				failed = true

			}
			fn(client, suite, fail)
		}
	}
	if failed {
		return 1
	}
	return 0
}
