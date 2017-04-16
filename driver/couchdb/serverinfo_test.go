package couchdb

import (
	"context"
	"testing"
)

func TestServerInfo(t *testing.T) {
	client := getClient(t)
	_, err := client.ServerInfo(context.Background(), nil)
	if err != nil {
		t.Fatalf("Faled to get server info: %s", err)
	}
}
