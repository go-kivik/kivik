package couchdb

import (
	"net/url"
	"os"
	"testing"
)

func dsn(t *testing.T) string {
	for _, env := range []string{"KIVIK_TEST_DSN_COUCH16", "KIVIK_TEST_DSN_COUCH20", "KIVIK_TEST_DSN_CLOUDANT"} {
		dsn := os.Getenv(env)
		if dsn != "" {
			return dsn
		}
	}
	t.Skip("DSN not set")
	return ""
}

func noAuthDSN(t *testing.T) string {
	dsn := dsn(t)
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("invalid DSN: %s", err)
	}
	parsed.User = nil
	return parsed.String()
}

func connect(dsn string, t *testing.T) *client {
	couch := &Couch{}
	driverClient, err := couch.NewClient(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to '%s': %s", dsn, err)
	}
	return driverClient.(*client)
}

func getClient(t *testing.T) *client {
	return connect(dsn(t), t)
}

func getNoAuthClient(t *testing.T) *client {
	return connect(noAuthDSN(t), t)
}
