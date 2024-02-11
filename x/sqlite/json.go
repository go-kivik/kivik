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

type revision struct {
	id  int
	rev string
}

func (r revision) String() string {
	return strconv.Itoa(r.id) + "-" + r.rev
}

func parseRev(s string) (revision, error) {
	if s == "" {
		return revision{}, &internal.Error{Status: http.StatusBadRequest, Message: "missing _rev"}
	}
	const revElements = 2
	parts := strings.SplitN(s, "-", revElements)
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return revision{}, &internal.Error{Status: http.StatusBadRequest, Err: err}
	}
	if len(parts) == 1 {
		// A rev that contains only a number is technically valid.
		return revision{id: int(id)}, nil
	}
	return revision{id: int(id), rev: parts[1]}, nil
}

// prepareDoc prepares the doc for insertion. It returns the new docID, rev, and
// marshaled doc with rev and id removed.
func prepareDoc(docID string, doc interface{}) (string, string, []byte, error) {
	tmpJSON, err := json.Marshal(doc)
	if err != nil {
		return "", "", nil, err
	}
	var tmp map[string]interface{}
	if err := json.Unmarshal(tmpJSON, &tmp); err != nil {
		return "", "", nil, err
	}
	delete(tmp, "_rev")
	if id, ok := tmp["_id"].(string); ok {
		if docID != "" && id != docID {
			return "", "", nil, &internal.Error{Status: http.StatusBadRequest, Message: "Document ID must match _id in document"}
		}
		docID = id
		delete(tmp, "_id")
	}
	h := md5.New()
	b, _ := json.Marshal(tmp)
	if _, err := io.Copy(h, bytes.NewReader(b)); err != nil {
		return "", "", nil, err
	}
	return docID, hex.EncodeToString(h.Sum(nil)), b, nil
}

// extractRev extracts the rev from the document.
func extractRev(doc interface{}) (string, error) {
	switch t := doc.(type) {
	case map[string]interface{}:
		r, _ := t["_rev"].(string)
		return r, nil
	case map[string]string:
		return t["_rev"], nil
	default:
		tmpJSON, err := json.Marshal(doc)
		if err != nil {
			return "", &internal.Error{Status: http.StatusBadRequest, Err: err}
		}
		var revDoc struct {
			Rev string `json:"_rev"`
		}
		if err := json.Unmarshal(tmpJSON, &revDoc); err != nil {
			return "", &internal.Error{Status: http.StatusBadRequest, Err: err}
		}
		return revDoc.Rev, nil
	}
}
