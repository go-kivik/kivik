package kt

import (
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/flimzy/kivik"
)

// DSN returns a testing DSN from the environment.
func DSN(t *testing.T) string {
	for _, env := range []string{"KIVIK_TEST_DSN_COUCH16", "KIVIK_TEST_DSN_COUCH20", "KIVIK_TEST_DSN_CLOUDANT"} {
		dsn := os.Getenv(env)
		if dsn != "" {
			return dsn
		}
	}
	t.Skip("DSN not set")
	return ""
}

// NoAuthDSN returns a testing DSN with credentials stripped.
func NoAuthDSN(t *testing.T) string {
	dsn := DSN(t)
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("invalid DSN: %s", err)
	}
	parsed.User = nil
	return parsed.String()
}

func connect(dsn string, t *testing.T) *kivik.Client {
	client, err := kivik.New(context.Background(), "couch", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to '%s': %s", dsn, err)
	}
	return client
}

// GetClient returns a connection to a CouchDB client, for testing.
func GetClient(t *testing.T) *kivik.Client {
	return connect(DSN(t), t)
}

// GetNoAuthClient returns an unauthenticated connection to a CouchDB client, for testing.
func GetNoAuthClient(t *testing.T) *kivik.Client {
	return connect(NoAuthDSN(t), t)
}
