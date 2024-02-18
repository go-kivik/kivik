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
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	ID string `json:"_id"`
	// RevID is the calculated revision ID, not the actual _rev field from the
	// document.
	RevID       string                `json:"-"`
	Revisions   revsInfo              `json:"_revisions"`
	Deleted     bool                  `json:"_deleted"`
	Attachments map[string]attachment `json:"_attachments"`
	Doc         []byte
}

type revsInfo struct {
	Start int      `json:"start"`
	IDs   []string `json:"ids"`
}

type attachment struct {
	ContentType string `json:"content_type"`
	Digest      string `json:"digest"`
	Length      int64  `json:"length"`
	// Encoding      string `json:"encoding"`
	// RevPos        int    `json:"revpos"`
	// Stub          bool   `json:"stub"`
	// EncodedLength int64  `json:"encoded_length"`

	// Data is the raw JSON representation of the attachment data. It is decoded
	// into Content by the [attachment.calculate] method.
	Data    json.RawMessage `json:"data"`
	Content []byte          `json:"-"`
}

func (a *attachment) calculate(filename string) error {
	if a.Data == nil {
		return &internal.Error{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid attachment data for %q", filename)}
	}
	if err := json.Unmarshal(a.Data, &a.Content); err != nil {
		return &internal.Error{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid attachment data for %q: %w", filename, err)}
	}
	a.Length = int64(len(a.Content))
	h := md5.New()
	if _, err := io.Copy(h, bytes.NewReader(a.Content)); err != nil {
		return err
	}
	a.Digest = "md5-" + base64.StdEncoding.EncodeToString(h.Sum(nil))
	return nil
}

func (r *revsInfo) revs() []revision {
	revs := make([]revision, len(r.IDs))
	for i, id := range r.IDs {
		revs[len(r.IDs)-i-1] = revision{rev: r.Start - i, id: id}
	}
	return revs
}

// leaf returns the leaf revision of the revsInfo.
func (r *revsInfo) leaf() revision {
	return revision{rev: r.Start, id: r.IDs[0]}
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
	data := &docData{}
	if err := json.Unmarshal(tmpJSON, &data); err != nil {
		return nil, &internal.Error{Status: http.StatusBadRequest, Err: err}
	}
	if !data.Deleted {
		delete(tmp, "_deleted")
	}
	delete(tmp, "_rev")
	delete(tmp, "_revisions")
	if docID != "" && data.ID != "" && docID != data.ID {
		return nil, &internal.Error{Status: http.StatusBadRequest, Message: "Document ID must match _id in document"}
	}
	if data.ID == "" {
		data.ID = docID
	}

	h := md5.New()
	b, _ := json.Marshal(tmp)
	if _, err := io.Copy(h, bytes.NewReader(b)); err != nil {
		return nil, err
	}
	data.RevID = hex.EncodeToString(h.Sum(nil))
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
