// +build !js,ignore

// This test is disabled in GopherJS, becuase node is ultra-picky about the
// apparently mal-formed headers that CouchDB sends on these requests.

package test

import (
	"net/http"
	"testing"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
)

func init() {
	for _, suite := range []string{SuiteCouch16, SuiteCouch20, SuiteCloudant, SuiteKivikServer} {
		RegisterTest(suite, "Log", false, Log)
	}
	// RegisterTest(SuiteCloudant, "Log", false, CloudantLog)
}

// Log tests the /_log endpoint
func Log(clients *Clients, suite string, t *testing.T) {
	t.Run("Admin", func(t *testing.T) {
		status := kivik.StatusNoError
		if suite == SuiteCloudant {
			status = http.StatusForbidden
		}
		testLogs(clients.Admin, status, t)
	})
	if clients.NoAuth == nil {
		return
	}
	t.Run("NoAuth", func(t *testing.T) {
		testLogs(clients.NoAuth, kivik.StatusUnauthorized, t)
	})
}

func testLogs(client *kivik.Client, status int, t *testing.T) {
	t.Parallel()
	logBuf := make([]byte, 1000)
	testLog(client, logBuf, status, t)
	logBuf = make([]byte, 0, 1000)
	testLog(client, logBuf, status, t)
}

func testLog(client *kivik.Client, buf []byte, status int, t *testing.T) {
	_, err := client.Log(buf, 0)
	switch errors.StatusCode(err) {
	case status:
		// Expected
	case 0:
		t.Errorf("Expected failure %d/%s", status, http.StatusText(status))
	default:
		t.Errorf("Failed to read %d log bytes: %s", len(buf), err)
	}
}
