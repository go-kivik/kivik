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
	"net/http"
	"sync"
	"time"

	"github.com/go-kivik/kivik/v4/driver"
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

// The additional possible values for the state field in the _scheduler docs.
const (
	ReplicationInitializing ReplicationState = "initializing"
	ReplicationRunning      ReplicationState = "running"
	ReplicationPending      ReplicationState = "pending"
	ReplicationCrashing     ReplicationState = "crashing"
	ReplicationFailed       ReplicationState = "failed"
)

// Replication represents a CouchDB replication process.
type Replication struct {
	meta      driver.ReplicationMetadata
	infoMU    sync.RWMutex
	info      *driver.ReplicationInfo
	statusErr error
	irep      driver.Replication
}

// Source is the URL to the replication source.
func (r *Replication) Source() string {
	return r.meta.Source
}

// Source is the URL to the replication target.
func (r *Replication) Target() string {
	return r.meta.Target
}

// DocsWritten returns the number of documents written, if known.
func (r *Replication) DocsWritten() int64 {
	if r != nil && r.info != nil {
		r.infoMU.RLock()
		defer r.infoMU.RUnlock()
		return r.info.DocsWritten
	}
	return 0
}

// DocsRead returns the number of documents read, if known.
func (r *Replication) DocsRead() int64 {
	if r != nil && r.info != nil {
		r.infoMU.RLock()
		defer r.infoMU.RUnlock()
		return r.info.DocsRead
	}
	return 0
}

// DocWriteFailures returns the number of doc write failures, if known.
func (r *Replication) DocWriteFailures() int64 {
	if r != nil && r.info != nil {
		r.infoMU.RLock()
		defer r.infoMU.RUnlock()
		return r.info.DocWriteFailures
	}
	return 0
}

// Progress returns the current replication progress, if known.
func (r *Replication) Progress() float64 {
	if r != nil && r.info != nil {
		r.infoMU.RLock()
		defer r.infoMU.RUnlock()
		return r.info.Progress
	}
	return 0
}

func newReplication(rep driver.Replication) *Replication {
	return &Replication{
		meta: rep.Metadata(),
		irep: rep,
	}
}

// ReplicationID returns the _replication_id field of the replicator document.
func (r *Replication) ReplicationID() string {
	return r.meta.ID
}

// StartTime returns the replication start time, once the replication has been
// triggered.
func (r *Replication) StartTime() time.Time {
	return r.meta.StartTime
}

// EndTime returns the replication end time, once the replication has terminated.
func (r *Replication) EndTime() time.Time {
	return r.meta.EndTime
}

// State returns the current replication state
func (r *Replication) State() ReplicationState {
	return ReplicationState(r.meta.State)
}

// Err returns the error, if any, that caused the replication to abort.
func (r *Replication) Err() error {
	if r == nil {
		return nil
	}
	return r.meta.Error
}

// IsActive returns true if the replication has not yet completed or
// errored.
func (r *Replication) IsActive() bool {
	if r == nil {
		return false
	}
	switch r.State() {
	case ReplicationError, ReplicationComplete, ReplicationCrashing, ReplicationFailed:
		return false
	default:
		return true
	}
}

// Delete deletes a replication. If it is currently running, it will be
// cancelled.
func (r *Replication) Delete(ctx context.Context) error {
	return r.irep.Delete(ctx)
}

// Update requests a replication state update from the server. If there is an
// error retrieving the update, it is returned and the replication state is
// unaltered.
func (r *Replication) Update(ctx context.Context) error {
	var info driver.ReplicationInfo
	r.statusErr = r.irep.Update(ctx, &info)
	if r.statusErr != nil {
		return r.statusErr
	}
	r.infoMU.Lock()
	r.info = &info
	r.infoMU.Unlock()
	return nil
}

var replicationNotImplemented = &Error{Status: http.StatusNotImplemented, Message: "kivik: driver does not support replication"}

// GetReplications returns a list of defined replications in the _replicator
// database. Options are in the same format as to [DB.AllDocs], except that
// "conflicts" and "update_seq" are ignored.
func (c *Client) GetReplications(ctx context.Context, options ...Option) ([]*Replication, error) {
	if err := c.startQuery(); err != nil {
		return nil, err
	}
	defer c.endQuery()
	replicator, ok := c.driverClient.(driver.ClientReplicator)
	if !ok {
		return nil, replicationNotImplemented
	}
	reps, err := replicator.GetReplications(ctx, allOptions(options))
	if err != nil {
		return nil, err
	}
	replications := make([]*Replication, len(reps))
	for i, rep := range reps {
		replications[i] = newReplication(rep)
	}
	return replications, nil
}

// Replicate initiates a replication from source to target.
//
// To use an object for either "source" or "target", pass the desired object
// in options. This will override targetDSN and sourceDSN function parameters.
func (c *Client) Replicate(ctx context.Context, targetDSN, sourceDSN string, options ...Option) (*Replication, error) {
	if err := c.startQuery(); err != nil {
		return nil, err
	}
	defer c.endQuery()
	replicator, ok := c.driverClient.(driver.ClientReplicator)
	if !ok {
		return nil, replicationNotImplemented
	}
	rep, err := replicator.Replicate(ctx, targetDSN, sourceDSN, allOptions(options))
	if err != nil {
		return nil, err
	}
	return newReplication(rep), nil
}

// ReplicationInfo represents a snapshot of the status of a replication.
type ReplicationInfo struct {
	DocWriteFailures int64
	DocsRead         int64
	DocsWritten      int64
	Progress         float64
}
