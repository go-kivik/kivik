package test

import "github.com/flimzy/kivik"

func init() {
	for _, suite := range []string{SuiteCouch, SuiteCouch20, SuiteCloudant, SuiteKivikMemory} { // FIXME: SuiteKivikServer,
		RegisterTest(suite, "UUIDs", false, UUIDs)
	}
}

// UUIDs tests the '/_uuids' endpoint
func UUIDs(client *kivik.Client, _ string, fail FailFunc) {
	uuidCount := 3
	uuids, err := client.UUIDs(uuidCount)
	if err != nil {
		fail("Failed to get UUIDs: %s", err)
	}
	if len(uuids) != uuidCount {
		fail("Expected %d UUIDs, got %d\n", uuidCount, len(uuids))
	}
}
