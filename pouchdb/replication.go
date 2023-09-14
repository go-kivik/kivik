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
	"net/http"
	"sync"
	"time"

	"github.com/gopherjs/gopherjs/js"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/pouchdb/bindings"
)

type replication struct {
	source    string
	target    string
	startTime time.Time
	endTime   time.Time
	state     kivik.ReplicationState
	err       error

	// mu protects the above values
	mu sync.RWMutex

	client *client
	rh     *replicationHandler
}

var _ driver.Replication = &replication{}

func (c *client) newReplication(target, source string, rep *js.Object) *replication {
	r := &replication{
		target: target,
		source: source,
		rh:     newReplicationHandler(rep),
		client: c,
	}
	c.replicationsMU.Lock()
	defer c.replicationsMU.Unlock()
	c.replications = append(c.replications, r)
	return r
}

func (r *replication) readLock() func() {
	r.mu.RLock()
	return r.mu.RUnlock
}

func (r *replication) Metadata() driver.ReplicationMetadata {
	return driver.ReplicationMetadata{
		ID:        "",
		Source:    r.source,
		Target:    r.target,
		StartTime: r.startTime,
		EndTime:   r.endTime,
		State:     string(r.state),
	}
}

func (r *replication) Err() error { defer r.readLock()(); return r.err }

func (r *replication) Update(_ context.Context, state *driver.ReplicationInfo) (err error) {
	defer bindings.RecoverError(&err)
	r.mu.Lock()
	defer r.mu.Unlock()
	event, info, err := r.rh.Status()
	if err != nil {
		return err
	}
	switch event {
	case bindings.ReplicationEventDenied, bindings.ReplicationEventError:
		r.state = kivik.ReplicationError
		r.err = bindings.NewPouchError(info.Object)
	case bindings.ReplicationEventComplete:
		r.state = kivik.ReplicationComplete
	case bindings.ReplicationEventPaused, bindings.ReplicationEventChange, bindings.ReplicationEventActive:
		r.state = kivik.ReplicationStarted
	}
	if info != nil {
		startTime, endTime := info.StartTime(), info.EndTime()
		if r.startTime.IsZero() && !startTime.IsZero() {
			r.startTime = startTime
		}
		if r.endTime.IsZero() && !endTime.IsZero() {
			r.endTime = endTime
		}
		if r.rh.state != nil {
			state.DocWriteFailures = r.rh.state.DocWriteFailures
			state.DocsRead = r.rh.state.DocsRead
			state.DocsWritten = r.rh.state.DocsWritten
		}
	}
	return nil
}

func (r *replication) Delete(context.Context) (err error) {
	defer bindings.RecoverError(&err)
	r.rh.Cancel()
	r.client.replicationsMU.Lock()
	defer r.client.replicationsMU.Unlock()
	for i, rep := range r.client.replications {
		if rep == r {
			last := len(r.client.replications) - 1
			r.client.replications[i] = r.client.replications[last]
			r.client.replications[last] = nil
			r.client.replications = r.client.replications[:last]
			return nil
		}
	}
	return &kivik.Error{Status: http.StatusNotFound, Message: "replication not found"}
}

func replicationEndpoint(dsn string, object interface{}) (name string, obj interface{}, err error) {
	defer bindings.RecoverError(&err)
	if object == nil {
		return dsn, dsn, nil
	}
	switch t := object.(type) {
	case *js.Object:
		tx := object.(*js.Object) // https://github.com/gopherjs/gopherjs/issues/682
		// Assume it's a raw PouchDB object
		return tx.Get("name").String(), tx, nil
	case *bindings.DB:
		// Unwrap the bare object
		return t.Object.Get("name").String(), t.Object, nil
	}
	// Just let it pass through
	return "<unknown>", obj, nil
}

func (c *client) Replicate(_ context.Context, targetDSN, sourceDSN string, options driver.Options) (driver.Replication, error) {
	pouchOpts := map[string]interface{}{}
	options.Apply(pouchOpts)
	// Allow overriding source and target with options, i.e. for PouchDB objects
	sourceName, sourceObj, err := replicationEndpoint(sourceDSN, pouchOpts["source"])
	if err != nil {
		return nil, err
	}
	targetName, targetObj, err := replicationEndpoint(targetDSN, pouchOpts["target"])
	if err != nil {
		return nil, err
	}
	delete(pouchOpts, "source")
	delete(pouchOpts, "target")
	rep, err := c.pouch.Replicate(sourceObj, targetObj, pouchOpts)
	if err != nil {
		return nil, err
	}
	return c.newReplication(targetName, sourceName, rep), nil
}

func (c *client) GetReplications(context.Context, driver.Options) ([]driver.Replication, error) {
	c.replicationsMU.RLock()
	defer c.replicationsMU.RUnlock()
	reps := make([]driver.Replication, len(c.replications))
	for i, rep := range c.replications {
		reps[i] = rep
	}
	return reps, nil
}
