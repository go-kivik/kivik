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
	"embed"
	"io"
	"net/http"
	"os"

	"github.com/go-kivik/kivik/v4/internal"
)

//go:embed files
var files embed.FS

// GetFavicon serves GET /favicon.ico
func (h *Handler) GetFavicon() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ico io.Reader
		if h.Favicon == "" {
			asset, err := files.Open("files/favicon.ico")
			if err != nil {
				panic(err)
			}
			defer asset.Close()
			ico = asset
		} else {
			file, err := os.Open(h.Favicon)
			if err != nil {
				if os.IsNotExist(err) {
					err = &internal.Error{Status: http.StatusNotFound, Message: "not found"}
				}
				h.HandleError(w, err)
				return
			}
			ico = file
			defer file.Close()
		}
		w.Header().Set("Content-Type", "image/x-icon")
		_, err := io.Copy(w, ico)
		h.HandleError(w, err)
	}
}
