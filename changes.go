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
)

// Changes is an iterator over the database changes feed.
type Changes struct {
	*iter
	changesi driver.Changes
}

// Next prepares the next result value for reading. It returns true on success
// or false if there are no more results, due to an error or the changes feed
// having been closed. [Changes.Err] should be consulted to determine any error.
func (c *Changes) Next() bool {
	return c.iter.Next()
}

// Err returns the error, if any, that was encountered during iteration. Err may
// be called after an explicit or implicit [Changes.Close].
func (c *Changes) Err() error {
	return c.iter.Err()
}

// Close closes the Changes feed, preventing further enumeration, and freeing
// any resources (such as the HTTP request body) of the underlying query. If
// [Changes.Next] is called and there are no further results, Changes is closed
// automatically and it will suffice to check the result of [Changes.Err]. Close
// is idempotent and does not affect the result of [Changes.Err].
func (c *Changes) Close() error {
	return c.iter.Close()
}

type changesIterator struct {
	driver.Changes
	*ChangesMetadata
}

var _ iterator = &changesIterator{}

func (c *changesIterator) Next(i interface{}) error {
	err := c.Changes.Next(i.(*driver.Change))
	if err == io.EOF || err == driver.EOQ {
		c.ChangesMetadata = &ChangesMetadata{
			LastSeq: c.Changes.LastSeq(),
			Pending: c.Changes.Pending(),
		}
	}
	return err
}

func newChanges(ctx context.Context, changesi driver.Changes) *Changes {
	return &Changes{
		iter:     newIterator(ctx, &changesIterator{Changes: changesi}, &driver.Change{}),
		changesi: changesi,
	}
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

// ScanDoc works the same as [Changes.ScanValue], but on the doc field of the
// result. It is only valid for results that include documents.
func (c *Changes) ScanDoc(dest interface{}) error {
	runlock, err := c.rlock()
	if err != nil {
		return err
	}
	defer runlock()
	return json.Unmarshal(c.curVal.(*driver.Change).Doc, dest)
}

// Changes returns an iterator over the real-time changes feed. The feed remains
// open until explicitly closed, or an error is encountered.
//
// See http://couchdb.readthedocs.io/en/latest/api/database/changes.html#get--db-_changes
func (db *DB) Changes(ctx context.Context, options ...Options) (*Changes, error) {
	changesi, err := db.driverDB.Changes(ctx, mergeOptions(options...))
	if err != nil {
		return nil, err
	}
	return newChanges(ctx, changesi), nil
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
// after [Next] returns false. Otherwise it will return an error.
func (c *Changes) Metadata() (*ChangesMetadata, error) {
	if c.iter == nil || (c.state != stateEOQ && c.state != stateClosed) {
		return nil, &Error{Status: http.StatusBadRequest, Err: errors.New("Metadata must not be called until result set iteration is complete")}
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
