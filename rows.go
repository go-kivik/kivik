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
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal"
)

type rows struct {
	*iter
	rowsi driver.Rows
}

var _ fullResultSet = &rows{}

func newRows(ctx context.Context, onClose func(), rowsi driver.Rows) *rows {
	return &rows{
		iter:  newIterator(ctx, onClose, &rowsIterator{Rows: rowsi}, &driver.Row{}),
		rowsi: rowsi,
	}
}

func (r *rows) NextResultSet() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.err != nil {
		return false
	}
	if r.state == stateClosed {
		return false
	}
	if r.state == stateRowReady {
		r.err = errors.New("must call NextResultSet before Next")
		return false
	}
	r.state = stateResultSetReady
	return true
}

func (r *rows) Metadata() (*ResultMetadata, error) {
	for r.iter == nil || (r.state != stateEOQ && r.state != stateClosed) {
		return nil, &internal.Error{Status: http.StatusBadRequest, Err: errors.New("Metadata must not be called until result set iteration is complete")}
	}
	return r.feed.(*rowsIterator).ResultMetadata, nil
}

func (r *rows) ScanValue(dest interface{}) (err error) {
	runlock, err := r.makeReady(&err)
	if err != nil {
		return err
	}
	defer runlock()
	row := r.curVal.(*driver.Row)
	if row.Error != nil {
		return row.Error
	}
	if row.Value != nil {
		return json.NewDecoder(row.Value).Decode(dest)
	}
	return nil
}

func (r *rows) ScanDoc(dest interface{}) (err error) {
	runlock, err := r.makeReady(&err)
	if err != nil {
		return err
	}
	defer runlock()
	row := r.curVal.(*driver.Row)
	if err := row.Error; err != nil {
		return err
	}
	if row.Doc != nil {
		return json.NewDecoder(row.Doc).Decode(dest)
	}
	return &internal.Error{Status: http.StatusBadRequest, Message: "kivik: doc is nil; does the query include docs?"}
}

func (r *rows) ScanKey(dest interface{}) (err error) {
	runlock, err := r.makeReady(&err)
	if err != nil {
		return err
	}
	defer runlock()
	row := r.curVal.(*driver.Row)
	if err := json.Unmarshal(row.Key, dest); err != nil {
		return err
	}
	return row.Error
}

func (r *rows) ID() (string, error) {
	runlock, err := r.makeReady(nil)
	if err != nil {
		return "", err
	}
	defer runlock()
	row := r.curVal.(*driver.Row)
	return row.ID, row.Error
}

func (r *rows) Key() (string, error) {
	runlock, err := r.makeReady(nil)
	if err != nil {
		return "", err
	}
	defer runlock()
	row := r.curVal.(*driver.Row)
	return string(row.Key), row.Error
}

func (r *rows) Attachments() (*AttachmentsIterator, error) {
	runlock, err := r.makeReady(nil)
	if err != nil {
		return nil, err
	}
	defer runlock()
	row := r.curVal.(*driver.Row)
	if row.Error != nil {
		return nil, row.Error
	}
	if row.Attachments == nil {
		return nil, nil // TODO: should this return an error?
	}
	return &AttachmentsIterator{atti: row.Attachments}, nil
}

func (r *rows) Rev() (string, error) {
	row := r.curVal.(*driver.Row)
	return row.Rev, row.Error
}
