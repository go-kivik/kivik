// +build !js

package test

import (
	"net/http/httptest"
	"testing"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/couchdb"
	_ "github.com/flimzy/kivik/driver/memory"
	"github.com/flimzy/kivik/driver/proxy"
	"github.com/flimzy/kivik/logger/memory"
	"github.com/flimzy/kivik/serve"
)

func TestServer(t *testing.T) {
	memClient, _ := kivik.New("memory", "")
	backend := &serve.LoggingClient{
		Client: proxy.NewClient(memClient),
		Logger: memory.New(10),
	}
	handler, err := serve.New(backend).Server()
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
