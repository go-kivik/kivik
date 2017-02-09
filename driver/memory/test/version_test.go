package test

import (
	"reflect"
	"testing"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/memory"
)

const TestServer = "memdb"
const ExpectedVersion = "0.0.1"

func TestVersion(t *testing.T) {
	s, err := kivik.New("memory", TestServer)
	if err != nil {
		t.Fatalf("Error connecting to %s: %s\n", TestServer, err)
	}
	version, err := s.Version()
	if err != nil {
		t.Fatalf("Failed to get server info: %s", err)
	}
	if ExpectedVersion != version {
		t.Errorf("Server version.\n\tExpected: %s\n\t  Actual: %s\n", ExpectedVersion, version)
	}
}

var ExpectedAllDBs = []string{}

func TestAllDBs(t *testing.T) {
	s, err := kivik.New("memory", TestServer)
	if err != nil {
		t.Fatalf("Error connecting to %s: %s\n", TestServer, err)
	}
	allDBs, err := s.AllDBs()
	if err != nil {
		t.Fatalf("Failed to get all DBs: %s", err)
	}
	if !reflect.DeepEqual(ExpectedAllDBs, allDBs) {
		t.Errorf("All DBs.\n\tExpected: %v\n\t  Actual: %v\n", ExpectedAllDBs, allDBs)
	}
}

func TestUUIDs(t *testing.T) {
	s, err := kivik.New("memory", TestServer)
	if err != nil {
		t.Fatalf("Error connecting to %s: %s\n", TestServer, err)
	}
	uuidCount := 3
	uuids, err := s.UUIDs(uuidCount)
	if err != nil {
		t.Fatalf("Failed to get UUIDs: %s", err)
	}
	if len(uuids) != uuidCount {
		t.Errorf("Expected %d UUIDs, got %d\n", uuidCount, len(uuids))
	}
}
