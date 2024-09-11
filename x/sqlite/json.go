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
	"database/sql/driver"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/x/mango"
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

func (r revision) IsZero() bool {
	return r.rev == 0
}

func parseRev(s string) (revision, error) {
	if s == "" {
		return revision{}, &internal.Error{Status: http.StatusBadRequest, Message: "missing _rev"}
	}
	const revElements = 2
	parts := strings.SplitN(s, "-", revElements)
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return revision{}, &internal.Error{Status: http.StatusBadRequest, Message: "invalid rev format"}
	}
	if len(parts) == 1 {
		// A rev that contains only a number is technically valid.
		return revision{rev: int(id)}, nil
	}
	return revision{rev: int(id), id: parts[1]}, nil
}

// docData represents the document id, rev, deleted status, etc.
type docData struct {
	ID                 string                `json:"_id"`
	Revisions          revsInfo              `json:"_revisions"`
	Deleted            bool                  `json:"_deleted"`
	Attachments        map[string]attachment `json:"_attachments"`
	RemovedAttachments []string              `json:"-"`
	Doc                []byte
	// MD5sum is the MD5sum of the document data. It, along with a hash of
	// attachment metadata, is used to calculate the document revision.
	MD5sum       md5sum        `json:"-"`
	DesignFields designDocData `json:"-"`
}

func (d *docData) IsDesignDoc() bool {
	return strings.HasPrefix(d.ID, "_design/")
}

type views struct {
	Map    string `json:"map"`
	Reduce string `json:"reduce,omitempty"`
}

type designDocViewOptions struct {
	// LocalSeq makes documents' local sequence numbers available to map
	// functions (as a `_local_seq` document property). See
	// https://docs.couchdb.org/en/stable/api/ddoc/views.html#view-options
	LocalSeq bool `json:"local_seq,omitempty"`
	// IncludeDesign allows map functions to be called on design documents as
	// well as regular documents. See
	// https://docs.couchdb.org/en/stable/api/ddoc/views.html#view-options
	IncludeDesign bool `json:"include_design,omitempty"`
	// Collation allows specifying an alternative collation sequence. Supported
	// options are:
	//
	//	- nil: CouchDB's default collation sequence.
	//	- "ascii": ASCII order, JSON values are sorted using their ASCII order.
	//	  This value is not supported by CouchDB.
	//	- "raw": Same as ASCII for now, but may change to match CouchDB's `raw`
	//	  collation in the future.
	//
	// See https://docs.couchdb.org/en/stable/ddocs/views/collation.html#raw-collation
	Collation *string `json:"collation,omitempty"`
}

// designDocData represents a design document. See
// https://docs.couchdb.org/en/stable/ddocs/ddocs.html#creation-and-structure
type designDocData struct {
	Language           string            `json:"language,omitempty"`
	Views              map[string]views  `json:"views,omitempty"`
	Updates            map[string]string `json:"updates,omitempty"`
	Filters            map[string]string `json:"filters,omitempty"`
	ValidateDocUpdates string            `json:"validate_doc_update,omitempty"`
	// AutoUpdate indicates whether to automatically build indexes defined in
	// this design document. Default is true.
	AutoUpdate *bool                `json:"autoupdate,omitempty"`
	Options    designDocViewOptions `json:"options,omitempty"`
}

// RevID returns calculated revision ID, possibly setting the MD5sum if it is
// not already set.
func (d *docData) RevID() string {
	if d.MD5sum.IsZero() {
		if len(d.Doc) == 0 {
			panic("MD5sum not set")
		}
		h := md5.New()
		_, _ = io.Copy(h, bytes.NewReader(d.Doc))
		copy(d.MD5sum[:], h.Sum(nil))
	}
	// The revision ID is a hash of:
	// - The MD5sum of the document data
	// - filenames and digests of attachments sorted by filename
	// - the deleted flag, if true
	h := md5.New()
	_, _ = h.Write(d.MD5sum[:])
	if len(d.Attachments) > 0 {
		filenames := make([]string, 0, len(d.Attachments))
		for filename := range d.Attachments {
			filenames = append(filenames, filename)
		}
		sort.Strings(filenames)
		for _, filename := range filenames {
			_, _ = h.Write(d.Attachments[filename].Digest.Bytes())
			_, _ = h.Write([]byte(filename))
			_, _ = h.Write([]byte{0})
		}
	}
	if d.Deleted {
		_, _ = h.Write([]byte{0xff})
	}
	return hex.EncodeToString(h.Sum(nil))
}

