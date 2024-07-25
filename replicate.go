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
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
)

// ReplicationResult represents the result of a replication.
type ReplicationResult struct {
	DocWriteFailures int       `json:"doc_write_failures"`
	DocsRead         int       `json:"docs_read"`
	DocsWritten      int       `json:"docs_written"`
	EndTime          time.Time `json:"end_time"`
	MissingChecked   int       `json:"missing_checked"`
	MissingFound     int       `json:"missing_found"`
	StartTime        time.Time `json:"start_time"`
}

const (
	eventSecurity = "security"
	eventChanges  = "changes"
	eventChange   = "change"
	eventRevsDiff = "revsdiff"
	eventDocument = "document"
)

// ReplicationEvent is an event emitted by the Replicate function, which
// represents a single read or write event, and its status.
type ReplicationEvent struct {
	// Type is the event type. Options are:
	//
	// - "security" -- Relates to the _security document.
	// - "changes"  -- Relates to the changes feed.
	// - "change"   -- Relates to a single change.
	// - "revsdiff" -- Relates to reading the revs diff.
	// - "document" -- Relates to a specific document.
	Type string
	// Read is true if the event relates to a read operation.
	Read bool
	// DocID is the relevant document ID, if any.
	DocID string
	// Error is the error associated with the event, if any.
	Error error
	// Changes is the list of changed revs, for a "change" event.
	Changes []string
}

// eventCallback is a function that receives replication events.
type eventCallback func(ReplicationEvent)

func (c eventCallback) Apply(target interface{}) {
	if r, ok := target.(*replicator); ok {
		r.cb = c
	}
}

// ReplicateCallback sets a callback function to be called on every replication
// event that takes place.
func ReplicateCallback(callback func(ReplicationEvent)) Option {
	return eventCallback(callback)
}

type replicateCopySecurityOption struct{}

func (r replicateCopySecurityOption) Apply(target interface{}) {
	if r, ok := target.(*replicator); ok {
		r.withSecurity = true
	}
}

// ReplicateCopySecurity will read the security object from source, and copy it
// to the target, before the replication. Use with caution! The security object
// is not versioned, and it will be unconditionally overwritten on the target!
func ReplicateCopySecurity() Option {
	return replicateCopySecurityOption{}
}

// Replicate performs a replication from source to target, using a limited
// version of the CouchDB replication protocol.
//
// This function supports the [ReplicateCopySecurity] and [ReplicateCallback]
// options. Additionally, the following standard options are passed along to
// the source when querying the changes feed, for server-side filtering, where
// supported:
//
//	filter (string)           - The name of a filter function.
//	doc_ids (array of string) - Array of document IDs to be synchronized.
func Replicate(ctx context.Context, target, source *DB, options ...Option) (*ReplicationResult, error) {
	opts := multiOptions(options)

	r := newReplicator(target, source)
	opts.Apply(r)
	err := r.replicate(ctx, opts)
	return r.result(), err
}

func (r *replicator) replicate(ctx context.Context, options Option) error {
	if err := r.copySecurity(ctx); err != nil {
		return err
	}

	group, ctx := errgroup.WithContext(ctx)
	changes := make(chan *change)
	group.Go(func() error {
		defer close(changes)
		return r.readChanges(ctx, changes, options)
	})

	diffs := make(chan *revDiff)
	group.Go(func() error {
		defer close(diffs)
		return r.readDiffs(ctx, changes, diffs)
	})

	docs := make(chan *document)
	group.Go(func() error {
		defer close(docs)
		return r.readDocs(ctx, diffs, docs)
	})

	group.Go(func() error {
		return r.storeDocs(ctx, docs)
	})

	return group.Wait()
}

// replicator manages a single replication.
type replicator struct {
	target, source *DB
	cb             eventCallback
	// withSecurity indicates that the security object should be read from
	// source, and copied to the target, before the replication. Use with
	// caution! The security object is not versioned, and will be
	// unconditionally overwritten!
	withSecurity bool
	// noOpenRevs is set if a call to OpenRevs returns unsupported
	noOpenRevs bool
	start      time.Time
	// replication stats counters
	writeFailures, reads, writes, missingChecks, missingFound int32
}

func newReplicator(target, source *DB) *replicator {
	return &replicator{
		target: target,
		source: source,
		start:  time.Now(),
	}
}

func (r *replicator) callback(e ReplicationEvent) {
	if r.cb == nil {
		return
	}
	r.cb(e)
}

