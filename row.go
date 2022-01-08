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

package kivik

import (
	"encoding/json"
	"io"
	"net/http"
	"sync/atomic"
)

// Row contains the result of calling Get for a single document. For most uses,
// it is sufficient just to call the ScanDoc method. For more advanced uses, the
// fields may be accessed directly.
type Row struct {
	// Rev is the revision ID of the returned document.
	Rev string

	// Body represents the document's content.
	//
	// Kivik will always return a non-nil Body, except when Err is non-nil. The
	// ScanDoc method will close Body. When not using ScanDoc, it is the
	// caller's responsibility to close Body
	Body io.ReadCloser

	// Err contains any error that occurred while fetching the document. It is
	// typically returned by ScanDoc.
	Err error

	// Attachments is experimental
	Attachments *AttachmentsIterator
}

// ScanDoc unmarshals the data from the fetched row into dest. It is an
// intelligent wrapper around json.Unmarshal which also handles
// multipart/related responses. When done, the underlying reader is closed.
func (r *Row) ScanDoc(dest interface{}) error {
	if r.Err != nil {
		return r.Err
	}
	defer r.Body.Close() // nolint:errcheck
	return json.NewDecoder(r.Body).Decode(dest)
}

type row struct {
	// prepared is set to true by the first call to Next()
	prepared int32
	closed   int32
	baseRows
	id string // TODO
	*Row
}

var _ ResultSet = &row{}

func (r *row) Close() error {
	atomic.StoreInt32(&r.closed, 1)
	return nil
}

func (r *row) Finish() (ResultMetadata, error) {
	return ResultMetadata{}, r.Close()
}

func (r *row) Err() error  { return r.Row.Err }
func (r *row) ID() string  { return r.id }
func (r *row) Rev() string { return r.Row.Rev }
func (r *row) Key() string { return "" }
func (r *row) ScanKey(interface{}) error {
	if atomic.LoadInt32(&r.closed) == 1 {
		return &Error{HTTPStatus: http.StatusBadRequest, Message: "kivik: Iterator is closed"}
	}

	return r.Row.Err
}

func (r *row) ScanValue(interface{}) error {
	if atomic.LoadInt32(&r.closed) == 1 {
		return &Error{HTTPStatus: http.StatusBadRequest, Message: "kivik: Iterator is closed"}
	}
	return r.Row.Err
}

func (r *row) Next() bool {
	if r.Row.Err != nil {
		return false
	}
	if atomic.SwapInt32(&r.prepared, 1) == 1 {
		return false
	}
	return true
}

func (r *row) ScanAllDocs(dest interface{}) error {
	if atomic.LoadInt32(&r.closed) == 1 {
		return &Error{HTTPStatus: http.StatusBadRequest, Message: "kivik: Iterator is closed"}
	}
	return scanAllDocs(r, dest)
}

func (r *row) ScanDoc(dest interface{}) error {
	if r.Row.Err != nil {
		return r.Row.Err
	}
	atomic.StoreInt32(&r.closed, 1)
	return r.Row.ScanDoc(dest)
}

func (r *row) Attachments() *AttachmentsIterator {
	return r.Row.Attachments
}
