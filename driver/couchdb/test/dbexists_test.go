package test

import (
	"testing"

	"github.com/flimzy/kivik"
)

func TestDBExists(t *testing.T) {
	s, err := kivik.New(TestDriver, TestServerAuth)
	if err != nil {
		t.Fatalf("Error connecting to %s: %s\n", TestServerAuth, err)
	}
	exists, err := s.DBExists("bogusDB")
	if err != nil {
		t.Fatalf("Failed to check DB existence: %s", err)
	}
	if exists {
		t.Errorf("DB 'bogusDB' should not exist")
	}
	exists, err = s.DBExists("_users")
	if err != nil {
		t.Fatalf("Failed to check DB existence: %s", err)
	}
	if !exists {
		t.Errorf("DB '_users' should exist")
	}

}
