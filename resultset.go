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
	"io"
	"net/http"
	"reflect"
	"sync"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal"
)

// ResultMetadata contains metadata about certain queries.
type ResultMetadata struct {
	// Offset is the starting offset where the result set started.
	Offset int64

	// TotalRows is the total number of rows in the view which would have been
	// returned if no limiting were used.
	TotalRows int64

	// UpdateSeq is the sequence id of the underlying database the view
	// reflects, if requested in the query.
	UpdateSeq string

	// Warning is a warning generated by the query, if any.
	Warning string

	// Bookmark is the paging bookmark, if one was provided with the result
	// set. This is intended for use with the Mango /_find interface, with
	// CouchDB 2.1.1 and later. Consult the [CouchDB documentation] for
	// detailed usage instructions.
	//
	// [CouchDB documentation]: http://docs.couchdb.org/en/2.1.1/api/database/find.html#pagination
	Bookmark string
}

// ResultSet is an iterator over a multi-value query result set.
//
// Call [ResultSet.Next] to advance the iterator to the next item in the result
// set.
//
// The Scan* methods are expected to be called only once per iteration, as
// they may consume data from the network, rendering them unusable a second
// time.
//
// Calling [ResultSet.ScanDoc], [ResultSet.ScanKey], [ResultSet.ScanValue],
// [ResultSet.ID], [ResultSet.Key], [ResultSet.Rev], or [ResultSet.Attachments]
// before calling [ResultSet.Next] will operate on the first item in the
// resultset, then close the iterator immediately. This is for convenience in
// cases where only a single item is expected, so the extra effort of iterating
// is otherwise wasted. In this case, if the result set is empty, as when a view
// returns no results, an error of "no results" will be returned.
type ResultSet struct {
	*iter
	rowsi driver.Rows
}

func newResultSet(ctx context.Context, onClose func(), rowsi driver.Rows) *ResultSet {
	return &ResultSet{
		iter:  newIterator(ctx, onClose, &rowsIterator{Rows: rowsi}, &driver.Row{}),
		rowsi: rowsi,
	}
}

// Next prepares the next result value for reading. It returns true on success
// or false if there are no more results or an error occurs while preparing it.
// [ResultSet.Err] should be consulted to distinguish between the two.
//
// When Next returns false, and there are no more results/result sets to be
// read, the [ResultSet.Close] is called implicitly, negating the need to call
// it explicitly.
func (r *ResultSet) Next() bool {
	return r.iter.Next()
}

// NextResultSet prepares the next result set for reading. It returns false if
// there is no further result set or if there is an error advancing to it.
// [ResultSet.Err] should be consulted to distinguish between the two cases.
//
// After calling NextResultSet, [ResultSet.Next] must be called to advance to
// the first result in the resultset before scanning.
func (r *ResultSet) NextResultSet() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
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

// Err returns the error, if any, that was encountered during iteration.
// [ResultSet.Err] may be called after an explicit or implicit call to
// [ResultSet.Close].
func (r *ResultSet) Err() error {
	return r.iter.Err()
}

// Close closes the result set, preventing further iteration, and freeing
// any resources (such as the HTTP request body) of the underlying query.
// Close is idempotent and does not affect the result of
// [ResultSet.Err]. Close is safe to call concurrently with other ResultSet
// operations and will block until all other ResultSet operations finish.
func (r *ResultSet) Close() error {
	return r.iter.Close()
}

// Metadata returns the result metadata for the current query. It must be
// called after [ResultSet.Next] returns false. Otherwise it will return an
// error.
func (r *ResultSet) Metadata() (*ResultMetadata, error) {
	for r.iter == nil || (r.state != stateEOQ && r.state != stateClosed) {
		return nil, &internal.Error{Status: http.StatusBadRequest, Err: errors.New("Metadata must not be called until result set iteration is complete")}
	}
	return r.feed.(*rowsIterator).ResultMetadata, nil
}

