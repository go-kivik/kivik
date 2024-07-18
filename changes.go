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

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
)

// Changes is an iterator over the database changes feed.
type Changes struct {
	*iter
	changesi driver.Changes
}

type changesIterator struct {
	driver.Changes
	*ChangesMetadata
}

var _ iterator = &changesIterator{}

func (c *changesIterator) Next(i interface{}) error {
	change := i.(*driver.Change)
	change.ID = ""
	change.Seq = ""
	change.Deleted = false
	change.Changes = change.Changes[:0]
	change.Doc = change.Doc[:0]
	err := c.Changes.Next(change)
	if err == io.EOF || err == driver.EOQ {
		c.ChangesMetadata = &ChangesMetadata{
			LastSeq: c.Changes.LastSeq(),
			Pending: c.Changes.Pending(),
		}
	}
	return err
}

func newChanges(ctx context.Context, onClose func(), changesi driver.Changes) *Changes {
	return &Changes{
		iter:     newIterator(ctx, onClose, &changesIterator{Changes: changesi}, &driver.Change{}),
		changesi: changesi,
	}
}

// Close closes the iterator, preventing further enumeration, and freeing any
// resources (such as the http request body) of the underlying feed. If
// [Changes.Next] is called and there are no further results, the iterator is
// closed automatically and it will suffice to check the result of
// [Changes.Err]. Close is idempotent and does not affect the result of
// [Changes.Err].
func (c *Changes) Close() error {
	return c.iter.Close()
}

// Err returns the error, if any, that was encountered during iteration. Err may
// be called after an explicit or implicit [Changes.Close].
func (c *Changes) Err() error {
	return c.iter.Err()
}

// Next prepares the next iterator result value for reading. It returns true on
// success, or false if there is no next result or an error occurs while
// preparing it. [Changes.Err] should be consulted to distinguish between the
// two.
func (c *Changes) Next() bool {
	return c.iter.Next()
}

// Changes returns a list of changed revs.
func (c *Changes) Changes() []string {
	return c.curVal.(*driver.Change).Changes
}

// Deleted returns true if the change relates to a deleted document.
func (c *Changes) Deleted() bool {
	return c.curVal.(*driver.Change).Deleted
}

// ID returns the ID of the current result.
func (c *Changes) ID() string {
	return c.curVal.(*driver.Change).ID
}

// ScanDoc copies the data from the result into dest.  See [ResultSet.ScanValue]
// for additional details.
func (c *Changes) ScanDoc(dest interface{}) error {
	err := c.isReady()
	if err != nil {
		return err
	}
	return json.Unmarshal(c.curVal.(*driver.Change).Doc, dest)
}

// Changes returns an iterator over the real-time [changes feed]. The feed remains
// open until explicitly closed, or an error is encountered.
//
// [changes feed]: http://couchdb.readthedocs.io/en/latest/api/database/changes.html#get--db-_changes
func (db *DB) Changes(ctx context.Context, options ...Option) *Changes {
	if db.err != nil {
		return &Changes{iter: errIterator(db.err)}
	}
	endQuery, err := db.startQuery()
	if err != nil {
		return &Changes{iter: errIterator(err)}
	}
	changesi, err := db.driverDB.Changes(ctx, multiOptions(options))
	if err != nil {
		endQuery()
		return &Changes{iter: errIterator(err)}
	}
	return newChanges(ctx, endQuery, changesi)
}

// Seq returns the Seq of the current result.
func (c *Changes) Seq() string {
	return c.curVal.(*driver.Change).Seq
}

// ChangesMetadata contains metadata about a changes feed.
type ChangesMetadata struct {
	// LastSeq is the last update sequence id present in the change set, if
	// returned by the server.
	LastSeq string
	// Pending is the count of remaining items in the change feed.
	Pending int64
}

// Metadata returns the result metadata for the changes feed. It must be called
// after [Changes.Next] returns false or [Changes.Iterator] has been completely
// and successfully iterated. Otherwise it will return an error.
func (c *Changes) Metadata() (*ChangesMetadata, error) {
	if c.iter == nil || (c.state != stateEOQ && c.state != stateClosed) {
		return nil, &internal.Error{Status: http.StatusBadRequest, Err: errors.New("Metadata must not be called until result set iteration is complete")}
	}
	return c.feed.(*changesIterator).ChangesMetadata, nil
}

// ETag returns the unquoted ETag header, if any.
func (c *Changes) ETag() string {
	if c.changesi == nil {
		return ""
	}
	return c.changesi.ETag()
}

// Change represents a single change in the changes feed, as returned by
// [Changes.Iterator].
//
// !!NOTICE!! This struct is considered experimental, and may change without
// notice.
type Change struct {
	// ID is the document ID to which the change relates.
	ID string `json:"id"`
	// Seq is the update sequence for the changes feed.
	Seq string `json:"seq"`
	// Deleted is set to true for the changes feed, if the document has been
	// deleted.
	Deleted bool `json:"deleted"`
	// Changes represents a list of document leaf revisions for the /_changes
	// endpoint.
	Changes []string `json:"-"`
	// Doc is the raw, un-decoded JSON document. This is only populated when
	// include_docs=true is set.
	doc json.RawMessage
}

// ScanDoc copies the data from the result into dest.  See [Row.ScanValue]
// for additional details.
func (c *Change) ScanDoc(dest interface{}) error {
	return json.Unmarshal(c.doc, dest)
}

// Iterator returns a function that can be used to iterate over the changes
// feed. This function works with Go 1.23's range functions, and is an
// alternative to using [Changes.Next] directly.
//
// !!NOTICE!! This function is considered experimental, and may change without
// notice.
func (c *Changes) Iterator() func(yield func(*Change, error) bool) {
	return func(yield func(*Change, error) bool) {
		for c.Next() {
			dChange := c.curVal.(*driver.Change)
			change := &Change{
				ID:      dChange.ID,
				Seq:     dChange.Seq,
				Deleted: dChange.Deleted,
				Changes: dChange.Changes,
				doc:     dChange.Doc,
			}
			if !yield(change, nil) {
				_ = c.Close()
				return
			}
		}
		if err := c.Err(); err != nil {
			yield(nil, err)
		}
	}
}
