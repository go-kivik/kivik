// +build !js

package test

import (
	"net/http/httptest"
	"testing"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/couchdb"
	_ "github.com/flimzy/kivik/driver/memory"
	"github.com/flimzy/kivik/serve"
)

func TestServer(t *testing.T) {
	handler, err := serve.New("memory", "").Server()
	if err != nil {
		t.Fatalf("Failed to initialize server: %s\n", err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()

	client, err := kivik.New("couch", server.URL)
	if err != nil {
		t.Fatalf("Failed to initialize client: %s\n", err)
	}
	clients := &Clients{
		Admin: client,
	}
	RunSubtests(clients, true, SuiteKivikServer, t)
}
