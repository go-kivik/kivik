package test

import (
	"testing"

	"github.com/flimzy/kivik"
)

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
