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

package cmd

import (
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"
)

func checkRequest(t *testing.T, req *http.Request) {
	t.Helper()
	t.Cleanup(func() {
		_ = req.Body.Close()
	})

	if d := testy.DiffHTTPRequest(testy.Snapshot(t), req, standardReplacements...); d != nil {
		t.Error(d)
	}
}
