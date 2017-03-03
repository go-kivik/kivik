package test

import "testing"

func init() {
	for _, suite := range []string{SuiteCouch16, SuiteCouch20, SuiteCloudant, SuiteKivikMemory} { // FIXME: SuiteKivikServer,
		RegisterTest(suite, "UUIDs", false, UUIDs)
	}
}

// UUIDs tests the '/_uuids' endpoint
func UUIDs(clients *Clients, _ string, t *testing.T) {
	client := clients.Admin
	uuidCount := 3
	uuids, err := client.UUIDs(uuidCount)
	if err != nil {
		t.Errorf("Failed to get UUIDs: %s", err)
	}
	if len(uuids) != uuidCount {
		t.Errorf("Expected %d UUIDs, got %d\n", uuidCount, len(uuids))
	}
}
