package couchdb

import (
	"testing"

	"github.com/flimzy/kivik/test/kt"
)

func TestServerInfo(t *testing.T) {
	client := getClient(t)
	_, err := client.ServerInfoContext(kt.CTX)
	if err != nil {
		t.Fatalf("Faled to get server info: %s", err)
	}
}
