package pouchdb

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/flimzy/kivik/driver/pouchdb/bindings"
	"github.com/gopherjs/gopherjs/js"
)

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

type replicationHandler struct {
	event *string
	state *replicationState
	err   error

	mu       sync.Mutex
	wg       sync.WaitGroup
	complete bool
	obj      *js.Object
}

func (r *replicationHandler) Cancel() {
	r.obj.Call("cancel")
}

// Status returns the last-read status. If the last-read status was already read,
// this blocks until the next event.  If the replication is complete, it will
// return io.EOF immediately.
func (r *replicationHandler) Status() (string, *replicationState, error) {
	if r.complete && r.event == nil {
		return "", nil, io.EOF
	}
	r.mu.Lock()
	if r.event == nil {
		r.mu.Unlock()
		// Wait for an event to be ready to read
		r.wg.Wait()
		r.mu.Lock()
	}
	event, state, err := r.event, r.state, r.err
	r.event = nil
	r.mu.Unlock()
	r.wg.Add(1)
	return *event, state, err
}

func (r *replicationHandler) handleEvent(event string, info *replicationState, err error) {
	if r.complete {
		panic(fmt.Sprintf("Unexpected replication event after complete. %v %v %v", event, info, err))
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.event = &event
	switch event {
	case bindings.ReplicationEventDenied, bindings.ReplicationEventError, bindings.ReplicationEventComplete:
		r.complete = true
	}
	switch event {
	case bindings.ReplicationEventDenied, bindings.ReplicationEventError:
		r.err = err
	}
	if r.state == nil {
		r.state = info
	}
	r.wg.Done()
}

func newReplicationHandler(rep *js.Object) *replicationHandler {
	r := &replicationHandler{obj: rep}
	rep.Call("on", bindings.ReplicationEventChange, func(info *js.Object) {
		r.handleEvent(bindings.ReplicationEventChange, &replicationState{Object: info}, nil)
	})
	rep.Call("on", bindings.ReplicationEventComplete, func(info *js.Object) {
		r.handleEvent(bindings.ReplicationEventComplete, &replicationState{Object: info}, nil)
	})
	rep.Call("on", bindings.ReplicationEventPaused, func(err *js.Object) {
		r.handleEvent(bindings.ReplicationEventPaused, nil, &js.Error{Object: err})
	})
	rep.Call("on", bindings.ReplicationEventActive, func(_ *js.Object) {
		r.handleEvent(bindings.ReplicationEventActive, nil, nil)
	})
	rep.Call("on", bindings.ReplicationEventDenied, func(err *js.Object) {
		r.handleEvent(bindings.ReplicationEventDenied, nil, &js.Error{Object: err})
	})
	rep.Call("on", bindings.ReplicationEventError, func(err *js.Object) {
		r.handleEvent(bindings.ReplicationEventError, nil, &js.Error{Object: err})
	})
	r.wg.Add(1)
	return r
}
