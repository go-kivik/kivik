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
	"sync"
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

type resultWrapper struct {
	*ReplicationResult
	mu sync.Mutex
}

func (r *resultWrapper) read() {
	r.mu.Lock()
	r.DocsRead++
	r.mu.Unlock()
}

func (r *resultWrapper) missingChecked() {
	r.mu.Lock()
	r.MissingChecked++
	r.mu.Unlock()
}

func (r *resultWrapper) missingFound() {
	r.mu.Lock()
	r.MissingFound++
	r.mu.Unlock()
}

func (r *resultWrapper) writeError() {
	r.mu.Lock()
	r.DocWriteFailures++
	r.mu.Unlock()
}

func (r *resultWrapper) write() {
	r.mu.Lock()
	r.DocsWritten++
	r.mu.Unlock()
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
	if ro, ok := target.(*replicateOptions); ok {
		ro.cb = c
	}
}

// ReplicateCallback sets a callback function to be called on every replication
// event that takes place.
func ReplicateCallback(callback func(ReplicationEvent)) Option {
	return eventCallback(callback)
}

type replicateOptions struct {
	cb eventCallback
	// CopySecurity indicates that the secuurity object should be read from
	// source, and copied to the target, before the replication. Use with
	// caution! The security object is not versioned, and will be
	// unconditionally overwritten!
	copySecurity bool
}

func (o replicateOptions) callback() eventCallback {
	if o.cb != nil {
		return o.cb
	}
	return func(ReplicationEvent) {}
}

type replicateCopySecurityOption struct{}

func (r replicateCopySecurityOption) Apply(target interface{}) {
	if ro, ok := target.(*replicateOptions); ok {
		ro.copySecurity = true
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
	result := &resultWrapper{
		ReplicationResult: &ReplicationResult{
			StartTime: time.Now(),
		},
	}
	defer func() {
		result.EndTime = time.Now()
	}()
	repOpts := &replicateOptions{}
	allOptions(options).Apply(repOpts)
	cb := repOpts.callback()

	if repOpts.copySecurity {
		if err := copySecurity(ctx, target, source, cb); err != nil {
			return result.ReplicationResult, err
		}
	}
	group, ctx := errgroup.WithContext(ctx)
	changes := make(chan *change)
	group.Go(func() error {
		defer close(changes)
		return readChanges(ctx, source, changes, allOptions(options), cb)
	})

	diffs := make(chan *revDiff)
	group.Go(func() error {
		defer close(diffs)
		return readDiffs(ctx, target, changes, diffs, cb)
	})

	docs := make(chan *document)
	group.Go(func() error {
		defer close(docs)
		return readDocs(ctx, source, diffs, docs, result, cb)
	})

	group.Go(func() error {
		return storeDocs(ctx, target, docs, result, cb)
	})

	return result.ReplicationResult, group.Wait()
}

func copySecurity(ctx context.Context, target, source *DB, cb eventCallback) error {
	sec, err := source.Security(ctx)
	cb(ReplicationEvent{
		Type:  eventSecurity,
		Read:  true,
		Error: err,
	})
	if err != nil {
		return fmt.Errorf("read security: %w", err)
	}
	err = target.SetSecurity(ctx, sec)
	cb(ReplicationEvent{
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
func readChanges(ctx context.Context, db *DB, results chan<- *change, options Option, cb eventCallback) error {
	changes := db.Changes(ctx, options, Param("feed", "normal"), Param("style", "all_docs"))
	cb(ReplicationEvent{
		Type: eventChanges,
		Read: true,
	})

	defer changes.Close() // nolint: errcheck
	for changes.Next() {
		ch := &change{
			ID:      changes.ID(),
			Changes: changes.Changes(),
		}
		cb(ReplicationEvent{
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
		cb(ReplicationEvent{
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
func readDiffs(ctx context.Context, db *DB, ch <-chan *change, results chan<- *revDiff, cb eventCallback) error {
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
		diffs := db.RevsDiff(ctx, revMap)
		err := diffs.Err()
		cb(ReplicationEvent{
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
				cb(ReplicationEvent{
					Type:  eventRevsDiff,
					Read:  true,
					Error: err,
				})
				return err
			}
			val.ID, _ = diffs.ID()
			cb(ReplicationEvent{
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
			cb(ReplicationEvent{
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
func readDocs(ctx context.Context, db *DB, diffs <-chan *revDiff, results chan<- *document, result *resultWrapper, cb eventCallback) error {
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
			for _, rev := range rd.Missing {
				result.missingChecked()
				d, err := readDoc(ctx, db, rd.ID, rev)
				cb(ReplicationEvent{
					Type:  eventDocument,
					Read:  true,
					DocID: rd.ID,
					Error: err,
				})
				if err != nil {
					return fmt.Errorf("read doc %s: %w", rd.ID, err)
				}
				result.read()
				result.missingFound()
				select {
				case <-ctx.Done():
					return ctx.Err()
				case results <- d:
				}
			}
		}
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
	// TODO: It seems silly this is necessary... I need better attachment
	// handling in kivik.
	if atts, _ := row.Attachments(); atts != nil {
		for {
			att, err := atts.Next()
			if err != nil {
				if err != io.EOF {
					return nil, err
				}
				break
			}
			var content []byte
			switch att.ContentEncoding {
			case "":
				var err error
				content, err = io.ReadAll(att.Content)
				if err != nil {
					return nil, err
				}
				if err := att.Content.Close(); err != nil {
					return nil, err
				}
			case "gzip":
				zr, err := gzip.NewReader(att.Content)
				if err != nil {
					return nil, err
				}
				content, err = io.ReadAll(zr)
				if err != nil {
					return nil, err
				}
				if err := zr.Close(); err != nil {
					return nil, err
				}
				if err := att.Content.Close(); err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("Unknown encoding '%s' for attachment '%s'", att.ContentEncoding, att.Filename)
			}
			att.Stub = false
			att.Follows = false
			att.Content = io.NopCloser(bytes.NewReader(content))
			doc.Attachments.Set(att.Filename, att)
		}
	}
	return doc, nil
}

// storeDocs updates the changed documents.
//
// https://docs.couchdb.org/en/stable/replication/protocol.html#upload-batch-of-changed-documents
func storeDocs(ctx context.Context, db *DB, docs <-chan *document, result *resultWrapper, cb eventCallback) error {
	for doc := range docs {
		_, err := db.Put(ctx, doc.ID, doc, Param("new_edits", false))
		cb(ReplicationEvent{
			Type:  "document",
			Read:  false,
			DocID: doc.ID,
			Error: err,
		})
		if err != nil {
			result.writeError()
			return fmt.Errorf("store doc %s: %w", doc.ID, err)
		}
		result.write()
	}
	return nil
}
