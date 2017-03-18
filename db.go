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
	driverDB driver.DB
}

// SetOption sets a database-specific option. Available options are driver
// specific.
func (db *DB) SetOption(key string, value interface{}) error {
	return db.driverDB.SetOption(key, value)
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

// Security calls SecurityContext with a background context.
func (db *DB) Security() (*Security, error) {
	return db.SecurityContext(context.Background())
}

// SecurityContext returns the database's security document.
// See http://couchdb.readthedocs.io/en/latest/api/database/security.html#get--db-_security
func (db *DB) SecurityContext(ctx context.Context) (*Security, error) {
	s, err := db.driverDB.SecurityContext(ctx)
	if err != nil {
		return nil, err
	}
	return &Security{
		Admins:  Members(s.Admins),
		Members: Members(s.Members),
	}, err
}

// SetSecurity calls SetSecurityContext with a background context.
func (db *DB) SetSecurity(security *Security) error {
	return db.SetSecurityContext(context.Background(), security)
}

// SetSecurityContext sets the database's security document.
// See http://couchdb.readthedocs.io/en/latest/api/database/security.html#put--db-_security
func (db *DB) SetSecurityContext(ctx context.Context, security *Security) error {
	sec := &driver.Security{
		Admins:  driver.Members(security.Admins),
		Members: driver.Members(security.Members),
	}
	return db.driverDB.SetSecurityContext(ctx, sec)
}

// Rev calls RevContext with a background context.
func (db *DB) Rev(docID string) (rev string, err error) {
	return db.RevContext(context.Background(), docID)
}

// RevContext returns the most current rev of the requested document. This can
// be more efficient than a full document fetch, becuase only the rev is
// fetched from the server.
func (db *DB) RevContext(ctx context.Context, docID string) (rev string, err error) {
	if r, ok := db.driverDB.(driver.Rever); ok {
		return r.RevContext(ctx, docID)
	}
	var doc struct {
		Rev string `json:"_rev"`
	}
	// These last two lines cannot be combined for GopherJS due to a bug.
	// See https://github.com/gopherjs/gopherjs/issues/608
	err = db.GetContext(ctx, docID, &doc, nil)
	return doc.Rev, err
}

// RevsLimit calls RevsLimitContext with a background context.
func (db *DB) RevsLimit() (limit int, err error) {
	return db.RevsLimitContext(context.Background())
}

// RevsLimitContext returns the maximum number of document revisions that will
// be tracked.
// See http://couchdb.readthedocs.io/en/latest/api/database/misc.html#get--db-_revs_limit
func (db *DB) RevsLimitContext(ctx context.Context) (limit int, err error) {
	return db.driverDB.RevsLimitContext(ctx)
}

// SetRevsLimit calls SetRevsLimitContext with a background context.
func (db *DB) SetRevsLimit(limit int) error {
	return db.SetRevsLimitContext(context.Background(), limit)
}

// SetRevsLimitContext sets the maximum number of document revisions that will
// be tracked.
// See http://couchdb.readthedocs.io/en/latest/api/database/misc.html#put--db-_revs_limit
func (db *DB) SetRevsLimitContext(ctx context.Context, limit int) error {
	return db.driverDB.SetRevsLimitContext(ctx, limit)
}
