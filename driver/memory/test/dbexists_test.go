package test

import (
	"testing"

	"github.com/flimzy/kivik"
)

func TestDBExists(t *testing.T) {
	s, err := kivik.New("memory", TestServer)
	if err != nil {
		t.Fatalf("Error connecting to %s: %s\n", TestServer, err)
	}
	exists, err := s.DBExists("bogusDB")
	if err != nil {
		t.Fatalf("Failed to check DB existence: %s", err)
	}
	if exists {
		t.Errorf("DB should not exist")
	}
	// FIXME: Uncomment this once it's possible to create a database
	// exists, err = s.DBExists("_users")
	// if err != nil {
	// 	t.Fatalf("Failed to check DB existence: %s", err)
	// }
	// if !exists {
	// 	t.Errorf("DB '_users' should exist")
	// }
}
