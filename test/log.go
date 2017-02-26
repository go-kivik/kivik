// +build !js

// This test is disabled in GopherJS, becuase node is ultra-picky about the
// apparently mal-formed headers that CouchDB sends on these requests.

package test

import "github.com/flimzy/kivik"

func init() {
	for _, suite := range []string{SuiteCouch, SuiteCouch20} { //FIXME: SuiteKivikServer
		RegisterTest(suite, "Log", false, Log)
	}
	RegisterTest(SuiteCloudant, "Log", false, CloudantLog)
}

// Log tests the /_log endpoint
func Log(client *kivik.Client, suite string, fail FailFunc) {
	logBuf := make([]byte, 1000)
	if _, err := client.Log(logBuf, 0); err != nil {
		fail("Error reading 1000 log bytes: %s", err)
	}
	logBuf = make([]byte, 0, 1000)
	if _, err := client.Log(logBuf, 0); err != nil {
		fail("Error reading 0 log bytes: %s", err)
	}
}

// CloudantLog tests the /_log endpoint for Cloudant, which returns an error
func CloudantLog(client *kivik.Client, suite string, fail FailFunc) {
	logBuf := make([]byte, 1000)
	_, err := client.Log(logBuf, 0)
	if err == nil {
		fail("Expected an error")
	}
	if !kivik.ErrForbidden(err) {
		fail("Expected 403/Forbidden, got %s", err)
	}
}
