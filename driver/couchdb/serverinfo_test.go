package couchdb

import "testing"

func TestServerInfo(t *testing.T) {
	client := getClient(t)
	_, err := client.ServerInfoContext(CTX)
	if err != nil {
		t.Fatalf("Faled to get server info: %s", err)
	}
}
