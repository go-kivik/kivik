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
	"sync"
	"sync/atomic"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/x/fsdb/cdb"
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
	docs, err := d.cdb.OpenDocIDOpenRevs(docID, options)
	if err != nil {
		return nil, err
	}
	docs[0].Options = opts
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(docs[0]); err != nil {
		return nil, err
	}

	switch opts["open_revs"].(type) {
	case string, []string:
		return &documentRevs{
			ID:        docID,
			mu:        sync.Mutex{},
			Revisions: docs[0].Revisions,
			Body:      buf,
		}, nil
	}
	attsIter, err := docs[0].Revisions[0].AttachmentsIterator()
	if err != nil {
		return nil, err
	}
	return &document{
		ID:          docID,
		Rev:         docs[0].Revisions[0].Rev.String(),
		Body:        io.NopCloser(buf),
		Attachments: attsIter,
	}, nil
}

type documentRevs struct {
	ID        string
	mu        sync.Mutex
	Revisions cdb.Revisions
	Body      io.Reader
}

var _ driver.Rows = (*documentRevs)(nil)

func (d *documentRevs) Next(row *driver.Row) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.Revisions) == 0 {
		return io.EOF
	}
	curRev := d.Revisions[0]
	d.Revisions = d.Revisions[1:]

	row.ID = d.ID
	row.Rev = curRev.Rev.String()
	row.Doc = d.Body

	return nil
}

func (*documentRevs) Close() error {
	return nil
}
func (*documentRevs) Offset() int64     { return 0 }
func (*documentRevs) TotalRows() int64  { return 0 }
func (*documentRevs) UpdateSeq() string { return "" }

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
	row.ID = d.ID
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
