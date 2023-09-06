package kivikmock

import (
	"context"
	"time"

	"github.com/go-kivik/kivik/v4/driver"
)

// Changes is a mocked collection of Changes results.
type Changes struct {
	iter
	lastSeq string
	pending int64
	etag    string
}

type driverChanges struct {
	context.Context
	*Changes
}

func coalesceChanges(changes *Changes) *Changes {
	if changes != nil {
		return changes
	}
	return &Changes{}
}

var _ driver.Changes = &driverChanges{}

func (r *driverChanges) Next(res *driver.Change) error {
	result, err := r.unshift(r.Context)
	if err != nil {
		return err
	}
	*res = *result.(*driver.Change)
	return nil
}

func (r *driverChanges) LastSeq() string { return r.lastSeq }
func (r *driverChanges) Pending() int64  { return r.pending }
func (r *driverChanges) ETag() string    { return r.etag }

// CloseError sets an error to be returned when the iterator is closed.
func (r *Changes) CloseError(err error) *Changes {
	r.closeErr = err
	return r
}

// LastSeq sets the last_seq value to be returned by the changes iterator.
func (r *Changes) LastSeq(seq string) *Changes {
	r.lastSeq = seq
	return r
}

// Pending sets the pending value to be returned by the changes iterator.
func (r *Changes) Pending(pending int64) *Changes {
	r.pending = pending
	return r
}

// ETag sets the etag value to be returned by the changes iterator.
func (r *Changes) ETag(etag string) *Changes {
	r.etag = etag
	return r
}

// AddChange adds a change result to be returned by the iterator. If
// AddResultError has been set, this method will panic.
func (r *Changes) AddChange(change *driver.Change) *Changes {
	if r.resultErr != nil {
		panic("It is invalid to set more changes after AddChangeError is defined.")
	}
	r.push(&item{item: change})
	return r
}

// AddChangeError adds an error to be returned during iteration.
func (r *Changes) AddChangeError(err error) *Changes {
	r.resultErr = err
	return r
}

// AddDelay adds a delay before the next iteration will complete.
func (r *Changes) AddDelay(delay time.Duration) *Changes {
	r.push(&item{delay: delay})
	return r
}

// Final converts the Changes object to a driver.Changes. This method is
// intended for use within WillExecute() to return results.
func (r *Changes) Final() driver.Changes {
	return &driverChanges{Changes: r}
}
