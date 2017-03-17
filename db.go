package kivik

import (
	"context"
	"strings"
	"time"

	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/errors"
)

// DB is a handle to a specific database.
type DB struct {
	// AutoFlush automatically requests a flush after each database write. This
	// creates additional network traffic, and may hurt server performance. This
	// option has no effect if the `delayed_commits` is false on the server.
	AutoFlush bool

	driverDB driver.DB
}

func (db *DB) autoFlush() {
	if db.AutoFlush {
		_, _ = db.Flush()
	}
}

// AllDocs calls AllDocsContext with a background context.
func (db *DB) AllDocs(docs interface{}, options Options) (offset, totalrows int, seq string, err error) {
	return db.AllDocsContext(context.Background(), docs, options)
}

// AllDocsContext returns a list of all documents in the database.
func (db *DB) AllDocsContext(ctx context.Context, docs interface{}, options Options) (offset, totalrows int, seq string, err error) {
	return db.driverDB.AllDocsContext(ctx, docs, options)
}

// Get calls GetContext with a background context.
func (db *DB) Get(docID string, doc interface{}, options Options) error {
	return db.GetContext(context.Background(), docID, doc, options)
}

// GetContext fetches the requested document.
func (db *DB) GetContext(ctx context.Context, docID string, doc interface{}, options Options) error {
	return db.driverDB.GetContext(ctx, docID, doc, options)
}

// CreateDoc calls CreateDocContext with a background context.
func (db *DB) CreateDoc(doc interface{}) (docID, rev string, err error) {
	return db.CreateDocContext(context.Background(), doc)
}

// CreateDocContext creates a new doc with an auto-generated unique ID. The generated
// docID and new rev are returned.
func (db *DB) CreateDocContext(ctx context.Context, doc interface{}) (docID, rev string, err error) {
	defer db.autoFlush()
	return db.driverDB.CreateDocContext(ctx, doc)
}

// Put calls PutContext with a background context.
func (db *DB) Put(docID string, doc interface{}) (rev string, err error) {
	return db.PutContext(context.Background(), docID, doc)
}

// PutContext creates a new doc or updates an existing one, with the specified
// docID. If the document already exists, the current revision must be included
// in doc, with JSON key '_rev', otherwise a conflict will occur. The new rev is
// returned.
func (db *DB) PutContext(ctx context.Context, docID string, doc interface{}) (rev string, err error) {
	defer db.autoFlush()
	// The '/' char is only permitted in the case of '_design/', so check that here
	if designDoc := strings.TrimPrefix(docID, "_design/"); strings.Contains(designDoc, "/") {
		return "", errors.Status(StatusBadRequest, "invalid document ID")
	}
	return db.driverDB.PutContext(ctx, docID, doc)
}

// Delete calls DeleteContext with a background context.
func (db *DB) Delete(docID, rev string) (newRev string, err error) {
	return db.DeleteContext(context.Background(), docID, rev)
}

// DeleteContext marks the specified document as deleted.
func (db *DB) DeleteContext(ctx context.Context, docID, rev string) (newRev string, err error) {
	defer db.autoFlush()
	return db.driverDB.DeleteContext(ctx, docID, rev)
}

// Flush calls FlushContext with a background context.
func (db *DB) Flush() (time.Time, error) {
	return db.FlushContext(context.Background())
}

// FlushContext requests a flush of disk cache to disk or other permanent storage.
// The response a timestamp when the database backend opened the storage
// backend.
//
// See http://docs.couchdb.org/en/2.0.0/api/database/compact.html#db-ensure-full-commit
func (db *DB) FlushContext(ctx context.Context) (time.Time, error) {
	if flusher, ok := db.driverDB.(driver.DBFlusher); ok {
		return flusher.FlushContext(ctx)
	}
	return time.Time{}, ErrNotImplemented
}

// DBInfo is a struct of information about a database instance. Not all fields
// are supported by all database drivers.
type DBInfo struct {
	// Name is the name of the database.
	Name string `json:"db_name"`
	// CompactRunning is true if the database is currently being compacted.
	CompactRunning bool `json:"compact_running"`
	// DocCount is the number of documents are currently stored in the database.
	DocCount int64 `json:"doc_count"`
	// DeletedCount is a count of documents which have been deleted from the
	// database.
	DeletedCount int64 `json:"doc_del_count"`
	// UpdateSeq is the current update sequence for the database.
	UpdateSeq string `json:"update_seq"`
	// DiskSize is the number of bytes used on-disk to store the database.
	DiskSize int64 `json:"disk_size"`
	// ActiveSize is the number of bytes used on-disk to store active documents.
	// If this number is lower than DiskSize, then compaction would free disk
	// space.
	ActiveSize int64 `json:"data_size"`
	// ExternalSize is the size of the documents in the database, as represented
	// as JSON, before compression.
	ExternalSize int64 `json:"-"`
}

// Info calls InfoContext with a background context.
func (db *DB) Info() (*DBInfo, error) {
	return db.InfoContext(context.Background())
}

// InfoContext returns basic statistics about the database.
func (db *DB) InfoContext(ctx context.Context) (*DBInfo, error) {
	i, err := db.driverDB.InfoContext(ctx)
	dbinfo := DBInfo(*i)
	return &dbinfo, err
}

// Compact calls CompactContext with a background context.
func (db *DB) Compact() error {
	return db.CompactContext(context.Background())
}

// CompactContext begins compaction of the database. Check the CompactRunning
// field returned by Info() to see if the compaction has completed.
// See http://docs.couchdb.org/en/2.0.0/api/database/compact.html#db-compact
func (db *DB) CompactContext(ctx context.Context) error {
	return db.driverDB.CompactContext(ctx)
}

// CompactView calls CompactViewContext with a background context.
func (db *DB) CompactView(ddocID string) error {
	return db.CompactViewContext(context.Background(), ddocID)
}

// CompactViewContext compats the view indexes associated with the specified
// design document.
// See http://docs.couchdb.org/en/2.0.0/api/database/compact.html#db-compact-design-doc
func (db *DB) CompactViewContext(ctx context.Context, ddocID string) error {
	return db.driverDB.CompactViewContext(ctx, ddocID)
}

// ViewCleanup calls ViewCleanupContext with a background context.
func (db *DB) ViewCleanup() error {
	return db.ViewCleanupContext(context.Background())
}

// ViewCleanupContext removes view index files that are no longer required as a
// result of changed views within design documents.
// See http://docs.couchdb.org/en/2.0.0/api/database/compact.html#db-view-cleanup
func (db *DB) ViewCleanupContext(ctx context.Context) error {
	return db.driverDB.ViewCleanupContext(ctx)
}
