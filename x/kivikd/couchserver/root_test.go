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

package couchserver

import (
	"net/http/httptest"
	"testing"

	"gitlab.com/flimzy/testy"
)

func TestGetRoot(t *testing.T) {
	h := Handler{
		CompatVersion: "1.6.1",
		Vendor:        "Acme",
		VendorVersion: "10.0",
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	handler := h.GetRoot()
	handler(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	expected := map[string]interface{}{
		"couchdb": "Välkommen",
		"version": "1.6.1",
		"vendor": map[string]string{
			"version": "10.0",
			"name":    "Acme",
		},
	}
	if d := testy.DiffAsJSON(expected, resp.Body); d != nil {
		t.Error(d)
	}
}
