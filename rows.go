package kivik

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"

	"golang.org/x/net/context"

	"github.com/flimzy/kivik/driver"
)

// Rows is the result of a multi-value query, such as a view or /_all_docs. Its
// cursor starts before the first value of a result set. Use Next to advance
// through the rows.
type Rows struct {
	rowsi  driver.Rows
	cancel func() // called when Rows is closed, may be nil.

	// closemu prevents Rows from closing while thereis an active streaming
	// result. It is held for read during non-close operations and exclusively
	// during close.
	//
	// closemu guards lasterr and closed.
	closemu sync.RWMutex
	closed  bool
	lasterr error // non-nil only if closed is true

	curRow *driver.Row
}

// initContextClose closes the Rows when the context is cancelled.
func (r *Rows) initContextClose(ctx context.Context) {
	ctx, r.cancel = context.WithCancel(ctx)
	go r.awaitDone(ctx)
}

// awaitDone blocks until the rows are closed or the context is cancelled.
func (r *Rows) awaitDone(ctx context.Context) {
	<-ctx.Done()
	_ = r.close(ctx.Err())
}

// Next prepares the next result value for reading with the Scan method. It
// returns true on success, or false if there is no next result or an error
// occurs while preparing it. Err should be consulted to distinguish between
// the two.
func (r *Rows) Next() bool {
	doClose, ok := r.next()
	if doClose {
		_ = r.Close()
	}
	return ok
}

func (r *Rows) next() (doClose, ok bool) {
	r.closemu.RLock()
	defer r.closemu.RUnlock()
	if r.closed {
		return false, false
	}
	if r.curRow == nil {
		r.curRow = &driver.Row{}
	}
	r.lasterr = r.rowsi.Next(r.curRow)
	if r.lasterr != nil {
		return true, false
	}
	return false, true
}

// Close closes the Rows, preventing further enumeration, and freeing any
// resources (such as the http request body) of the underlying query. If Next is
// called and there are no further results, Rows is closed automatically and it
// will suffice to check the result of Err. Close is idempotent and does not
// affect the result of Err.
func (r *Rows) Close() error {
	return r.close(nil)
}

func (r *Rows) close(err error) error {
	r.closemu.Lock()
	defer r.closemu.Unlock()
	if r.closed {
		return nil
	}
	r.closed = true

	if r.lasterr == nil {
		r.lasterr = err
	}

	err = r.rowsi.Close()

	if r.cancel != nil {
		r.cancel()
	}
	return err
}

var errNilPtr = errors.New("kivik: destination pointer is nil")

// ScanValue copies the data from the result value into the value pointed at by
// dest. Think of this as a json.Unmarshal into dest.
//
// If the dest argument has type *[]byte, Scan stores a copy of the input data.
// The copy is owned by the caller and can be modified and held indefinitely.
//
// The copy can be avoided by using an argument of type *json.RawMessage
// instead. After a Scaninto a json.RawMessage, the slice is only valid until
// the next call to Next, Scan, or Close.
//
// For all other types, refer to the documentation for json.Unmarshal for type
// conversion rules.
func (r *Rows) ScanValue(dest interface{}) error {
	return r.scan(dest, r.curRow.Value)
}

// ScanDoc works the same as ScanValue, but on the doc field of the result. It
// is only valid for results that include documents.
func (r *Rows) ScanDoc(dest interface{}) error {
	return r.scan(dest, r.curRow.Doc)
}

// ScanKey works the same as ScanValue, but on the key field of the result. For
// simple keys, which are just strings, the Key() method may be easier to use.
func (r *Rows) ScanKey(dest interface{}) error {
	return r.scan(dest, r.curRow.Key)
}

func (r *Rows) scan(dest interface{}, val json.RawMessage) error {
	r.closemu.RLock()
	if r.closed {
		r.closemu.RUnlock()
		return errors.New("kivik: Rows are closed")
	}
	r.closemu.RUnlock()
	if r.curRow == nil {
		return errors.New("kivik: Scan called without calling Next")
	}
	switch d := dest.(type) {
	case *[]byte:
		if d == nil {
			return errNilPtr
		}
		tgt := make([]byte, len(val))
		copy(tgt, val)
		*d = tgt
		return nil
	case *json.RawMessage:
		if d == nil {
			return errNilPtr
		}
		*d = val
		return nil
	}
	return json.Unmarshal(val, dest)
}

// ID returns the ID of the last-read result.
func (r *Rows) ID() string {
	if r.curRow != nil {
		return r.curRow.ID
	}
	return ""
}

// Key returns the Key of the last-read result as a de-quoted JSON object. For
// compound keys, the ScanKey() method may be more convenient.
func (r *Rows) Key() string {
	if r.curRow != nil {
		return strings.Trim(string(r.curRow.Key), `"`)
	}
	return ""
}

// Changes returns a list of changed revs. Only valid for the changes feed..
func (r *Rows) Changes() []string {
	if r.curRow != nil {
		return r.curRow.Changes
	}
	return nil
}

// Deleted returns true for the changes feed if the change relates to a deleted
// document.
func (r *Rows) Deleted() bool {
	if r.curRow != nil {
		return r.curRow.Deleted
	}
	return false
}

// Offset returns the starting offset where the result set started. It is
// only guaranteed to be set after all result rows have been enumerated through
// by Next, and thus should only be read after processing all rows in a result
// set. Calling Close before enumerating will render this value unreliable.
func (r *Rows) Offset() int64 {
	return r.rowsi.Offset()
}

// TotalRows returns the total number of rows in the view which would have been
// returned if no limiting were used. This value is only guaranteed to be set
// after all result rows have been enumerated through by Next, and thus should
// only be read after processing all rows in a result set. Calling Close before
// enumerating will render this value unreliable.
func (r *Rows) TotalRows() int64 {
	return r.rowsi.TotalRows()
}

// UpdateSeq returns the sequence id of the underlying database the view
// reflects, if requested in the query.
func (r *Rows) UpdateSeq() string {
	return r.rowsi.UpdateSeq()
}

// Err returns the error, if any, that was encountered during iteration. Err
// may be called after an explicit or implicit Close.
func (r *Rows) Err() error {
	r.closemu.RLock()
	defer r.closemu.RUnlock()
	if r.lasterr == io.EOF {
		return nil
	}
	return r.lasterr
}
