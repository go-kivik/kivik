package pouchdb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/pouchdb/bindings"
	"github.com/flimzy/kivik/errors"
	"github.com/gopherjs/gopherjs/js"
)

type replication struct {
	source    string
	target    string
	startTime time.Time
	endTime   time.Time
	state     string
	err       error

	// mu protects the above values
	mu sync.RWMutex

	client    *client
	jsObj     *js.Object
	lastEvent string
	lastState *replicationState
	lastError error
}

var _ driver.Replication = &replication{}

func (c *client) newReplication(target, source string, rep *js.Object) *replication {
	r := &replication{
		target: target,
		source: source,
		jsObj:  rep,
	}
	rep.Call("on", "change", func(info *js.Object) {
		r.handleUpdate("change", &replicationState{Object: info}, nil)
	})
	rep.Call("on", "paused", func(err *js.Object) {
		r.state = string(kivik.ReplicationNotStarted)
	})
	rep.Call("on", "active", func() {
		r.state = string(kivik.ReplicationStarted)
	})
	rep.Call("on", "denied", func(err *js.Object) {
		r.handleUpdate("denied", nil, fmt.Errorf("%v", err))
	})
	rep.Call("on", "complete", func(info *js.Object) {
		r.handleUpdate("complete", &replicationState{Object: info}, nil)
	})
	rep.Call("on", "error", func(err *js.Object) {
		r.handleUpdate("error", nil, fmt.Errorf("%v", err))
	})
	c.replicationsMU.Lock()
	defer c.replicationsMU.Unlock()
	c.replications = append(c.replications, r)
	return r
}

func (r *replication) handleUpdate(event string, state *replicationState, err error) {
	r.lastEvent = event
	r.lastError = err
	if state != nil {
		r.lastState = state
	}
}

type replicationState struct {
	*js.Object
	StartTime        time.Time `js:"start_time"`
	EndTime          time.Time `js:"end_time"`
	DocsRead         int64     `js:"docs_read"`
	DocsWritten      int64     `js:"docs_written"`
	DocWriteFailures int64     `js:"doc_write_failures"`
	Status           string    `js:"status"`
	LastSeq          string    `js:"last_seq"`
}

func (r *replication) readLock() func() {
	r.mu.RLock()
	return r.mu.RUnlock
}

func (r *replication) ReplicationID() string { return "" }
func (r *replication) Source() string        { defer r.readLock()(); return r.source }
func (r *replication) Target() string        { defer r.readLock()(); return r.target }
func (r *replication) StartTime() time.Time  { defer r.readLock()(); return r.startTime }
func (r *replication) EndTime() time.Time    { defer r.readLock()(); return r.endTime }
func (r *replication) State() string         { defer r.readLock()(); return r.state }
func (r *replication) Err() error            { defer r.readLock()(); return r.err }

func (r *replication) Update(ctx context.Context, state *driver.ReplicationInfo) (err error) {
	defer bindings.RecoverError(&err)
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.lastError != nil {
		r.err = r.lastError
	}
	switch r.lastEvent {
	case "":
		return nil
	case "denied", "error":
		r.state = string(kivik.ReplicationError)
	case "complete":
		r.state = string(kivik.ReplicationComplete)
	case "change":
		r.state = string(kivik.ReplicationStarted)
	default:
		return fmt.Errorf("Unexpected event: %s", r.lastEvent)
	}
	if !r.lastState.StartTime.IsZero() && r.startTime.IsZero() {
		r.startTime = r.lastState.StartTime
	}
	if !r.lastState.EndTime.IsZero() && r.endTime.IsZero() {
		r.endTime = r.lastState.EndTime
	}
	r.lastState = nil
	return nil
}

func (r *replication) Delete(ctx context.Context) (err error) {
	defer bindings.RecoverError(&err)
	r.jsObj.Call("cancel")
	r.client.replicationsMU.Lock()
	defer r.client.replicationsMU.Unlock()
	for i, rep := range r.client.replications {
		if rep == r {
			last := len(r.client.replications) - 1
			r.client.replications[i] = r.client.replications[last]
			r.client.replications[last] = nil
			r.client.replications = r.client.replications[:last]
		}
	}
	return errors.Status(kivik.StatusNotFound, "replication not found")
}

func replicationEndpoint(dsn string, object interface{}) (name string, obj interface{}, err error) {
	defer bindings.RecoverError(&err)
	if object == nil {
		return dsn, dsn, nil
	}
	switch t := object.(type) {
	case *js.Object:
		// Assume it's a raw PouchDB object
		return t.Get("name").String(), t, nil
	case *bindings.DB:
		// Unwrap the bare object
		return t.Object.Get("name").String(), t.Object, nil
	}
	// Just let it pass through
	return "<unknown>", obj, nil
}

func (c *client) Replicate(_ context.Context, targetDSN, sourceDSN string, options map[string]interface{}) (driver.Replication, error) {
	opts, err := c.options(options)
	if err != nil {
		return nil, err
	}
	// Allow overriding source and target with options, i.e. for PouchDB objects
	sourceName, sourceObj, err := replicationEndpoint(sourceDSN, opts["source"])
	if err != nil {
		return nil, err
	}
	targetName, targetObj, err := replicationEndpoint(targetDSN, opts["target"])
	if err != nil {
		return nil, err
	}
	delete(opts, "source")
	delete(opts, "target")
	rep, err := c.pouch.Replicate(sourceObj, targetObj, opts)
	if err != nil {
		return nil, err
	}
	return c.newReplication(targetName, sourceName, rep), nil
}

func (c *client) GetReplications(_ context.Context, options map[string]interface{}) ([]driver.Replication, error) {
	for range options {
		return nil, errors.Status(kivik.StatusBadRequest, "options not yet supported")
	}
	c.replicationsMU.RLock()
	defer c.replicationsMU.RUnlock()
	reps := make([]driver.Replication, len(c.replications))
	for i, rep := range c.replications {
		reps[i] = rep
	}
	return reps, nil
}
