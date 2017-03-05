// +build !js

package test

import (
	"net/http/httptest"
	"testing"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/config"
	_ "github.com/flimzy/kivik/driver/couchdb"
	_ "github.com/flimzy/kivik/driver/memory"
	"github.com/flimzy/kivik/driver/proxy"
	"github.com/flimzy/kivik/logger/memlogger"
	"github.com/flimzy/kivik/serve"
	"github.com/flimzy/kivik/serve/config/memconf"
)

func TestServer(t *testing.T) {
	memClient, _ := kivik.New("memory", "")
	logger := &memlogger.Logger{}
	backend := &serve.LoggingClient{
		Client: proxy.NewClient(memClient),
		Logger: logger,
	}
	service := serve.Service{}
	service.Client = backend
	service.LogWriter = logger
	service.Config = config.New(memconf.New())
	service.Config.Set("log", "capacity", "10")
	handler, err := service.Init()
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
