package couchdb

import (
	"encoding/hex"
	"testing"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
	"github.com/flimzy/kivik/test/kt"
)

func TestAllDBs(t *testing.T) {
	client := getClient(t)
	_, err := client.AllDBsContext(CTX)
	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
}

func TestUUIDs(t *testing.T) {
	client := getClient(t)
	uuids, err := client.UUIDsContext(CTX, 5)
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

func TestMembership(t *testing.T) {
	client := getClient(t)
	_, _, err := client.MembershipContext(CTX)
	if err != nil && errors.StatusCode(err) != kivik.StatusNotImplemented {
		t.Errorf("Failed: %s", err)
	}
}

func TestDBExists(t *testing.T) {
	client := getClient(t)
	exists, err := client.DBExistsContext(CTX, "_users")
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
	defer client.DestroyDBContext(CTX, dbName)
	if err := client.CreateDBContext(CTX, dbName); err != nil {
		t.Errorf("Create failed: %s", err)
	}
	if err := client.DestroyDBContext(CTX, dbName); err != nil {
		t.Errorf("Destroy failed: %s", err)
	}
}
