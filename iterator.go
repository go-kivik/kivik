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
	"io"
	"net/http"
	"sync"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
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
	feed    iterator
	onClose func()

	mu    sync.Mutex
	state int   // Set to true once Next() has been called
	err   error // non-nil only if state == stateClosed
	wg    sync.WaitGroup

	cancel func() // cancel function to exit context goroutine when iterator is closed

	curVal interface{}
}

// isReady returns an error if the iterator is not ready, because it has been
// closed, or has not been made ready yet.
func (i *iter) isReady() error {
	if i.state == stateClosed {
		return &internal.Error{Status: http.StatusBadRequest, Message: "kivik: Iterator is closed"}
	}
	if !stateIsReady(i.state) {
		return &internal.Error{Status: http.StatusBadRequest, Message: "kivik: Iterator access before calling Next"}
	}
	return nil
}

func stateIsReady(state int) bool {
	switch state {
	case stateRowReady, stateResultSetReady, stateResultSetRowReady, stateClosed:
		return true
	}
	return false
}

// newIterator instantiates a new iterator.
//
// ctx is a possibly-cancellable context.  zeroValue is an empty instance of
// the data type this iterator iterates over feed is the iterator interface,
// which typically wraps a driver.X iterator
func newIterator(ctx context.Context, onClose func(), feed iterator, zeroValue interface{}) *iter {
	i := &iter{
		onClose: onClose,
		feed:    feed,
		curVal:  zeroValue,
	}
	ctx, i.cancel = context.WithCancel(ctx)
	go i.awaitDone(ctx)
	return i
}

// errIterator instantiates a new iterator that is already closed, and only
// returns an error.
func errIterator(err error) *iter {
	return &iter{
		state: stateClosed,
		err:   err,
	}
}

// awaitDone blocks until the rows are closed or the context is cancelled, then
// closes the iterator if it's still open.
func (i *iter) awaitDone(ctx context.Context) {
	<-ctx.Done()
	i.mu.Lock()
	_ = i.closeErr(ctx.Err())
	i.mu.Unlock()
}

// Next prepares the next iterator result value for reading. It returns true on
// success, or false if there is no next result or an error occurs while
// preparing it. [Err] should be consulted to distinguish between the two.
func (i *iter) Next() bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.next()
}

// next is the same as Next but doesn't do its own locking.
func (i *iter) next() bool {
	if i.state == stateClosed {
		return false
	}
	for {
		err := i.feed.Next(i.curVal)
		if err == driver.EOQ {
			if i.state == stateResultSetReady || i.state == stateResultSetRowReady {
				i.state = stateEOQ
				i.err = nil
				return false
			}
			continue
		}
		switch i.state {
		case stateResultSetReady, stateResultSetRowReady:
			i.state = stateResultSetRowReady
		default:
			i.state = stateRowReady
		}
		i.err = err
		if i.err != nil {
			_ = i.closeErr(nil)
			return false
		}
		return true
	}
}

// Close closes the iterator, preventing further enumeration, and freeing any
// resources (such as the http request body) of the underlying feed. If [Next]
// is called and there are no further results, the iterator is closed
// automatically and it will suffice to check the result of [Err]. Close is
// idempotent and does not affect the result of [Err].
func (i *iter) Close() error {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.wg.Wait()
	return i.closeErr(nil)
}

func (i *iter) closeErr(err error) error {
	if i.state == stateClosed {
		return nil
	}
	i.state = stateClosed

	if i.err == nil {
		i.err = err
	}

	err = i.feed.Close()

	if i.cancel != nil {
		i.cancel()
	}

	if i.onClose != nil {
		i.onClose()
	}

	return err
}

// Err returns the error, if any, that was encountered during iteration. Err may
// be called after an explicit or implicit [Close].
func (i *iter) Err() error {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.err == io.EOF {
		return nil
	}
	return i.err
}
