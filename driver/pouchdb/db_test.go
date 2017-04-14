package pouchdb

import (
	"testing"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
	"github.com/flimzy/kivik/test/kt"
	"github.com/gopherjs/gopherjs/js"
)

func init() {
	js.Global.Get("PouchDB").Call("defaults", map[string]interface{}{
		"db": js.Global.Call("require", "memdown"),
	})
}

func TestPut(t *testing.T) {
	client, err := kivik.New("pouch", "")
	if err != nil {
		t.Errorf("Failed to connect to PouchDB/memdown driver: %s", err)
		return
	}
	dbname := kt.TestDBName(t)
	defer client.DestroyDB(dbname)
	if err = client.CreateDB(dbname); err != nil {
		t.Fatalf("Failed to create db: %s", err)
	}
	db, err := client.DB(dbname)
	if err != nil {
		t.Fatalf("Failed to connect to db: %s", err)
	}
	_, err = db.Put("foo", map[string]string{"_id": "bar"})
	if errors.StatusCode(err) != kivik.StatusBadRequest {
		t.Errorf("Expected Bad Request for mismatched IDs, got %s", err)
	}
}
