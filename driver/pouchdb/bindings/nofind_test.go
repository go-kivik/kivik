package bindings

import (
	"context"
	"testing"

	"github.com/flimzy/kivik"
	"github.com/gopherjs/gopherjs/js"
)

// TestNoFind tests that Find() properly returns NotImplemented when the
// pouchdb-find plugin is not loaded.
func TestNoFindPlugin(t *testing.T) {
	memdown := js.Global.Call("require", "memdown")
	t.Run("FindLoaded", func(t *testing.T) {
		db := GlobalPouchDB().New("foo", map[string]interface{}{"db": memdown})
		_, err := db.Find(context.Background(), "")
		if err == kivik.ErrNotImplemented {
			t.Errorf("Got ErrNotImplemented when pouchdb-find should be loaded")
		}
	})
	t.Run("FindNotLoaded", func(t *testing.T) {
		db := GlobalPouchDB().New("foo", map[string]interface{}{"db": memdown})
		db.Object.Set("find", nil) // Fake it
		_, err := db.Find(context.Background(), "")
		if err != kivik.ErrNotImplemented {
			t.Errorf("Expected %s error, got %s\n", kivik.ErrNotImplemented, err)
		}
	})
}
