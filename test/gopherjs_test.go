// +build js

package test

import (
	"testing"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/pouchdb"
)

func TestPouch(t *testing.T) {
	client, err := kivik.New("pouch", "")
	if err != nil {
		t.Errorf("Failed to connect to PouchDB driver: %s", err)
		return
	}
	RunSubtests(client, []string{SuitePouch}, t)
}