func (r *replicator) result() *ReplicationResult {
	return &ReplicationResult{
		StartTime:        r.start,
		EndTime:          time.Now(),
		DocWriteFailures: int(r.writeFailures),
		DocsRead:         int(r.reads),
		DocsWritten:      int(r.writes),
		MissingChecked:   int(r.missingChecks),
		MissingFound:     int(r.missingFound),
	}
}

func (r *replicator) copySecurity(ctx context.Context) error {
	if !r.withSecurity {
		return nil
	}
	sec, err := r.source.Security(ctx)
	r.callback(ReplicationEvent{
		Type:  eventSecurity,
		Read:  true,
		Error: err,
	})
	if err != nil {
		return fmt.Errorf("read security: %w", err)
	}
	err = r.target.SetSecurity(ctx, sec)
	r.callback(ReplicationEvent{
		Type:  eventSecurity,
		Read:  false,
		Error: err,
	})
	if err != nil {
		return fmt.Errorf("set security: %w", err)
	}
	return nil
}

type change struct {
	ID      string
	Changes []string
}

// readChanges reads the changes feed.
//
// https://docs.couchdb.org/en/stable/replication/protocol.html#listen-to-changes-feed
func (r *replicator) readChanges(ctx context.Context, results chan<- *change, options Option) error {
	changes := r.source.Changes(ctx, options, Param("feed", "normal"), Param("style", "all_docs"))
	r.callback(ReplicationEvent{
		Type: eventChanges,
		Read: true,
	})

	defer changes.Close() // nolint: errcheck
	for changes.Next() {
		ch := &change{
			ID:      changes.ID(),
			Changes: changes.Changes(),
		}
		r.callback(ReplicationEvent{
			Type:    eventChange,
			DocID:   ch.ID,
			Read:    true,
			Changes: ch.Changes,
		})
		select {
		case <-ctx.Done():
			return ctx.Err()
		case results <- ch:
		}
	}
	if err := changes.Err(); err != nil {
		r.callback(ReplicationEvent{
			Type:  eventChanges,
			Read:  true,
			Error: err,
		})
		return fmt.Errorf("read changes feed: %w", err)
	}
	return nil
}

type revDiff struct {
	ID                string   `json:"-"`
	Missing           []string `json:"missing"`
	PossibleAncestors []string `json:"possible_ancestors"`
}

const rdBatchSize = 10

// readDiffs reads the diffs for the reported changes.
//
// https://docs.couchdb.org/en/stable/replication/protocol.html#calculate-revision-difference
func (r *replicator) readDiffs(ctx context.Context, ch <-chan *change, results chan<- *revDiff) error {
	for {
		revMap := map[string][]string{}
		var change *change
		var ok bool
	loop:
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case change, ok = <-ch:
				if !ok {
					break loop
				}
				revMap[change.ID] = change.Changes
				if len(revMap) >= rdBatchSize {
					break loop
				}
			}
		}

		if len(revMap) == 0 {
			return nil
		}
		diffs := r.target.RevsDiff(ctx, revMap)
		err := diffs.Err()
		r.callback(ReplicationEvent{
			Type:  eventRevsDiff,
			Read:  true,
			Error: err,
		})
		if err != nil {
			return err
		}
		defer diffs.Close() // nolint: errcheck
		for diffs.Next() {
			var val revDiff
			if err := diffs.ScanValue(&val); err != nil {
				r.callback(ReplicationEvent{
					Type:  eventRevsDiff,
					Read:  true,
					Error: err,
				})
				return err
			}
			val.ID, _ = diffs.ID()
			r.callback(ReplicationEvent{
				Type:  eventRevsDiff,
				Read:  true,
				DocID: val.ID,
			})
			select {
			case <-ctx.Done():
				return ctx.Err()
			case results <- &val:
			}
		}
		if err := diffs.Err(); err != nil {
			r.callback(ReplicationEvent{
				Type:  eventRevsDiff,
				Read:  true,
				Error: err,
			})
			return fmt.Errorf("read revs diffs: %w", err)
		}
	}
}

// readDocs reads the document revisions that have changed between source and
// target.
//
// https://docs.couchdb.org/en/stable/replication/protocol.html#fetch-changed-documents
func (r *replicator) readDocs(ctx context.Context, diffs <-chan *revDiff, results chan<- *document) error {
	for {
		var rd *revDiff
		var ok bool
		select {
		case <-ctx.Done():
			return ctx.Err()
		case rd, ok = <-diffs:
			if !ok {
				return nil
			}
			if err := r.readDoc(ctx, rd.ID, rd.Missing, results); err != nil {
				return err
			}
		}
	}
}

