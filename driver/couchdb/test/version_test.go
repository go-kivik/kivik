package test

import (
	"reflect"
	"testing"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/couchdb"
)

const TestServer = "https://kivik.cloudant.com/"
const TestServerAuth = "https://kivik:K1v1k123@kivik.cloudant.com/"
const ExpectedVersion = "2.0.0"

func TestVersion(t *testing.T) {
	s, err := kivik.New("couch", TestServer)
	if err != nil {
		t.Fatalf("Error connecting to " + TestServer)
	}
	version, err := s.Version()
	if err != nil {
		t.Fatalf("Failed to get server info: %s", err)
	}
	if ExpectedVersion != version {
		t.Errorf("Server version.\n\tExpected: %s\n\t  Actual: %s\n", ExpectedVersion, version)
	}
}

var ExpectedAllDBs = []string{"_replicator", "_users"}

func TestAllDBs(t *testing.T) {
	s, err := kivik.New("couch", TestServerAuth)
	if err != nil {
		t.Fatalf("Error connecting to " + TestServerAuth)
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
	s, err := kivik.New("couch", TestServer)
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

func TestMembership(t *testing.T) {
	s, err := kivik.New("couch", TestServerAuth)
	if err != nil {
		t.Fatalf("Error connecting to %s: %s\n", TestServerAuth, err)
	}
	all, cluster, err := s.Membership()
	if err != nil {
		t.Fatalf("Failed to get Membership: %s", err)
	}
	if len(all) < 2 {
		t.Fatalf("Only got %d nodes, expected 2+\n", len(all))
	}
	if len(cluster) < 2 {
		t.Fatalf("Only got %d cluster nodes, expected 2+\n", len(cluster))
	}
}
