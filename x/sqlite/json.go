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
	rev int
	id  string
}

func (r revision) String() string {
	if r.rev == 0 {
		return ""
	}
	return strconv.Itoa(r.rev) + "-" + r.id
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
		return revision{rev: int(id)}, nil
	}
	return revision{rev: int(id), id: parts[1]}, nil
}

// docData represents the document id, rev, deleted status, etc.
type docData struct {
	ID      string
	Rev     string
	Doc     []byte
	Deleted bool
}

// prepareDoc prepares the doc for insertion. It returns the new docID, rev, and
// marshaled doc with rev and id removed.
func prepareDoc(docID string, doc interface{}) (*docData, error) {
	tmpJSON, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	var tmp map[string]interface{}
	if err := json.Unmarshal(tmpJSON, &tmp); err != nil {
		return nil, err
	}
	data := &docData{
		ID: docID,
	}
	delete(tmp, "_rev")
	if id, ok := tmp["_id"].(string); ok {
		if docID != "" && id != docID {
			return nil, &internal.Error{Status: http.StatusBadRequest, Message: "Document ID must match _id in document"}
		}
		data.ID = id
		delete(tmp, "_id")
	}
	switch deleted := tmp["_deleted"].(type) {
	case bool:
		data.Deleted = deleted
		if !deleted {
			delete(tmp, "_deleted")
		}
	case nil:
	default:
		return nil, &internal.Error{Status: http.StatusBadRequest, Message: "_deleted must be a boolean"}
	}

	h := md5.New()
	b, _ := json.Marshal(tmp)
	if _, err := io.Copy(h, bytes.NewReader(b)); err != nil {
		return nil, err
	}
	data.Rev = hex.EncodeToString(h.Sum(nil))
	data.Doc = b
	return data, nil
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

func mergeIntoDoc(doc []byte, partials ...map[string]interface{}) ([]byte, error) {
	var merged map[string]interface{}
	if err := json.Unmarshal(doc, &merged); err != nil {
		return nil, err
	}
	for _, p := range partials {
		for k, v := range p {
			merged[k] = v
		}
	}
	return json.Marshal(merged)
}
