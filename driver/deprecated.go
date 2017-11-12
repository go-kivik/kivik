package driver

import (
	"context"
	"encoding/json"
	"io"
)

// OldDB is deprecated and will be removed in Kivik 2.0. Please use DB instead.
type OldDB interface {
	// AllDocs returns all of the documents in the database, subject to the
	// options provided.
	AllDocs(ctx context.Context, options map[string]interface{}) (Rows, error)
	// Get fetches the requested document from the database, and unmarshals it
	// into doc.
	Get(ctx context.Context, docID string, options map[string]interface{}) (json.RawMessage, error)
	// CreateDoc creates a new doc, with a server-generated ID.
	CreateDoc(ctx context.Context, doc interface{}) (docID, rev string, err error)
	// Put writes the document in the database.
	Put(ctx context.Context, docID string, doc interface{}) (rev string, err error)
	// Delete marks the specified document as deleted.
	Delete(ctx context.Context, docID, rev string) (newRev string, err error)
	// Stats returns database statistics.
	Stats(ctx context.Context) (*DBStats, error)
	// Compact initiates compaction of the database.
	Compact(ctx context.Context) error
	// CompactView initiates compaction of the view.
	CompactView(ctx context.Context, ddocID string) error
	// ViewCleanup cleans up stale view files.
	ViewCleanup(ctx context.Context) error
	// Security returns the database's security document.
	Security(ctx context.Context) (*Security, error)
	// SetSecurity sets the database's security document.
	SetSecurity(ctx context.Context, security *Security) error
	// Changes returns a Rows iterator for the changes feed. In continuous mode,
	// the iterator will continue indefinitely, until Close is called.
	Changes(ctx context.Context, options map[string]interface{}) (Changes, error)
	// PutAttachment uploads an attachment to the specified document, returning
	// the new revision.
	PutAttachment(ctx context.Context, docID, rev, filename, contentType string, body io.Reader) (newRev string, err error)
	// GetAttachment fetches an attachment for the associated document ID. rev
	// may be an empty string to fetch the most recent document version.
	GetAttachment(ctx context.Context, docID, rev, filename string) (contentType string, md5sum MD5sum, body io.ReadCloser, err error)
	// DeleteAttachment deletes an attachment from a document, returning the
	// document's new revision.
	DeleteAttachment(ctx context.Context, docID, rev, filename string) (newRev string, err error)
	// Query performs a query against a view, subject to the options provided.
	// ddoc will be the design doc name without the '_design/' previx.
	// view will be the view name without the '_view/' prefix.
	Query(ctx context.Context, ddoc, view string, options map[string]interface{}) (Rows, error)
}

// OldBulkDocer is deprecated and will be removed in Kivik 2.0. Use BulkDocer instead.
type OldBulkDocer interface {
	// BulkDocs alls bulk create, update and/or delete operations. It returns an
	// iterator over the results.
	BulkDocs(ctx context.Context, docs []interface{}) (BulkResults, error)
}
