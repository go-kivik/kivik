// +build !js

package test

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/auth"
	"github.com/flimzy/kivik/auth/basic"
	"github.com/flimzy/kivik/auth/cookie"
	"github.com/flimzy/kivik/authdb/confadmin"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/proxy"
	"github.com/flimzy/kivik/serve"
	"github.com/flimzy/kivik/serve/conf"
	"github.com/spf13/viper"
)

type customDriver struct {
	driver.Client
}

func (cd customDriver) NewClient(_ context.Context, _ string) (driver.Client, error) {
	return cd, nil
}

func TestServer(t *testing.T) {
	memClient, _ := kivik.New(context.Background(), "memory", "")
	kivik.Register("custom", customDriver{proxy.NewClient(memClient)})
	backend, err := kivik.New(context.Background(), "custom", "")
	if err != nil {
		t.Fatalf("Failed to connect to custom driver: %s", err)
	}
	c := &conf.Conf{Viper: viper.New()}
	// Set admin/abc123 credentials
	c.Set("admins.admin", "-pbkdf2-792221164f257de22ad72a8e94760388233e5714,7897f3451f59da741c87ec5f10fe7abe,10")
	service := serve.Service{}
	service.Config = c
	service.Client = backend
	service.UserStore = confadmin.New(c)
	service.AuthHandlers = []auth.Handler{
		&basic.HTTPBasicAuth{},
		&cookie.Auth{},
	}
	handler, err := service.Init()
	if err != nil {
		t.Fatalf("Failed to initialize server: %s\n", err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()

	dsn, _ := url.Parse(server.URL)
	dsn.User = url.UserPassword("admin", "abc123")
	clients, err := connectClients("couch", dsn.String(), t)
	if err != nil {
		t.Fatalf("Failed to initialize client: %s", err)
	}
	clients.RW = true
	runTests(clients, SuiteKivikServer, t)
}
