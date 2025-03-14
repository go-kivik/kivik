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
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
)

// DBUpdates is an iterator for database updates.
type DBUpdates struct {
	*iter
}

type updatesIterator struct{ driver.DBUpdates }

var _ iterator = &updatesIterator{}

func (r *updatesIterator) Next(i interface{}) error {
	update := i.(*driver.DBUpdate)
	update.DBName = ""
	update.Seq = ""
	update.Type = ""
	return r.DBUpdates.Next(update)
}

func newDBUpdates(ctx context.Context, onClose func(), updatesi driver.DBUpdates) *DBUpdates {
	return &DBUpdates{
		iter: newIterator(ctx, onClose, &updatesIterator{updatesi}, &driver.DBUpdate{}),
	}
}

// Close closes the iterator, preventing further enumeration, and freeing any
// resources (such as the http request body) of the underlying feed. If
// [DBUpdates.Next] is called and there are no further results, the iterator is
// closed automatically and it will suffice to check the result of
// [DBUpdates.Err]. Close is idempotent and does not affect the result of
// [DBUpdates.Err].
func (f *DBUpdates) Close() error {
	return f.iter.Close()
}

// Err returns the error, if any, that was encountered during iteration. Err may
// be called after an explicit or implicit [DBUpdates.Close].
func (f *DBUpdates) Err() error {
	return f.iter.Err()
}

// Next prepares the next iterator result value for reading. It returns true on
// success, or false if there is no next result or an error occurs while
// preparing it. [DBUpdates.Err] should be consulted to distinguish between the
// two.
func (f *DBUpdates) Next() bool {
	return f.iter.Next()
}

// DBName returns the database name for the current update.
func (f *DBUpdates) DBName() string {
	err := f.isReady()
	if err != nil {
		return ""
	}
	return f.curVal.(*driver.DBUpdate).DBName
}

// Type returns the type of the current update.
func (f *DBUpdates) Type() string {
	err := f.isReady()
	if err != nil {
		return ""
	}
	return f.curVal.(*driver.DBUpdate).Type
}

// Seq returns the update sequence of the current update.
func (f *DBUpdates) Seq() string {
	err := f.isReady()
	if err != nil {
		return ""
	}
	return f.curVal.(*driver.DBUpdate).Seq
}

// LastSeq returns the last sequence ID reported, or in the case no results
// were returned due to `since`	being set to `now`, or some other value that
// excludes all results, the current sequence ID. It must be called after
// [DBUpdates.Next] returns false or [DBUpdates.Iterator] has been completely
// and successfully iterated. Otherwise it will return an error.
func (f *DBUpdates) LastSeq() (string, error) {
	for f.iter == nil || f.state != stateEOQ && f.state != stateClosed {
		return "", &internal.Error{Status: http.StatusBadRequest, Err: errors.New("LastSeq must not be called until results iteration is complete")}
	}
	driverUpdates := f.feed.(*updatesIterator).DBUpdates
	if lastSeqer, ok := driverUpdates.(driver.LastSeqer); ok {
		return lastSeqer.LastSeq()
	}
	return "", nil
}

// DBUpdates begins polling for database updates. Canceling the context will
// close the iterator. The iterator will also close automatically if there are
// no more updates, when an error occurs, or when the [DBUpdates.Close] method
// is called. The [DBUpdates.Err] method should be consulted to determine if
// there was an error during iteration.
//
// For historical reasons, the CouchDB driver's implementation of this function
// defaults to feed=continuous and since=now. To use the default CouchDB
// behavior, set feed to either the empty string or "normal", and since to the
// empty string. In kivik/v5, the default behavior will be to use feed=normal
// as CouchDB does by default.
func (c *Client) DBUpdates(ctx context.Context, options ...Option) *DBUpdates {
	updater, ok := c.driverClient.(driver.DBUpdater)
	if !ok {
		return &DBUpdates{errIterator(&internal.Error{Status: http.StatusNotImplemented, Message: "kivik: driver does not implement DBUpdater"})}
	}

	endQuery, err := c.startQuery()
	if err != nil {
		return &DBUpdates{errIterator(err)}
	}

	updatesi, err := updater.DBUpdates(ctx, multiOptions(options))
	if err != nil {
		endQuery()
		return &DBUpdates{errIterator(err)}
	}
	return newDBUpdates(context.Background(), endQuery, updatesi)
}

// DBUpdate represents a database update as returned by [DBUpdates.Iterator].
//
// !!NOTICE!! This struct is considered experimental, and may change without
// notice.
type DBUpdate struct {
	DBName string `json:"db_name"`
	Type   string `json:"type"`
	Seq    string `json:"seq"`
}

// Iterator returns a function that can be used to iterate over the DB updates
// feed. This function works with Go 1.23's range functions, and is an
// alternative to using [DBUpdates.Next] directly.
//
// !!NOTICE!! This function is considered experimental, and may change without
// notice.
func (f *DBUpdates) Iterator() func(yield func(*DBUpdate, error) bool) {
	return func(yield func(*DBUpdate, error) bool) {
		for f.Next() {
			update := f.curVal.(*driver.DBUpdate)
			if !yield((*DBUpdate)(update), nil) {
				_ = f.Close()
				return
			}
		}
		if err := f.Err(); err != nil {
			yield(nil, err)
		}
	}
}
