package test

import (
	"testing"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/memory"
)

func TestMemory(t *testing.T) {
	client, err := kivik.New("memory", "")
	if err != nil {
		t.Errorf("Failed to connect to memory driver: %s", err)
		return
	}
	runSubTests(client, []string{SuiteKivikMemory}, t)
}
