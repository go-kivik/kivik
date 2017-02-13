package test

import (
	"os"
	"testing"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/couchdb"
	_ "github.com/flimzy/kivik/driver/memory"
)

func TestMemory(t *testing.T) {
	client, err := kivik.New("memory", "")
	if err != nil {
		t.Errorf("Failed to connect to memory driver: %s\n", err)
		return
	}
	RunSubtests(client, true, []string{SuiteKivikMemory}, t)
}

func TestCloudant(t *testing.T) {
	dsn := os.Getenv("KIVIK_CLOUDANT_DSN")
	if dsn == "" {
		t.Skip("KIVIK_CLOUDANT_DSN: Cloudant DSN not set; skipping tests")
	}
	client, err := kivik.New("couch", dsn)
	if err != nil {
		t.Errorf("Failed to connect to cloudant: %s\n", err)
		return
	}
	RunSubtests(client, true, []string{SuiteCloudant}, t)
}

func TestCouch16(t *testing.T) {
	dsn := os.Getenv("KIVIK_COUCH16_DSN")
	if dsn == "" {
		t.Skip("KIVIK_COUCH16_DSN: Couch 1.6 DSN not set; skipping tests")
	}
	client, err := kivik.New("couch", dsn)
	if err != nil {
		t.Errorf("Failed to connect to CouchDB 1.6: %s\n", err)
		return
	}
	RunSubtests(client, true, []string{SuiteCouch}, t)
}

func TestCouch20(t *testing.T) {
	dsn := os.Getenv("KIVIK_COUCH20_DSN")
	if dsn == "" {
		t.Skip("KIVIK_COUCH16_DSN: Couch 2.0 DSN not set; skipping tests")
	}
	client, err := kivik.New("couch", dsn)
	if err != nil {
		t.Errorf("Failed to connect to CouchDB 2.0: %s\n", err)
		return
	}
	RunSubtests(client, true, []string{SuiteCouch20}, t)
}
