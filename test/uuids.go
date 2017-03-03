package test

func init() {
	for _, suite := range []string{SuiteCouch16, SuiteCouch20, SuiteCloudant, SuiteKivikMemory} { // FIXME: SuiteKivikServer,
		RegisterTest(suite, "UUIDs", false, UUIDs)
	}
}

// UUIDs tests the '/_uuids' endpoint
func UUIDs(clients *Clients, _ string, fail FailFunc) {
	client := clients.Admin
	uuidCount := 3
	uuids, err := client.UUIDs(uuidCount)
	if err != nil {
		fail("Failed to get UUIDs: %s", err)
	}
	if len(uuids) != uuidCount {
		fail("Expected %d UUIDs, got %d\n", uuidCount, len(uuids))
	}
}
