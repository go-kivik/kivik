// +build !js

// This test is disabled in GopherJS, becuase node is ultra-picky about the
// apparently mal-formed headers that CouchDB sends on these requests.

package test

import (
	"testing"

	"github.com/flimzy/kivik"
)

func init() {
	for _, suite := range []string{SuiteCouch16, SuiteCouch20} { //FIXME: SuiteKivikServer
		RegisterTest(suite, "Log", false, Log)
	}
	RegisterTest(SuiteCloudant, "Log", false, CloudantLog)
}

// Log tests the /_log endpoint
func Log(clients *Clients, suite string, t *testing.T) {
	client := clients.Admin
	logBuf := make([]byte, 1000)
	if _, err := client.Log(logBuf, 0); err != nil {
		t.Errorf("Error reading 1000 log bytes: %s", err)
	}
	logBuf = make([]byte, 0, 1000)
	if _, err := client.Log(logBuf, 0); err != nil {
		t.Errorf("Error reading 0 log bytes: %s", err)
	}
}

// CloudantLog tests the /_log endpoint for Cloudant, which returns an error
func CloudantLog(clients *Clients, suite string, t *testing.T) {
	client := clients.Admin
	logBuf := make([]byte, 1000)
	_, err := client.Log(logBuf, 0)
	if err == nil {
		t.Errorf("Expected an error")
	}
	if !kivik.ErrForbidden(err) {
		t.Errorf("Expected 403/Forbidden, got %s", err)
	}
}
