package test

import (
	"testing"

	"github.com/flimzy/kivik"
)

func init() {
	for _, suite := range []string{SuiteCouch16, SuiteCouch20, SuiteCloudant, SuiteKivikMemory} { // FIXME: SuiteKivikServer,
		RegisterTest(suite, "UUIDs", false, UUIDs)
	}
}

// UUIDs tests the '/_uuids' endpoint
func UUIDs(clients *Clients, _ string, t *testing.T) {
	t.Run("Admin", func(t *testing.T) {
		testUUIDs(clients.Admin, t)
	})
	if clients.NoAuth == nil {
		return
	}
	t.Run("NoAuth", func(t *testing.T) {
		testUUIDs(clients.NoAuth, t)
	})
}

func testUUIDs(client *kivik.Client, t *testing.T) {
	t.Parallel()
	uuidCount := 3
	uuids, err := client.UUIDs(uuidCount)
	if err != nil {
		t.Errorf("Failed to get UUIDs: %s", err)
	}
	if len(uuids) != uuidCount {
		t.Errorf("Expected %d UUIDs, got %d\n", uuidCount, len(uuids))
	}
}
