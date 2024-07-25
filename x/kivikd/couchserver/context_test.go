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

//go:build !js

package couchserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
)

func TestDB(t *testing.T) {
	router := chi.NewRouter()
	var result string
	router.Get("/{db}", func(_ http.ResponseWriter, r *http.Request) {
		result = DB(r)
	})
	req := httptest.NewRequest("GET", "/foo", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	expected := "foo"
	if result != expected {
		t.Errorf("Expected '%s', Got '%s'", expected, result)
	}
}
