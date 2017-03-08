package test

import "github.com/flimzy/kivik/test/kt"

var suites = make(map[string]kt.SuiteConfig)

// RegisterSuite registers a Suite as available for testing.
func RegisterSuite(suite string, conf kt.SuiteConfig) {
	suites[suite] = conf
}
