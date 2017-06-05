package test

import (
	"context"
	"testing"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/memory"
	"github.com/flimzy/kivik/test/kt"
)

func TestMemory(t *testing.T) {
	client, err := kivik.New(context.Background(), "memory", "")
	if err != nil {
		t.Errorf("Failed to connect to memory driver: %s\n", err)
		return
	}
	clients := &kt.Context{
		RW:    true,
		Admin: client,
	}
	if err := client.CreateDB(context.Background(), "_users"); err != nil {
		t.Fatal(err)
	}
	runTests(clients, SuiteKivikMemory, t)
}
