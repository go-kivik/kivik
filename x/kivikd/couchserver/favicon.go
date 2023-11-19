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
	"bytes"
	"io"
	"net/http"
	"os"

	"github.com/go-kivik/kivik/v4/internal"
)

//go:generate go-bindata -pkg couchserver -nometadata -nocompress -prefix files -o files.go files

// GetFavicon serves GET /favicon.ico
func (h *Handler) GetFavicon() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ico io.Reader
		if h.Favicon == "" {
			asset, err := Asset("favicon.ico")
			if err != nil {
				panic(err)
			}
			ico = bytes.NewBuffer(asset)
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