const md5sumLen = 16

type md5sum [md5sumLen]byte

func parseMD5sum(s string) (md5sum, error) {
	x, err := hex.DecodeString(s)
	if err != nil {
		return md5sum{}, err
	}
	var m md5sum
	copy(m[:], x)
	return m, nil
}

func parseDigest(s string) (md5sum, error) {
	if !strings.HasPrefix(s, "md5-") {
		return md5sum{}, fmt.Errorf("invalid digest: %s", s)
	}
	x, err := base64.StdEncoding.DecodeString(s[4:])
	if err != nil {
		return md5sum{}, err
	}
	var m md5sum
	copy(m[:], x)
	return m, nil
}

func (m md5sum) IsZero() bool {
	return m == md5sum{}
}

func (m md5sum) String() string {
	return hex.EncodeToString(m[:])
}

func (m md5sum) Value() (driver.Value, error) {
	if m.IsZero() {
		panic("zero value")
	}
	return m[:], nil
}

func (m *md5sum) Scan(src interface{}) error {
	x, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("unsupported type: %T", src)
	}
	if len(x) != md5sumLen {
		return fmt.Errorf("invalid length: %d", len(x))
	}

	copy(m[:], x)
	return nil
}

func (m md5sum) Digest() string {
	s, _ := m.MarshalText()
	return string(s)
}

func (m md5sum) MarshalText() ([]byte, error) {
	const prefix = "md5-"
	enc := base64.StdEncoding
	b := make([]byte, len(prefix)+enc.EncodedLen(md5sumLen))
	copy(b, "md5-")
	enc.Encode(b[4:], m[:])
	return b, nil
}

func (m md5sum) Bytes() []byte {
	return m[:]
}

type revsInfo struct {
	Start int      `json:"start"`
	IDs   []string `json:"ids"`
}

type attachment struct {
	ContentType string `json:"content_type"`
	Digest      md5sum `json:"digest"`
	Length      int64  `json:"length"`
	RevPos      int    `json:"revpos"`
	Stub        bool   `json:"stub,omitempty"`
	// TODO: Add encoding support to compress certain types of attachments.
	// Encoding      string `json:"encoding"`
	// EncodedLength int64  `json:"encoded_length"`

	// Data is the raw JSON representation of the attachment data. It is decoded
	// into Content by the [attachment.calculate] method.
	Data    json.RawMessage `json:"data,omitempty"`
	Content []byte          `json:"-"`
}

func (a *attachment) MarshalJSON() ([]byte, error) {
	alias := struct {
		attachment
		Stub        bool            `json:"stub,omitempty"`
		Data        json.RawMessage `json:"data,omitempty"`
		MarshalJSON struct{}        `json:"-"`
	}{
		attachment: *a,
	}
	if a.Stub || len(a.Data) == 0 {
		alias.Data = nil
		alias.Stub = true
	} else {
		alias.Data = a.Data
		alias.Stub = false
	}
	return json.Marshal(alias)
}

