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

//go:build js
// +build js

package pouchdb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/gopherjs/gopherjs/js"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/pouchdb/bindings"
)

type changesFeed struct {
	changes *js.Object
	ctx     context.Context
	feed    chan *driver.Change
	errMu   sync.Mutex
	err     error
	lastSeq string
}

var _ driver.Changes = &changesFeed{}

func newChangesFeed(ctx context.Context, changes *js.Object) *changesFeed {
	const chanLen = 32
	feed := make(chan *driver.Change, chanLen)
	c := &changesFeed{
		ctx:     ctx,
		changes: changes,
		feed:    feed,
	}

	changes.Call("on", "change", c.change)
	changes.Call("on", "complete", c.complete)
	changes.Call("on", "error", c.error)
	return c
}

type changeRow struct {
	*js.Object
	ID      string     `js:"id"`
	Seq     string     `js:"seq"`
	Changes *js.Object `js:"changes"`
	Doc     *js.Object `js:"doc"`
	Deleted bool       `js:"deleted"`
}

func (c *changesFeed) setErr(err error) {
	c.errMu.Lock()
	c.err = err
	c.errMu.Unlock()
}

func (c *changesFeed) Next(row *driver.Change) error {
	c.errMu.Lock()
	if c.err != nil {
		c.errMu.Unlock()
		return c.err
	}
	c.errMu.Unlock()
	select {
	case <-c.ctx.Done():
		err := c.ctx.Err()
		c.setErr(err)
		return err
	case newRow, ok := <-c.feed:
		if !ok {
			c.setErr(io.EOF)
			return io.EOF
		}
		*row = *newRow
	}
	return nil
}

func (c *changesFeed) Close() error {
	c.changes.Call("cancel")
	return nil
}

// LastSeq returns the last_seq id, as returned by PouchDB.
func (c *changesFeed) LastSeq() string {
	return c.lastSeq
}

// Pending returns 0 for PouchDB.
func (c *changesFeed) Pending() int64 {
	return 0
}

// ETag returns an empty string for PouchDB.
func (c *changesFeed) ETag() string {
	return ""
}

func (c *changesFeed) change(change *changeRow) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				_ = c.Close()
				if e, ok := r.(error); ok {
					c.err = e
				} else {
					c.err = fmt.Errorf("%v", r)
				}
			}
		}()
		changedRevs := make([]string, 0, change.Changes.Length())
		for i := 0; i < change.Changes.Length(); i++ {
			changedRevs = append(changedRevs, change.Changes.Index(i).Get("rev").String())
		}
		var doc json.RawMessage
		if change.Doc != js.Undefined {
			doc = json.RawMessage(js.Global.Get("JSON").Call("stringify", change.Doc).String())
		}
		row := &driver.Change{
			ID:      change.ID,
			Seq:     change.Seq,
			Deleted: change.Deleted,
			Doc:     doc,
			Changes: changedRevs,
		}
		c.feed <- row
	}()
}

func (c *changesFeed) complete(info *js.Object) {
	if results := info.Get("results"); results != js.Undefined {
		for _, result := range results.Interface().([]interface{}) {
			c.change(&changeRow{
				Object: result.(*js.Object),
			})
		}
	}

	c.lastSeq = info.Get("last_seq").String()

	close(c.feed)
}

func (c *changesFeed) error(e *js.Object) {
	c.setErr(bindings.NewPouchError(e))
}

func (d *db) Changes(ctx context.Context, options driver.Options) (driver.Changes, error) {
	opts := map[string]interface{}{}
	options.Apply(opts)
	changes, err := d.db.Changes(ctx, opts)
	if err != nil {
		return nil, err
	}

	return newChangesFeed(ctx, changes), nil
}
