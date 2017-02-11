package test

import (
	"testing"

	"github.com/flimzy/kivik"
)

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
