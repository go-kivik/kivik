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

	"github.com/go-kivik/kivik/v4"
	_ "github.com/go-kivik/kivik/v4/x/memorydb"
)

func TestAllDBs(t *testing.T) {
	client, err := kivik.New("memory", "")
	if err != nil {
		panic(err)
	}
	h := &Handler{client: &clientWrapper{client}}
	handler := h.GetAllDBs()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/_all_dbs", nil)
	handler(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	expected := []string{}
	if d := testy.DiffAsJSON(expected, resp.Body); d != nil {
		t.Error(d)
	}
}
