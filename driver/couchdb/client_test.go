package couchdb

import (
	"encoding/hex"
	"testing"

	"github.com/flimzy/kivik/errors"
	"github.com/flimzy/kivik/test/kt"
)

func TestAllDBs(t *testing.T) {
	client := getClient(t)
	_, err := client.AllDBsContext(kt.CTX, nil)
	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
}

func TestUUIDs(t *testing.T) {
	client := getClient(t)
	uuids, err := client.UUIDsContext(kt.CTX, 5)
	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if len(uuids) != 5 {
		t.Errorf("Expected 5 UUIDs, got %d", len(uuids))
	}
	for _, v := range uuids {
		if err := validateUUID(v); err != nil {
			t.Errorf("Invalid UUID '%s': %s", v, err)
		}
	}
}

func validateUUID(uuid string) error {
	if len(uuid) != 32 {
		return errors.Errorf("UUID length is %d, expected 32", len(uuid))
	}
	_, err := hex.DecodeString(uuid)
	return err
}

func TestDBExists(t *testing.T) {
	client := getClient(t)
	exists, err := client.DBExistsContext(kt.CTX, "_users", nil)
	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if !exists {
		t.Error("Expected _users to exist")
	}
}

func TestCreateAndDestroyDB(t *testing.T) {
	client := getClient(t)
	dbName := kt.TestDBName(t)
	defer client.DestroyDBContext(kt.CTX, dbName, nil)
	if err := client.CreateDBContext(kt.CTX, dbName, nil); err != nil {
		t.Errorf("Create failed: %s", err)
	}
	if err := client.DestroyDBContext(kt.CTX, dbName, nil); err != nil {
		t.Errorf("Destroy failed: %s", err)
	}
}
