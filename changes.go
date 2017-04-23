package kivik

import (
	"github.com/flimzy/kivik/driver"
	"golang.org/x/net/context"
)

// Changes is an iterator over the database changes feed.
type Changes struct {
	*Iterator
	changesi driver.Changes
}

type changesIterator struct{ driver.Changes }

var _ iterator = &changesIterator{}

func (c *changesIterator) SetValue() interface{}    { return &driver.Change{} }
func (c *changesIterator) Next(i interface{}) error { return c.Changes.Next(i.(*driver.Change)) }

func newChanges(ctx context.Context, changesi driver.Changes) *Changes {
	return &Changes{
		Iterator: newIterator(ctx, &changesIterator{changesi}),
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
	return c.curVal.(*driver.Row).ID
}

// ScanDoc works the same as ScanValue, but on the doc field of the result. It
// is only valid for results that include documents.
func (c *Changes) ScanDoc(dest interface{}) error {
	runlock, err := c.rlock()
	if err != nil {
		return err
	}
	defer runlock()
	return scan(dest, c.curVal.(*driver.Change).Doc)
}

// Changes returns an iterator over the real-time changes feed. The feed remains
// open until explicitly closed, or an error is encountered.
// See http://couchdb.readthedocs.io/en/latest/api/database/changes.html#get--db-_changes
func (db *DB) Changes(ctx context.Context, options ...Options) (*Changes, error) {
	opts, err := mergeOptions(options...)
	if err != nil {
		return nil, err
	}
	changesi, err := db.driverDB.Changes(ctx, opts)
	if err != nil {
		return nil, err
	}
	return newChanges(ctx, changesi), nil
}
