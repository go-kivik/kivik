// +build !js

// JS disabled until net/http polyfill can be figured out. See https://github.com/gopherjs/gopherjs/issues/586

package test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/couchdb"
	_ "github.com/flimzy/kivik/driver/fs"
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

func TestFS(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "kivik.test.")
	if err != nil {
		t.Errorf("Failed to create temp dir to test FS driver: %s\n", err)
		return
	}
	os.RemoveAll(tempDir)       // So the driver can re-create it as desired
	defer os.RemoveAll(tempDir) // To clean up after tests
	client, err := kivik.New("fs", tempDir)
	if err != nil {
		t.Errorf("Failed to connect to FS driver: %s\n", err)
		return
	}
	RunSubtests(client, true, []string{SuiteKivikFS}, t)
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
