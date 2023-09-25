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

package fs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync/atomic"

	"github.com/go-kivik/kivik/v4/driver"
)

// TODO:
// - atts_since
// - conflicts
// - deleted_conflicts
// - latest
// - local_seq
// - meta
// - open_revs
func (d *db) Get(_ context.Context, docID string, options driver.Options) (driver.Rows, error) {
	opts := map[string]interface{}{}
	options.Apply(opts)
	if docID == "" {
		return nil, statusError{status: http.StatusBadRequest, error: errors.New("no docid specified")}
	}
	doc, err := d.cdb.OpenDocID(docID, options)
	if err != nil {
		return nil, err
	}
	doc.Options = opts
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(doc); err != nil {
		return nil, err
	}
	attsIter, err := doc.Revisions[0].AttachmentsIterator()
	if err != nil {
		return nil, err
	}
	return &document{
		ID:          docID,
		Rev:         doc.Revisions[0].Rev.String(),
		Body:        io.NopCloser(buf),
		Attachments: attsIter,
	}, nil
}

type document struct {
	ID          string
	Rev         string
	Body        io.ReadCloser
	Attachments driver.Attachments

	// closed will be non-zero once the iterator is closed or the document
	// has been read.
	closed int32
}

func (d *document) Next(row *driver.Row) error {
	if atomic.SwapInt32(&d.closed, 1) != 0 {
		return io.EOF
	}
	row.Rev = d.ID
	row.Rev = d.Rev
	row.Doc = d.Body
	row.Attachments = d.Attachments
	return nil
}

func (d *document) Close() error {
	atomic.StoreInt32(&d.closed, 1)
	return nil
}

func (*document) UpdateSeq() string { return "" }
func (*document) Offset() int64     { return 0 }
func (*document) TotalRows() int64  { return 0 }
