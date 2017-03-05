// +build !js

package test

import (
	"net/http/httptest"
	"testing"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/config"
	"github.com/flimzy/kivik/driver"
	_ "github.com/flimzy/kivik/driver/couchdb"
	_ "github.com/flimzy/kivik/driver/memory"
	"github.com/flimzy/kivik/driver/proxy"
	"github.com/flimzy/kivik/logger/memlogger"
	"github.com/flimzy/kivik/serve"
	"github.com/flimzy/kivik/serve/config/memconf"
)

type customDriver struct {
	driver.Client
	driver.Logger
}

func (cd customDriver) NewClient(_ string) (driver.Client, error) {
	return cd, nil
}

func TestServer(t *testing.T) {
	memClient, _ := kivik.New("memory", "")
	log := &memlogger.Logger{}
	kivik.Register("custom", customDriver{
		Client: proxy.NewClient(memClient),
		Logger: log,
	})
	service := serve.Service{}
	backend, err := kivik.New("custom", "")
	if err != nil {
		t.Fatalf("Failed to connect to custom driver: %s", err)
	}
	service.Client = backend
	service.LogWriter = log
	service.SetConfig(config.New(memconf.New()))
	service.Config().Set("log", "capacity", "10")
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
