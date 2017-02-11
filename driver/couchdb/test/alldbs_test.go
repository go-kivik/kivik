package test

import (
	"reflect"
	"testing"

	"github.com/flimzy/kivik"
)

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
