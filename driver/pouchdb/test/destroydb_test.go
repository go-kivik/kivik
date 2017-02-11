package test

import (
	"testing"

	"github.com/flimzy/kivik"
)

func TestDestroyDB(t *testing.T) {
	s, err := kivik.New(TestDriver, TestServer)
	if err != nil {
		t.Fatalf("Error connecting to %s: %s\n", TestServer, err)
	}
	if err = s.CreateDB("nosuchdb"); err != nil {
		t.Errorf("Destroying a non-existent db should not error in PouchDB")
	}
}

func TestDestroyRemoteDB(t *testing.T) {
	s, err := kivik.New(TestDriver, RemoteServer)
	if err != nil {
		t.Fatalf("Error connecting to %s: %s\n", TestServer, err)
	}
	err = s.DestroyDB(RemoteServer + "/nosuchdb")
	if err == nil {
		t.Errorf("Destroying a non-existent db should error")
	}
}
