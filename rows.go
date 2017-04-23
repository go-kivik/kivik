package kivik

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/flimzy/kivik/driver"
)

// Rows is the result of a multi-value query, such as a view or /_all_docs. Its
// cursor starts before the first value of a result set. Use Next to advance
// through the rows.
type Rows struct {
	*Iterator
	rowsi driver.Rows
}

type rowsIterator struct{ driver.Rows }

func (r *rowsIterator) SetValue() interface{}    { return &driver.Row{} }
func (r *rowsIterator) Next(i interface{}) error { return r.Rows.Next(i.(*driver.Row)) }

func newRows(ctx context.Context, rowsi driver.Rows) *Rows {
	return &Rows{
		Iterator: newIterator(ctx, &rowsIterator{rowsi}),
		rowsi:    rowsi,
	}
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
	return r.scan(dest, r.curVal.(*driver.Row).Value)
}

// ScanDoc works the same as ScanValue, but on the doc field of the result. It
// is only valid for results that include documents.
func (r *Rows) ScanDoc(dest interface{}) error {
	return r.scan(dest, r.curVal.(*driver.Row).Doc)
}

// ScanKey works the same as ScanValue, but on the key field of the result. For
// simple keys, which are just strings, the Key() method may be easier to use.
func (r *Rows) ScanKey(dest interface{}) error {
	return r.scan(dest, r.curVal.(*driver.Row).Key)
}

func (r *Rows) scan(dest interface{}, val json.RawMessage) error {
	r.closemu.RLock()
	if r.closed {
		r.closemu.RUnlock()
		return errors.New("kivik: Rows are closed")
	}
	r.closemu.RUnlock()
	if !r.ready {
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
	return r.curVal.(*driver.Row).ID
}

// Key returns the Key of the last-read result as a de-quoted JSON object. For
// compound keys, the ScanKey() method may be more convenient.
func (r *Rows) Key() string {
	return strings.Trim(string(r.curVal.(*driver.Row).Key), `"`)
}

// Changes returns a list of changed revs. Only valid for the changes feed..
func (r *Rows) Changes() []string {
	return r.curVal.(*driver.Row).Changes
}

// Deleted returns true for the changes feed if the change relates to a deleted
// document.
func (r *Rows) Deleted() bool {
	return r.curVal.(*driver.Row).Deleted
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
