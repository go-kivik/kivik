// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

//go:build js
// +build js

package bindings

import (
	"context"
	"net/http"
	"testing"

	"github.com/gopherjs/gopherjs/js"

	kivik "github.com/go-kivik/kivik/v4"
)

func init() {
	memPouch := js.Global.Get("PouchDB").Call("defaults", map[string]interface{}{
		"db": js.Global.Call("require", "memdown"),
	})
	js.Global.Set("PouchDB", memPouch)
}

// TestNoFind tests that Find() properly returns NotImplemented when the
// pouchdb-find plugin is not loaded.
func TestNoFindPlugin(t *testing.T) {
	t.Run("FindLoaded", func(t *testing.T) {
		db := GlobalPouchDB().New("foo", nil)
		_, err := db.Find(context.Background(), "")
		if kivik.HTTPStatus(err) == http.StatusNotImplemented {
			t.Errorf("Got StatusNotImplemented when pouchdb-find should be loaded")
		}
	})
	t.Run("FindNotLoaded", func(t *testing.T) {
		db := GlobalPouchDB().New("foo", nil)
		db.Object.Set("find", nil) // Fake it
		_, err := db.Find(context.Background(), "")
		if code := kivik.HTTPStatus(err); code != http.StatusNotImplemented {
			t.Errorf("Expected %d error, got %d/%s\n", http.StatusNotImplemented, code, err)
		}
	})
}
