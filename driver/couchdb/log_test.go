package couchdb

import (
	"context"
	"testing"
)

func TestLog(t *testing.T) {
	ctx := context.Background()
	client := getClient(t)
	log, err := client.LogContext(ctx, 100, 0)
	if err != nil {
		t.Errorf("Failed to read log: %s", err)
	}
	log.Close()

	if _, err := client.LogContext(ctx, -100, 0); err == nil {
		t.Errorf("No error for invalid length argument")
	}
	if _, err := client.LogContext(ctx, 0, -100); err == nil {
		t.Errorf("No error for invalid offset argument")
	}
}
