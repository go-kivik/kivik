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
	"encoding/json"
	"net/http"
)

// GetRoot handles requests for: GET /
func (h *Handler) GetRoot() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		compatVer, vendName, vendVers := h.vendor()
		w.Header().Set("Content-Type", typeJSON)
		h.HandleError(w, json.NewEncoder(w).Encode(serverInfo{
			CouchDB: "Välkommen",
			Version: compatVer,
			Vendor: vendorInfo{
				Name:    vendName,
				Version: vendVers,
			},
		}))
	}
}
