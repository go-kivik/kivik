package kivik

import (
	"context"
	"io"
	"sync"

	"github.com/flimzy/kivik/driver"
)

// BulkResults is an iterator over the results of a BulkDocs query.
type BulkResults struct {
	bulki driver.BulkResults

	// closemu prevents Rows from closing while thereis an active streaming
	// result. It is held for read during non-close operations and exclusively
	// during close.
	//
	// closemu guards lasterr and closed.
	closemu sync.RWMutex
	closed  bool
	lasterr error // non-nil only if closed is true

	curResult *driver.BulkResult
}

// Next returns the next BulkResult from the feed. If an error occurs, it will
// be returned and the feed closed. io.EOF will be returned when there are no
// more results.
func (r *BulkResults) Next() bool {
	doClose, ok := r.next()
	if doClose {
		_ = r.Close()
	}
	return ok
}

func (r *BulkResults) next() (doClose, ok bool) {
	r.closemu.RLock()
	defer r.closemu.RUnlock()
	if r.closed {
		return false, false
	}
	if r.curResult == nil {
		r.curResult = &driver.BulkResult{}
	}
	r.lasterr = r.bulki.Next(r.curResult)
	if r.lasterr != nil {
		return true, false
	}
	return false, true
}

// Close closes the feed. Any unread updates will still be accessible via
// Next().
func (r *BulkResults) Close() error {
	return r.close(nil)
}

func (r *BulkResults) close(err error) error {
	r.closemu.Lock()
	defer r.closemu.Unlock()
	if r.closed {
		return nil
	}
	r.closed = true

	if r.lasterr == nil {
		r.lasterr = err
	}

	return r.bulki.Close()
}

// ID returns the document ID name for the current result.
func (r *BulkResults) ID() string {
	return r.curResult.ID
}

// Rev returns the revision of the current curResult.
func (r *BulkResults) Rev() string {
	return r.curResult.Rev
}

// UpdateErr returns the error associated with the current result, or nil
// if none. Do not confuse this with Err, which returns an error for the
// iterator itself.
func (r *BulkResults) UpdateErr() error {
	return r.curResult.Error
}

// Err returns the error, if any, that was encountered during iteration. Err
// may be called after an explicit or implicit Close.
func (r *BulkResults) Err() error {
	r.closemu.RLock()
	defer r.closemu.RUnlock()
	if r.lasterr == io.EOF {
		return nil
	}
	return r.lasterr
}

// BulkDocs allows you to create and update multiple documents at the same time
// within a single request. This function returns an iterator over the results
// of the bulk operation.
// See http://docs.couchdb.org/en/2.0.0/api/database/bulk-api.html#db-bulk-docs
func (db *DB) BulkDocs(ctx context.Context, docs ...interface{}) (*BulkResults, error) {
	bulki, err := db.driverDB.BulkDocs(ctx, docs...)
	if err != nil {
		return nil, err
	}
	return &BulkResults{bulki: bulki}, nil
}
