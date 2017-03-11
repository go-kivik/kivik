// +build !js

package test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/auth/confadmin"
	"github.com/flimzy/kivik/config"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/proxy"
	"github.com/flimzy/kivik/logger/memlogger"
	"github.com/flimzy/kivik/serve"
	"github.com/flimzy/kivik/serve/config/memconf"
	"github.com/flimzy/kivik/test/kt"
)

type customDriver struct {
	driver.Client
	driver.LogReader
}

func (cd customDriver) NewClientContext(_ context.Context, _ string) (driver.Client, error) {
	return cd, nil
}

func TestServer(t *testing.T) {
	memClient, _ := kivik.New("memory", "")
	log := &memlogger.Logger{}
	kivik.Register("custom", customDriver{
		Client:    proxy.NewClient(memClient),
		LogReader: log,
	})
	service := serve.Service{}
	backend, err := kivik.New("custom", "")
	if err != nil {
		t.Fatalf("Failed to connect to custom driver: %s", err)
	}
	conf := config.New(memconf.New())
	conf.Set("log", "capacity", "10")
	// Set admin/abc123 credentials
	conf.Set("admins", "admin", "-pbkdf2-792221164f257de22ad72a8e94760388233e5714,7897f3451f59da741c87ec5f10fe7abe,10")
	service.Client = backend
	service.AuthHandler = confadmin.New(conf)
	service.LogWriter = log
	service.SetConfig(conf)
	handler, err := service.Init()
	if err != nil {
		t.Fatalf("Failed to initialize server: %s\n", err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()

	dsn, _ := url.Parse(server.URL)
	dsn.User = url.UserPassword("admin", "abc123")
	fmt.Printf("dsn = %s\n", dsn.String())
	client, err := kivik.New("couch", dsn.String())
	if err != nil {
		t.Fatalf("Failed to initialize client: %s\n", err)
	}
	clients := &kt.Context{
		RW:    true,
		Admin: client,
	}
	runTests(clients, SuiteKivikServer, t)
}