func (r *replicator) readDoc(ctx context.Context, id string, revs []string, results chan<- *document) error {
	if !r.noOpenRevs {
		err := r.readOpenRevs(ctx, id, revs, results)
		if HTTPStatus(err) == http.StatusNotImplemented {
			r.noOpenRevs = true
		} else {
			return err
		}
	}
	return r.readIndividualDocs(ctx, id, revs, results)
}

func (r *replicator) readOpenRevs(ctx context.Context, id string, revs []string, results chan<- *document) error {
	rs := r.source.OpenRevs(ctx, id, revs, Params(map[string]interface{}{
		"revs":   true,
		"latest": true,
	}))
	defer rs.Close()
	for rs.Next() {
		atomic.AddInt32(&r.reads, 1)
		atomic.AddInt32(&r.missingFound, 1)
		doc := new(document)
		err := rs.ScanDoc(&doc)
		if err != nil {
			return err
		}
		r.callback(ReplicationEvent{
			Type:  eventDocument,
			Read:  true,
			DocID: id,
			Error: err,
		})
		atts, _ := rs.Attachments()
		if err := prepareAttachments(doc, atts); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case results <- doc:
		}
	}
	err := rs.Err()
	if err == nil {
		atomic.AddInt32(&r.missingChecks, int32(len(revs)))
	}
	return err
}

func (r *replicator) readIndividualDocs(ctx context.Context, id string, revs []string, results chan<- *document) error {
	for _, rev := range revs {
		atomic.AddInt32(&r.missingChecks, 1)
		d, err := readDoc(ctx, r.source, id, rev)
		r.callback(ReplicationEvent{
			Type:  eventDocument,
			Read:  true,
			DocID: id,
			Error: err,
		})
		if err != nil {
			return fmt.Errorf("read doc %s: %w", id, err)
		}
		atomic.AddInt32(&r.reads, 1)
		atomic.AddInt32(&r.missingFound, 1)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case results <- d:
		}
	}
	return nil
}

// prepareAttachments reads attachments from atts, prepares them, and adds them
// to doc.
func prepareAttachments(doc *document, atts *AttachmentsIterator) error {
	if atts == nil {
		return nil
	}
	// TODO: It seems silly this is necessary... I need better attachment
	// handling in kivik.
	for {
		att, err := atts.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		var content []byte
		switch att.ContentEncoding {
		case "":
			var err error
			content, err = io.ReadAll(att.Content)
			if err != nil {
				return err
			}
			if err := att.Content.Close(); err != nil {
				return err
			}
		case "gzip":
			zr, err := gzip.NewReader(att.Content)
			if err != nil {
				return err
			}
			content, err = io.ReadAll(zr)
			if err != nil {
				return err
			}
			if err := zr.Close(); err != nil {
				return err
			}
			if err := att.Content.Close(); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Unknown encoding '%s' for attachment '%s'", att.ContentEncoding, att.Filename)
		}
		att.Stub = false
		att.Follows = false
		att.Content = io.NopCloser(bytes.NewReader(content))
		doc.Attachments.Set(att.Filename, att)
	}
}

func readDoc(ctx context.Context, db *DB, docID, rev string) (*document, error) {
	doc := new(document)
	row := db.Get(ctx, docID, Params(map[string]interface{}{
		"rev":         rev,
		"revs":        true,
		"attachments": true,
	}))
	if err := row.ScanDoc(&doc); err != nil {
		return nil, err
	}
	atts, _ := row.Attachments()
	if err := prepareAttachments(doc, atts); err != nil {
		return nil, err
	}

	return doc, nil
}

// storeDocs updates the changed documents.
//
// https://docs.couchdb.org/en/stable/replication/protocol.html#upload-batch-of-changed-documents
func (r *replicator) storeDocs(ctx context.Context, docs <-chan *document) error {
	for doc := range docs {
		_, err := r.target.Put(ctx, doc.ID, doc, Param("new_edits", false))
		r.callback(ReplicationEvent{
			Type:  "document",
			Read:  false,
			DocID: doc.ID,
			Error: err,
		})
		if err != nil {
			atomic.AddInt32(&r.writeFailures, 1)
			return fmt.Errorf("store doc %s: %w", doc.ID, err)
		}
		atomic.AddInt32(&r.writes, 1)
	}
	return nil
}
