package test

import (
	"testing"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/memory"
	"github.com/flimzy/kivik/test/kt"
)

func TestMemory(t *testing.T) {
	client, err := kivik.New("memory", "")
	if err != nil {
		t.Errorf("Failed to connect to memory driver: %s\n", err)
		return
	}
	clients := &kt.Context{
		Admin: client,
	}
	runTests(clients, SuiteKivikMemory, t)
}
