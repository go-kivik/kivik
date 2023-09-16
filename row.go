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
	"sync/atomic"
)

type row struct {
	id   string
	rev  string
	body io.ReadCloser
	atts *AttachmentsIterator

	// prepared is set to true by the first call to Next()
	prepared int32
	err      error
}

var _ basicResultSet = &row{}

// ScanDoc unmarshals the data from the fetched row into dest. It is an
// intelligent wrapper around [encoding/json.Unmarshal] which also handles
// multipart/related responses. When done, the underlying reader is closed.
func (r *row) ScanDoc(dest interface{}) error {
	defer r.body.Close()
	return json.NewDecoder(r.body).Decode(dest)
}

func (r *row) Close() error {
	if _, err := io.Copy(io.Discard, r.body); err != nil {
		return err
	}
	r.err = r.body.Close()
	return r.err
}

func (r *row) Err() error           { return r.err }
func (r *row) ID() (string, error)  { return r.id, r.err }
func (r *row) Rev() (string, error) { return r.rev, r.err }

func (r *row) Next() bool {
	if r.err != nil {
		return false
	}
	return atomic.SwapInt32(&r.prepared, 1) != 1
}

func (r *row) Attachments() (*AttachmentsIterator, error) {
	return r.atts, r.err
}
