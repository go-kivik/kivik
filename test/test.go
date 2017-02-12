package test

import (
	"fmt"
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

// Test runs the requested test suites against the requested driver and DSN.
func Test(driver, dsn string, testSuites []string) error {
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
