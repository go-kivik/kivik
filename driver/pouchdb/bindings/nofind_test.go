package bindings

import (
	"context"
	"testing"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
	"github.com/gopherjs/gopherjs/js"
)

// TestNoFind tests that Find() properly returns NotImplemented when the
// pouchdb-find plugin is not loaded.
func TestNoFindPlugin(t *testing.T) {
	memdown := js.Global.Call("require", "memdown")
	t.Run("FindLoaded", func(t *testing.T) {
		db := GlobalPouchDB().New("foo", map[string]interface{}{"db": memdown})
		_, err := db.Find(context.Background(), "")
		if errors.StatusCode(err) == kivik.StatusNotImplemented {
			t.Errorf("Got StatusNotImplemented when pouchdb-find should be loaded")
		}
	})
	t.Run("FindNotLoaded", func(t *testing.T) {
		db := GlobalPouchDB().New("foo", map[string]interface{}{"db": memdown})
		db.Object.Set("find", nil) // Fake it
		_, err := db.Find(context.Background(), "")
		if code := errors.StatusCode(err); code != kivik.StatusNotImplemented {
			t.Errorf("Expected %d error, got %d/%s\n", kivik.StatusNotImplemented, code, err)
		}
	})
}
