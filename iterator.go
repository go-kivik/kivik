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
	"errors"
	"io"
	"net/http"
	"sync"

	"github.com/go-kivik/kivik/v4/driver"
)

type iterator interface {
	Next(interface{}) error
	Close() error
}

// possible states of the result set iterator
const (
	// stateReady is the initial state before [ResultSet.Next] or
	// [ResultSet.NextResultSet] is called.
	stateReady = iota
	// stateResultSetReady is the state after calling [ResultSet.NextResultSet]
	stateResultSetReady
	// stateResultSetRowReady is the state after calling [ResultSet.Next] within
	// a result set.
	stateResultSetRowReady
	// stateEOQ is the state after having reached the final row in a result set.
	// [ResultSet.ResultSetNext] should be called next.
	stateEOQ
	// stateRowReady is the state when not iterating resultsets, after
	// [ResultSet.Next] has been called.
	stateRowReady
	// stateClosed means the last row has been retrieved. The iterator is no
	// longer usable.
	stateClosed
)

type iter struct {
	feed iterator

	mu      sync.RWMutex
	state   int   // Set to true once Next() has been called
	lasterr error // non-nil only if state == stateClosed

	cancel func() // cancel function to exit context goroutine when iterator is closed

	curVal interface{}
}

func (i *iter) rlock() (unlock func(), err error) {
	i.mu.RLock()
	if i.state == stateClosed {
		i.mu.RUnlock()
		return nil, &Error{Status: http.StatusBadRequest, Message: "kivik: Iterator is closed"}
	}
	if !i.ready() {
		i.mu.RUnlock()
		return nil, &Error{Status: http.StatusBadRequest, Message: "kivik: Iterator access before calling Next"}
	}
	return i.mu.RUnlock, nil
}

func (i *iter) ready() bool {
	return i.state == stateRowReady || i.state == stateResultSetReady || i.state == stateResultSetRowReady
}

// makeReady ensures that the iterator is ready to be read from. In the case
// that [iter.Next] has not been called, the returned unlock function will also
// close the iterator, and set e if [iter.Close] errors and e != nil.
func (i *iter) makeReady(e *error) (unlock func()) {
	i.mu.RLock()
	if !i.ready() {
		i.Next()
		return func() {
			i.mu.RUnlock()
			if err := i.Close(); err != nil && e != nil {
				*e = err
			}
		}
	}
	return i.mu.RUnlock
}

// newIterator instantiates a new iterator.
//
// ctx is a possibly-cancellable context.  zeroValue is an empty instance of
// the data type this iterator iterates over feed is the iterator interface,
// which typically wraps a driver.X iterator
func newIterator(ctx context.Context, feed iterator, zeroValue interface{}) *iter {
	i := &iter{
		feed:   feed,
		curVal: zeroValue,
	}
	ctx, i.cancel = context.WithCancel(ctx)
	go i.awaitDone(ctx)
	return i
}

// awaitDone blocks until the rows are closed or the context is cancelled, then
// closes the iterator if it's still open.
func (i *iter) awaitDone(ctx context.Context) {
	<-ctx.Done()
	_ = i.close(ctx.Err())
}

// NextResultSet prepares the iterator to read the next result set. It returns
// ture on success, or false if there are no more result sets to read, or if
// an error occurs while preparing it. [iter.Err] should be consulted to
// distinguish between the two.
func (i *iter) NextResultSet() bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if i.lasterr != nil {
		return false
	}
	if i.state == stateClosed {
		return false
	}
	if i.state == stateRowReady {
		i.lasterr = errors.New("must call NextResultSet before Next")
		return false
	}
	i.state = stateResultSetReady
	return true
}

// Next prepares the next iterator result value for reading. It returns true on
// success, or false if there is no next result or an error occurs while
// preparing it. [iter.Err] should be consulted to distinguish between the two.
func (i *iter) Next() bool {
	doClose, ok := i.next()
	if doClose {
		_ = i.Close()
	}
	return ok
}

func (i *iter) next() (doClose, ok bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if i.state == stateClosed {
		return false, false
	}
	err := i.feed.Next(i.curVal)
	if err == driver.EOQ {
		if i.state == stateResultSetReady || i.state == stateResultSetRowReady {
			i.state = stateEOQ
			i.lasterr = nil
			return false, false
		}
		return i.next()
	}
	switch i.state {
	case stateResultSetReady, stateResultSetRowReady:
		i.state = stateResultSetRowReady
	default:
		i.state = stateRowReady
	}
	i.lasterr = err
	if i.lasterr != nil {
		return true, false
	}
	return false, true
}

// Close closes the Iterator, preventing further enumeration, and freeing any
// resources (such as the http request body) of the underlying feed. If Next is
// called and there are no further results, Iterator is closed automatically and
// it will suffice to check the result of [iter.Err]. Close is idempotent and
// does not affect the result of [iter.Err].
func (i *iter) Close() error {
	return i.close(nil)
}

func (i *iter) close(err error) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.state == stateClosed {
		return nil
	}
	i.state = stateClosed

	if i.lasterr == nil {
		i.lasterr = err
	}

	err = i.feed.Close()

	if i.cancel != nil {
		i.cancel()
	}

	return err
}

// Err returns the error, if any, that was encountered during iteration. Err
// may be called after an explicit or implicit Close.
func (i *iter) Err() error {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if i.lasterr == io.EOF {
		return nil
	}
	return i.lasterr
}
