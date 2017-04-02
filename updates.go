package kivik

import (
	"io"
	"sync"

	"github.com/flimzy/kivik/driver"
)

// DBUpdate represents a single DB Update event.
type DBUpdate struct {
	DBName string
	Seq    string
	Type   string
}

// DBUpdateFeed provides access to database updates.
type DBUpdateFeed struct {
	updatesi driver.DBUpdates

	// closemu prevents Rows from closing while thereis an active streaming
	// result. It is held for read during non-close operations and exclusively
	// during close.
	//
	// closemu guards lasterr and closed.
	closemu sync.RWMutex
	closed  bool
	lasterr error // non-nil only if closed is true

	curUpdate *driver.DBUpdate
}

// Next returns the next DBUpdate from the feed. This function will block
// until an event is received. If an error occurs, it will be returned and
// the feed closed. If the feed was closed normally, io.EOF will be returned
// when there are no more events in the buffer.
func (f *DBUpdateFeed) Next() bool {
	doClose, ok := f.next()
	if doClose {
		_ = f.Close()
	}
	return ok
}

func (f *DBUpdateFeed) next() (doClose, ok bool) {
	f.closemu.RLock()
	defer f.closemu.RUnlock()
	if f.closed {
		return false, false
	}
	if f.curUpdate == nil {
		f.curUpdate = &driver.DBUpdate{}
	}
	f.lasterr = f.updatesi.Next(f.curUpdate)
	if f.lasterr != nil {
		return true, false
	}
	return false, true
}

// Close closes the feed. Any unread updates will still be accessible via
// Next().
func (f *DBUpdateFeed) Close() error {
	return f.close(nil)
}

func (f *DBUpdateFeed) close(err error) error {
	f.closemu.Lock()
	defer f.closemu.Unlock()
	if f.closed {
		return nil
	}
	f.closed = true

	if f.lasterr == nil {
		f.lasterr = err
	}

	return f.updatesi.Close()
}

// DBName returns the database name for the current update.
func (f *DBUpdateFeed) DBName() string {
	return f.curUpdate.DBName
}

// Type returns the type of the current update.
func (f *DBUpdateFeed) Type() string {
	return f.curUpdate.Type
}

// Seq returns the update sequence of the current update.
func (f *DBUpdateFeed) Seq() string {
	return f.curUpdate.Seq
}

// Err returns the error, if any, that was encountered during iteration. Err
// may be called after an explicit or implicit Close.
func (f *DBUpdateFeed) Err() error {
	f.closemu.RLock()
	defer f.closemu.RUnlock()
	if f.lasterr == io.EOF {
		return nil
	}
	return f.lasterr
}

// DBUpdates begins polling for database updates.
func (c *Client) DBUpdates() (*DBUpdateFeed, error) {
	updater, ok := c.driverClient.(driver.DBUpdater)
	if !ok {
		return nil, ErrNotImplemented
	}
	updatesi, err := updater.DBUpdates()
	if err != nil {
		return nil, err
	}
	return &DBUpdateFeed{updatesi: updatesi}, nil
}