// ScanValue copies the data from the result value into dest, which must be a
// pointer. This acts as a wrapper around [encoding/json.Unmarshal].
//
// If the row returned an error, it will be returned rather than unmarshaling
// the value, as error rows do not include values.
//
// Refer to the documentation for [encoding/json.Unmarshal] for unmarshaling
// details.
func (r *ResultSet) ScanValue(dest interface{}) (err error) {
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

// ScanDoc works the same as [ResultSet.ScanValue], but on the doc field of
// the result. It will return an error if the query does not include
// documents.
//
// If the row returned an error, it will be returned rather than
// unmarshaling the doc, as error rows do not include docs.
func (r *ResultSet) ScanDoc(dest interface{}) (err error) {
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

// ScanKey works the same as [ResultSet.ScanValue], but on the key field of the
// result. For simple keys, which are just strings, [ResultSet.Key] may be
// easier to use.
//
// Unlike [ResultSet.ScanValue] and [ResultSet.ScanDoc], this may successfully
// scan the key, and also return an error, if the row itself represents an error.
func (r *ResultSet) ScanKey(dest interface{}) (err error) {
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

// ID returns the ID of the most recent result.
func (r *ResultSet) ID() (string, error) {
	runlock, err := r.makeReady(nil)
	if err != nil {
		return "", err
	}
	defer runlock()
	row := r.curVal.(*driver.Row)
	return row.ID, row.Error
}

// Rev returns the document revision, when known. Not all result sets (such
// as those from views) include revision IDs, so this will return an empty
// string in such cases.
func (r *ResultSet) Rev() (string, error) {
	runlock, err := r.makeReady(nil)
	if err != nil {
		return "", err
	}
	defer runlock()
	row := r.curVal.(*driver.Row)
	return row.Rev, row.Error
}

// Key returns the Key of the most recent result as a raw JSON string. For
// compound keys, [ResultSet.ScanKey] may be more convenient.
func (r *ResultSet) Key() (string, error) {
	runlock, err := r.makeReady(nil)
	if err != nil {
		return "", err
	}
	defer runlock()
	row := r.curVal.(*driver.Row)
	return string(row.Key), row.Error
}

// Attachments returns an attachments iterator. At present, it is only set
// by [DB.Get] when doing a multi-part get from CouchDB (which is the
// default where supported). This may be extended to other cases in the
// future.
func (r *ResultSet) Attachments() (*AttachmentsIterator, error) {
	runlock, err := r.makeReady(nil)
	if err != nil {
		return nil, err
	}
	row := r.curVal.(*driver.Row)
	if row.Error != nil {
		runlock()
		return nil, row.Error
	}
	if row.Attachments == nil {
		runlock()
		return nil, nil // TODO: #804 return a proper error
	}
	return &AttachmentsIterator{
		onClose: runlock,
		atti:    row.Attachments,
	}, nil
}

// makeReady ensures that the iterator is ready to be read from. If i.err is
// set, it is returned. In the case that [iter.Next] has not been called, the
// returned unlock function will also close the iterator, and set e if
// [iter.Close] errors and e != nil.
func (r *ResultSet) makeReady(e *error) (unlock func(), err error) {
	r.mu.Lock()
	if r.err != nil {
		r.mu.Unlock()
		return nil, r.err
	}
	if !stateIsReady(r.state) {
		r.mu.Unlock()
		if !r.Next() {
			return nil, &internal.Error{Status: http.StatusNotFound, Message: "no results"}
		}
		r.wg.Add(1)
		return sync.OnceFunc(func() {
			r.wg.Done()
			if err := r.Close(); err != nil && e != nil {
				*e = err
			}
		}), nil
	}
	r.wg.Add(1)
	return sync.OnceFunc(func() {
		r.wg.Done()
		r.mu.Unlock()
	}), nil
}

type rowsIterator struct {
	driver.Rows
	*ResultMetadata
}

var _ iterator = &rowsIterator{}

func (r *rowsIterator) Next(i interface{}) error {
	row := i.(*driver.Row)
	row.ID = ""
	row.Rev = ""
	row.Key = row.Key[:0]
	row.Value = nil
	row.Doc = nil
	row.Attachments = nil
	row.Error = nil
	err := r.Rows.Next(row)
	if err == io.EOF || err == driver.EOQ {
		var warning, bookmark string
		if w, ok := r.Rows.(driver.RowsWarner); ok {
			warning = w.Warning()
		}
		if b, ok := r.Rows.(driver.Bookmarker); ok {
			bookmark = b.Bookmark()
		}
		r.ResultMetadata = &ResultMetadata{
			Offset:    r.Rows.Offset(),
			TotalRows: r.Rows.TotalRows(),
			UpdateSeq: r.Rows.UpdateSeq(),
			Warning:   warning,
			Bookmark:  bookmark,
		}
	}
	return err
}

// ScanAllDocs loops through the remaining documents in the resultset, and scans
// them into dest which must be a pointer to a slice or an array. Passing any
// other type will result in an error. If dest is an array, scanning will stop
// once the array is filled.  The iterator is closed by this method. It is
// possible that an error will be returned, and that one or more documents were
// successfully scanned.
func ScanAllDocs(r *ResultSet, dest interface{}) error {
	return scanAll(r, dest, r.ScanDoc)
}

// ScanAllValues works like [ScanAllDocs], but scans the values rather than docs.
func ScanAllValues(r *ResultSet, dest interface{}) error {
	return scanAll(r, dest, r.ScanValue)
}

func scanAll(r *ResultSet, dest interface{}, scan func(interface{}) error) (err error) {
	defer func() {
		closeErr := r.Close()
		if err == nil {
			err = closeErr
		}
	}()
	if err := r.Err(); err != nil {
		return err
	}

	value := reflect.ValueOf(dest)
	if value.Kind() != reflect.Ptr {
		return errors.New("must pass a pointer to ScanAllDocs")
	}
	if value.IsNil() {
		return errors.New("nil pointer passed to ScanAllDocs")
	}

	direct := reflect.Indirect(value)
	var limit int

	switch direct.Kind() {
	case reflect.Array:
		limit = direct.Len()
		if limit == 0 {
			return errors.New("0-length array passed to ScanAllDocs")
		}
	case reflect.Slice:
	default:
		return errors.New("dest must be a pointer to a slice or array")
	}

	base := value.Type()
	if base.Kind() == reflect.Ptr {
		base = base.Elem()
	}
	base = base.Elem()

	for i := 0; r.Next(); i++ {
		if limit > 0 && i >= limit {
			return nil
		}
		vp := reflect.New(base)
		err = scan(vp.Interface())
		if limit > 0 { // means this is an array
			direct.Index(i).Set(reflect.Indirect(vp))
		} else {
			direct.Set(reflect.Append(direct, reflect.Indirect(vp)))
		}
	}
	return nil
}
