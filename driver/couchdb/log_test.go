package couchdb

import "testing"

func TestLog(t *testing.T) {
	client := getClient(t)
	log, err := client.Log(100, 0)
	if err != nil {
		t.Errorf("Failed to read log: %s", err)
	}
	log.Close()

	if _, err := client.Log(-100, 0); err == nil {
		t.Errorf("No error for invalid length argument")
	}
	if _, err := client.Log(0, -100); err == nil {
		t.Errorf("No error for invalid offset argument")
	}
}
