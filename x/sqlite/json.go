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
	"strconv"
	"strings"

	"github.com/go-kivik/kivik/v4/internal"
)

// prepareDoc prepares the doc for insertion. It returns the new rev and
// marshaled doc.
func prepareDoc(docID string, doc interface{}) (string, []byte, error) {
	tmpJSON, err := json.Marshal(doc)
	if err != nil {
		return "", nil, err
	}
	var tmp map[string]interface{}
	if err := json.Unmarshal(tmpJSON, &tmp); err != nil {
		return "", nil, err
	}
	delete(tmp, "_rev")
	if id, ok := tmp["_id"].(string); ok && docID != "" && id != docID {
		return "", nil, &internal.Error{Status: http.StatusBadRequest, Message: "Document ID must match _id in document"}
	}
	if docID != "" {
		tmp["_id"] = docID
	}
	h := md5.New()
	b, _ := json.Marshal(tmp)
	if _, err := io.Copy(h, bytes.NewReader(b)); err != nil {
		return "", nil, err
	}
	return hex.EncodeToString(h.Sum(nil)), b, nil
}

// extractRev extracts the rev from the document.
func extractRev(doc interface{}) (int, string, error) {
	var rev string
	switch t := doc.(type) {
	case map[string]interface{}:
		rev, _ = t["_rev"].(string)
	case map[string]string:
		rev = t["_rev"]
	default:
		tmpJSON, err := json.Marshal(doc)
		if err != nil {
			return 0, "", &internal.Error{Status: http.StatusBadRequest, Err: err}
		}
		var revDoc struct {
			Rev string `json:"_rev"`
		}
		if err := json.Unmarshal(tmpJSON, &revDoc); err != nil {
			return 0, "", &internal.Error{Status: http.StatusBadRequest, Err: err}
		}
		rev = revDoc.Rev
	}
	if rev == "" {
		return 0, "", &internal.Error{Status: http.StatusBadRequest, Message: "missing _rev"}
	}
	const revElements = 2
	parts := strings.SplitN(rev, "-", revElements)
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", &internal.Error{Status: http.StatusBadRequest, Err: err}
	}
	if len(parts) == 1 {
		// A rev that contains only a number is technically valid.
		return int(id), "", nil
	}
	return int(id), parts[1], nil
}
