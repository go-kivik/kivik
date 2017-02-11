package test

import (
	"testing"

	"github.com/flimzy/kivik"
)

func TestCreateDB(t *testing.T) {
	s, err := kivik.New(TestDriver, TestServerAuth)
	if err != nil {
		t.Fatalf("Error connecting to %s: %s\n", TestServerAuth, err)
	}
	if err = s.CreateDB("test"); err != nil {
		t.Errorf("Error creating test db: %s", err)
	}
	if err = s.DestroyDB("test"); err != nil {
		t.Errorf("Error destroying test db: %s", err)
	}
}
