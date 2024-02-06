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

package sqlite

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-kivik/kivik/v4/internal"
)

func calculateRev(docID string, doc interface{}) (string, error) {
	tmpJSON, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}
	var tmp map[string]interface{}
	if err := json.Unmarshal(tmpJSON, &tmp); err != nil {
		return "", err
	}
	delete(tmp, "_rev")
	if id, ok := tmp["_id"].(string); ok && docID != "" && id != docID {
		return "", &internal.Error{Status: http.StatusBadRequest, Message: "Document ID must match _id in document"}
	}
	if docID != "" {
		tmp["_id"] = docID
	}
	h := md5.New()
	b, _ := json.Marshal(tmp)
	if _, err := io.Copy(h, bytes.NewReader(b)); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