// calculate calculates the length, digest, and content of the attachment.
func (a *attachment) calculate(filename string) error {
	if a.Data == nil && len(a.Content) == 0 {
		return &internal.Error{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid attachment data for %q", filename)}
	}
	if len(a.Content) == 0 {
		if err := json.Unmarshal(a.Data, &a.Content); err != nil {
			return &internal.Error{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid attachment data for %q: %w", filename, err)}
		}
	}
	a.Length = int64(len(a.Content))
	h := md5.New()
	if _, err := io.Copy(h, bytes.NewReader(a.Content)); err != nil {
		return err
	}
	copy(a.Digest[:], h.Sum(nil))
	return nil
}

// revs returns the revision list in oldest first order.
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
		return nil, &internal.Error{Status: http.StatusBadRequest, Err: err}
	}
	var ddocData designDocData
	if strings.HasPrefix(docID, "_design/") {
		if err := json.Unmarshal(tmpJSON, &ddocData); err != nil {
			return nil, &internal.Error{Status: http.StatusBadRequest, Err: err}
		}
		if ddocData.Language == "" {
			ddocData.Language = "javascript"
		}
		if ddocData.AutoUpdate == nil {
			ddocData.AutoUpdate = &[]bool{true}[0]
		}
	}
	var tmp map[string]interface{}
	if err := json.Unmarshal(tmpJSON, &tmp); err != nil {
		return nil, err
	}
	data := &docData{
		DesignFields: ddocData,
	}
	if err := json.Unmarshal(tmpJSON, &data); err != nil {
		return nil, &internal.Error{Status: http.StatusBadRequest, Err: err}
	}
	for key := range tmp {
		if strings.HasPrefix(key, "_") {
			delete(tmp, key)
		}
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
	sum := h.Sum(nil)
	data.Doc = b
	copy(data.MD5sum[:], sum)
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

type fullDoc struct {
	ID               string                 `json:"-"`
	Rev              string                 `json:"-"`
	Doc              json.RawMessage        `json:"-"`
	Conflicts        []string               `json:"_conflicts,omitempty"`
	DeletedConflicts []string               `json:"_deleted_conflicts,omitempty"`
	RevsInfo         []map[string]string    `json:"_revs_info,omitempty"`
	Revisions        *revsInfo              `json:"_revisions,omitempty"`
	LocalSeq         int                    `json:"_local_seq,omitempty"`
	Attachments      map[string]*attachment `json:"_attachments,omitempty"`
	Deleted          bool                   `json:"_deleted,omitempty"`
}

// hasMeta returns true if d contains metadata (_* fields other than _id and
// _rev) that would be included when marshaling.
func (d fullDoc) hasMeta() bool {
	return len(d.Conflicts) > 0 ||
		len(d.DeletedConflicts) > 0 ||
		len(d.RevsInfo) > 0 ||
		d.Revisions != nil ||
		d.LocalSeq > 0 ||
		len(d.Attachments) > 0 ||
		d.Deleted
}

func (d fullDoc) rev() (revision, error) {
	return parseRev(d.Rev)
}

// renders only the requested fields. If fields is is empty,
// the entire document is returned.
func (d *fullDoc) toRaw(fields ...string) json.RawMessage {
	want := func(string) bool { return true }
	if len(fields) > 0 {
		wantFields := make(map[string]struct{}, len(fields))
		for _, f := range fields {
			wantFields[f] = struct{}{}
		}
		want = func(field string) bool {
			_, ok := wantFields[field]
			return ok
		}
	}
	buf := bytes.Buffer{}
	_ = buf.WriteByte('{')
	if want("_id") {
		if id := d.ID; id != "" {
			_, _ = buf.WriteString(`"_id":`)
			_, _ = buf.Write(jsonMarshal(id))
			_ = buf.WriteByte(',')
		}
	}
	if want("_rev") {
		if rev := d.Rev; rev != "" {
			_, _ = buf.WriteString(`"_rev":`)
			_, _ = buf.Write(jsonMarshal(rev))
			_ = buf.WriteByte(',')
		}
	}

	const minJSONObjectLen = 2
	if len(d.Doc) > minJSONObjectLen {
		if len(fields) == 0 {
			// The main doc
			_, _ = buf.Write(d.Doc[1 : len(d.Doc)-1]) // Omit opening and closing braces
			_ = buf.WriteByte(',')
		} else {
			var doc map[string]interface{}
			tmp := map[string]interface{}{}
			_ = json.Unmarshal(d.Doc, &doc)
			for _, f := range fields {
				if f == "_id" || f == "_rev" {
					continue
				}
				keys := mango.SplitKeys(f)
				if v, ok := extractValue(doc, keys...); ok {
					insertValue(tmp, v, keys)
				}
			}
			if len(tmp) > 0 {
				jsonText, _ := json.Marshal(tmp)
				_, _ = buf.Write(jsonText[1 : len(jsonText)-1])
				_ = buf.WriteByte(',')
			}
		}
	}

	if d.hasMeta() {
		if len(fields) == 0 {
			tmp, _ := json.Marshal(d)
			_, _ = buf.Write(tmp[1 : len(tmp)-1])
			_ = buf.WriteByte(',')
		} else {
			// These are the only two meta fields supported by the _find endpoint
			if want("_conflicts") && len(d.Conflicts) > 0 {
				_, _ = buf.WriteString(`"_conflicts":`)
				_, _ = buf.Write(jsonMarshal(d.Conflicts))
				_ = buf.WriteByte(',')
			}
			if want("_attachments") && len(d.Attachments) > 0 {
				_, _ = buf.WriteString(`"_attachments":`)
				_, _ = buf.Write(jsonMarshal(d.Attachments))
				_ = buf.WriteByte(',')
			}
		}
	}

	result := buf.Bytes()
	// replace final ',' with '}'
	result[len(result)-1] = '}'
	return result
}

func extractValue(obj map[string]interface{}, keys ...string) (interface{}, bool) {
	for i, key := range keys {
		if i == len(keys)-1 {
			o, ok := obj[key]
			return o, ok
		}
		if v, ok := obj[key]; ok {
			if m, ok := v.(map[string]interface{}); ok {
				obj = m
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	}
	return nil, false
}

func insertValue(obj map[string]interface{}, value interface{}, keys []string) {
	for i, key := range keys {
		if i == len(keys)-1 {
			obj[key] = value
			return
		}
		if v, ok := obj[key]; ok {
			m, ok := v.(map[string]interface{})
			if !ok {
				return
			}

			obj = m
		} else {
			m := make(map[string]interface{}, 1)
			obj[key] = m
			obj = m
		}
	}
}

func (d *fullDoc) toReader(fields ...string) io.ReadCloser {
	return io.NopCloser(bytes.NewReader(d.toRaw(fields...)))
}

func jsonMarshal(s interface{}) []byte {
	j, _ := json.Marshal(s)
	return j
}

// toMap produces a map from the fullDoc. It only considers a few of the
// meta fields; those used by map/reduce functions:
//
// - _id
// - _rev
// - _deleted
// - _attachments
func (d *fullDoc) toMap() map[string]interface{} {
	var result map[string]interface{}
	if err := json.Unmarshal(d.Doc, &result); err != nil {
		panic(err)
	}
	result["_id"] = d.ID
	result["_rev"] = d.Rev
	if d.Deleted {
		result["_deleted"] = true
	}
	if len(d.Attachments) > 0 {
		attachments := make(map[string]interface{}, len(d.Attachments))
		for name, att := range d.Attachments {
			attachments[name] = map[string]interface{}{
				"content_type": att.ContentType,
				"digest":       att.Digest.Digest(),
				"length":       att.Length,
				"revpos":       att.RevPos,
				"stub":         att.Stub,
			}
		}
		result["_attachments"] = attachments
	}
	if d.LocalSeq > 0 {
		result["_local_seq"] = d.LocalSeq
	}
	/*
		Conflicts        []string               `json:"_conflicts,omitempty"`
		DeletedConflicts []string               `json:"_deleted_conflicts,omitempty"`
		RevsInfo         []map[string]string    `json:"_revs_info,omitempty"`
		Revisions        *revsInfo              `json:"_revisions,omitempty"`
	*/
	return result
}
