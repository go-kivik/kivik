package kivik

import (
	"context"
	"time"

	"github.com/flimzy/kivik/driver"
)

// Replication represents an active or completed replication.
type Replication struct {
	// Constant values

	ReplicationID string    // Available immediately
	Source        string    // "
	Target        string    // "
	StartTime     time.Time // "
	SourceSeq     string

	// Values updated with Update()

	DocWriteFailures int64
	DocsRead         int64
	DocsWritten      int64
	Progress         float64
	LastUpdate       time.Time
	EndTime          time.Time
	Status           string
	lastError        error

	irep driver.Replication
	done bool
}

// Err returns the error, if any, that caused the replication to abort.
func (r *Replication) Err() error {
	return r.lastError
}

func newReplication(rep driver.Replication) *Replication {
	return &Replication{
		ReplicationID: rep.ReplicationID(),
		StartTime:     rep.StartTime(),
		Source:        rep.Source(),
		Target:        rep.Target(),
		irep:          rep,
	}
}

// Delete deletes a replication. If it is currently running, it will be
// cancelled.
func (r *Replication) Delete(ctx context.Context) error {
	if err := r.irep.Delete(ctx); err != nil {
		return err
	}
	r.done = true
	return nil
}

// Update requests a replication state update from the server. If there is an
// error retrieving the update, it is returned and the replication state is
// unaltered.
func (r *Replication) Update(ctx context.Context) error {
	var rep driver.ReplicationState
	if err := r.irep.Update(ctx, &rep); err != nil {
		return err
	}
	r.DocWriteFailures = rep.DocWriteFailures
	r.DocsRead = rep.DocsRead
	r.DocsWritten = rep.DocsWritten
	r.Progress = rep.Progress
	r.LastUpdate = rep.LastUpdate
	r.EndTime = rep.EndTime
	r.SourceSeq = rep.SourceSeq
	r.Status = rep.Status
	r.lastError = rep.Error
	if rep.Status == "complete" || rep.Status == "error" || r.lastError != nil {
		r.done = true
	}
	return nil
}

// Active returns true if the replication is still active. Note that a
// replication can switch from inactive to active if it is restarted.
func (r *Replication) Active() bool {
	return !r.done
}

// Cancel cancels the replication.
func (r *Replication) Cancel(ctx context.Context) error {
	if err := r.irep.Cancel(ctx); err != nil {
		return err
	}
	_ = r.Update(ctx)
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
