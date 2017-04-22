package kivik

import (
	"context"
	"sync"
	"time"

	"github.com/flimzy/kivik/driver"
)

// ReplicationState represents a replication's state
type ReplicationState string

// The possible values for the _replication_state field in _replicator documents
// plus a blank value for unstarted replications.
const (
	ReplicationNotStarted ReplicationState = ""
	ReplicationStarted    ReplicationState = "triggered"
	ReplicationError      ReplicationState = "error"
	ReplicationComplete   ReplicationState = "completed"
)

// A RateLimiter may be used to limit the rate at which a server is polled for
// replication status updates.  It receives the last valid ReplicationInfo, and
// the last error returned by the Update method, and should return a duration to
// wait until the next Update call is executed.
type RateLimiter func(status *ReplicationInfo, err error) time.Duration

// DefaultRateLimiter implements a rate limit of up to 2.5 seconds.
var DefaultRateLimiter = ConstantRateLimiter(500 * time.Millisecond)

// ConstantRateLimiter returns a RateLimiter that always returns the same delay.
func ConstantRateLimiter(delay time.Duration) RateLimiter {
	return RateLimiter(func(_ *ReplicationInfo, _ error) time.Duration {
		return delay
	})
}

// Replication represents a CouchDB replication process.
type Replication struct {
	Source string
	Target string

	info      *driver.ReplicationInfo
	statusErr error
	irep      driver.Replication

	rateLimitFunc RateLimiter
	// rateLimitChan is closed when the rate limit has expired.
	rateLimitTimer *time.Timer
	// This mutex protects both rateLimitFunc setting, and rateLimitChan
	rateLimitMU sync.Mutex
}

func newReplication(rep driver.Replication) *Replication {
	r := &Replication{
		Source: rep.Source(),
		Target: rep.Target(),
		irep:   rep,
	}
	return r
}

// ReplicationID returns the _replication_id field of the replicator document.
func (r *Replication) ReplicationID() string {
	return r.irep.ReplicationID()
}

// SetRateLimiter sets a rate limit function to prevent fast polling of the
// server for replication status updates. This allows calling Update() in a
// tight loop, without worrying about hitting the server too fast.
//
// Example:
//
//  rep, _ := db.Replicate("target","source")
//  rep.SetRateLimiter(ConstantRateLimiter(500 * time.Millisecond))
//  for {
//      if err := rep.Update(); err != nil {
//          break
//      }
//      // Push update status to UI...
//  }
func (r *Replication) SetRateLimiter(fn RateLimiter) {
	r.rateLimitMU.Lock()
	defer r.rateLimitMU.Unlock()
	r.rateLimitFunc = fn
}

// StartTime returns the replication start time, once the replication has been
// triggered.
func (r *Replication) StartTime() time.Time {
	return r.irep.StartTime()
}

// EndTime returns the replication end time, once the replication has terminated.
func (r *Replication) EndTime() time.Time {
	return r.irep.EndTime()
}

// State returns the current replication state
func (r *Replication) State() ReplicationState {
	return ReplicationState(r.irep.State())
}

// Err returns the error, if any, that caused the replication to abort.
func (r *Replication) Err() error {
	return r.irep.Err()
}

// IsActive returns true if the replication has not yet completed or
// errored.
func (r *Replication) IsActive() bool {
	return r.State() != ReplicationError && r.State() != ReplicationComplete
}

// Delete deletes a replication. If it is currently running, it will be
// cancelled.
func (r *Replication) Delete(ctx context.Context) error {
	return r.irep.Delete(ctx)
}

func (r *Replication) rateLimitWait() {
	if r.rateLimitFunc != nil {
		r.rateLimitMU.Lock()
		// if the timer isn't set, it means this is the first request, so no
		// rate limiting is in effect yet.
		if r.rateLimitTimer != nil {
			<-r.rateLimitTimer.C
		}
	}
}

func (r *Replication) resetRateLimit(info driver.ReplicationInfo, err error) {
	if r.rateLimitFunc != nil {
		defer r.rateLimitMU.Unlock()
		kivikInfo := ReplicationInfo(info)
		delay := r.rateLimitFunc(&kivikInfo, err)
		r.rateLimitTimer = time.NewTimer(delay)
	}
}

// Update requests a replication state update from the server. If there is an
// error retrieving the update, it is returned and the replication state is
// unaltered.
func (r *Replication) Update(ctx context.Context) error {
	r.rateLimitWait()
	var info driver.ReplicationInfo
	defer r.resetRateLimit(info, r.statusErr)
	r.statusErr = r.irep.Update(ctx, &info)
	if r.statusErr != nil {
		return r.statusErr
	}
	r.info = &info
	return nil
}

// GetReplications returns a list of defined replications in the _replicator
// database. Options are in the same format as to AllDocs(), except that
// "conflicts" and "update_seq" are ignored.
func (c *Client) GetReplications(ctx context.Context, options ...Options) ([]*Replication, error) {
	if replicator, ok := c.driverClient.(driver.ClientReplicator); ok {
		opts, err := mergeOptions(options...)
		if err != nil {
			return nil, err
		}
		reps, err := replicator.GetReplications(ctx, opts)
		if err != nil {
			return nil, err
		}
		replications := make([]*Replication, len(reps))
		for i, rep := range reps {
			replications[i] = newReplication(rep)
		}
		return replications, nil
	}
	return nil, ErrNotImplemented
}

// Replicate initiates a replication from source to target.
func (c *Client) Replicate(ctx context.Context, targetDSN, sourceDSN string, options ...Options) (*Replication, error) {
	if replicator, ok := c.driverClient.(driver.ClientReplicator); ok {
		opts, err := mergeOptions(options...)
		if err != nil {
			return nil, err
		}
		rep, err := replicator.Replicate(ctx, targetDSN, sourceDSN, opts)
		if err != nil {
			return nil, err
		}
		return newReplication(rep), nil
	}
	return nil, ErrNotImplemented
}

// ReplicationInfo represents a snapshot of the status of a replication.
type ReplicationInfo struct {
	DocWriteFailures int64
	DocsRead         int64
	DocsWritten      int64
	Progress         float64
}
