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
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
)

// DB is a handle to a specific database.
type DB struct {
	client   *Client
	name     string
	driverDB driver.DB
	err      error

	closed bool
	mu     sync.Mutex
	wg     sync.WaitGroup
}

func (db *DB) startQuery() (end func(), _ error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.closed {
		return nil, ErrDatabaseClosed
	}
	endQuery, err := db.client.startQuery()
	if err != nil {
		return nil, err
	}
	var once sync.Once
	db.wg.Add(1)
	return func() {
		once.Do(func() {
			db.mu.Lock()
			db.wg.Done()
			endQuery()
			db.mu.Unlock()
		})
	}, nil
}

// Client returns the client used to connect to the database.
func (db *DB) Client() *Client {
	return db.client
}

// Name returns the database name as passed when creating the DB connection.
func (db *DB) Name() string {
	return db.name
}

// Err returns the error, if any, that occurred while connecting to or creating
// the database. This error will be deferred until the next call, normally, so
// using this method is only necessary if you need to directly check the error,
// and intend to do nothing else with the DB object.
func (db *DB) Err() error {
	return db.err
}

// AllDocs returns a list of all documents in the database.
func (db *DB) AllDocs(ctx context.Context, options ...Option) *ResultSet {
	if db.err != nil {
		return &ResultSet{iter: errIterator(db.err)}
	}
	endQuery, err := db.startQuery()
	if err != nil {
		return &ResultSet{iter: errIterator(err)}
	}
	rowsi, err := db.driverDB.AllDocs(ctx, multiOptions(options))
	if err != nil {
		endQuery()
		return &ResultSet{iter: errIterator(err)}
	}
	return newResultSet(ctx, endQuery, rowsi)
}

// DesignDocs returns a list of all documents in the database.
func (db *DB) DesignDocs(ctx context.Context, options ...Option) *ResultSet {
	if db.err != nil {
		return &ResultSet{iter: errIterator(db.err)}
	}
	ddocer, ok := db.driverDB.(driver.DesignDocer)
	if !ok {
		return &ResultSet{iter: errIterator(&internal.Error{Status: http.StatusNotImplemented, Err: errors.New("kivik: design doc view not supported by driver")})}
	}

	endQuery, err := db.startQuery()
	if err != nil {
		return &ResultSet{iter: errIterator(err)}
	}
	rowsi, err := ddocer.DesignDocs(ctx, multiOptions(options))
	if err != nil {
		endQuery()
		return &ResultSet{iter: errIterator(err)}
	}
	return newResultSet(ctx, endQuery, rowsi)
}

// LocalDocs returns a list of all documents in the database.
func (db *DB) LocalDocs(ctx context.Context, options ...Option) *ResultSet {
	if db.err != nil {
		return &ResultSet{iter: errIterator(db.err)}
	}
	ldocer, ok := db.driverDB.(driver.LocalDocer)
	if !ok {
		return &ResultSet{iter: errIterator(&internal.Error{Status: http.StatusNotImplemented, Err: errors.New("kivik: local doc view not supported by driver")})}
	}
	endQuery, err := db.startQuery()
	if err != nil {
		return &ResultSet{iter: errIterator(err)}
	}
	rowsi, err := ldocer.LocalDocs(ctx, multiOptions(options))
	if err != nil {
		endQuery()
		return &ResultSet{iter: errIterator(err)}
	}
	return newResultSet(ctx, endQuery, rowsi)
}

// Query executes the specified view function from the specified design
// document. ddoc and view may or may not be be prefixed with '_design/'
// and '_view/' respectively.
//
// See [views] in the CouchDB documentation.
//
// If supported by the backend and database (i.e. CouchDB 2.2+), you may pass
// multiple queries to a single view by passing an option called `queries` with
// a multi-query object as a value.
//
// See [multiple queries] in the CouchDB documentation.
//
// [views]: https://docs.couchdb.org/en/stable/api/ddoc/views.html#
// [multiple queries]: https://docs.couchdb.org/en/stable/api/ddoc/views.html#sending-multiple-queries-to-a-view
func (db *DB) Query(ctx context.Context, ddoc, view string, options ...Option) *ResultSet {
	if db.err != nil {
		return &ResultSet{iter: errIterator(db.err)}
	}
	endQuery, err := db.startQuery()
	if err != nil {
		return &ResultSet{iter: errIterator(err)}
	}
	ddoc = strings.TrimPrefix(ddoc, "_design/")
	view = strings.TrimPrefix(view, "_view/")
	rowsi, err := db.driverDB.Query(ctx, ddoc, view, multiOptions(options))
	if err != nil {
		endQuery()
		return &ResultSet{iter: errIterator(err)}
	}
	return newResultSet(ctx, endQuery, rowsi)
}

// Document is a single document result returned by [DB.Get].
type Document struct {
	err         error
	rev         string
	body        io.Reader
	attachments driver.Attachments

	mu sync.Mutex
}

// Err returns the error, if any, that was encountered fetching the document.
func (r *Document) Err() error {
	return r.err
}

// Rev returns the document revision.
func (r *Document) Rev() (string, error) {
	return r.rev, r.err
}

// ScanDoc unmarshals the document into i.
func (r *Document) ScanDoc(i interface{}) error {
	if r.err != nil {
		return r.err
	}
	return json.NewDecoder(r.body).Decode(i)
}

// Attachments returns an attachments iterator if the document includes
// attachments and they are not inline.
func (r *Document) Attachments() (*AttachmentsIterator, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.attachments == nil {
		return nil, errNoAttachments
	}
	r.mu.Lock()
	return &AttachmentsIterator{
		atti:    r.attachments,
		onClose: r.mu.Unlock,
	}, nil
}

// Close closes the document resources.
func (r *Document) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if atts := r.attachments; atts != nil {
		_ = atts.Close()
	}
	if c, ok := r.body.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// Get fetches the requested document. Any errors are deferred until the first
// call to [ResultSet.ScanDoc] or any other result set method.
func (db *DB) Get(ctx context.Context, docID string, options ...Option) *Document {
	if db.err != nil {
		return &Document{err: db.err}
	}
	endQuery, err := db.startQuery()
	if err != nil {
		return &Document{err: err}
	}
	defer endQuery()
	result, err := db.driverDB.Get(ctx, docID, multiOptions(options))
	if err != nil {
		return &Document{err: err}
	}
	return &Document{
		rev:         result.Rev,
		body:        result.Body,
		attachments: result.Attachments,
	}
}

// OpenRevs returns documents of specified leaf revisions. Additionally, it
// accepts a revs value of "all" to return all leaf revisions.
//
// This function is experimental, and may change without notice.
func (db *DB) OpenRevs(ctx context.Context, docID string, revs []string, options ...Option) *ResultSet {
	if db.err != nil {
		return &ResultSet{iter: errIterator(db.err)}
	}
	if openRever, ok := db.driverDB.(driver.OpenRever); ok {
		endQuery, err := db.startQuery()
		if err != nil {
			return &ResultSet{iter: errIterator(err)}
		}
		rowsi, err := openRever.OpenRevs(ctx, docID, revs, multiOptions(options))
		if err != nil {
			endQuery()
			return &ResultSet{iter: errIterator(err)}
		}
		return newResultSet(ctx, endQuery, rowsi)
	}
	return &ResultSet{iter: errIterator(errOpenRevsNotImplemented)}
}

// GetRev returns the active rev of the specified document. GetRev accepts
// the same options as [DB.Get].
func (db *DB) GetRev(ctx context.Context, docID string, options ...Option) (rev string, err error) {
	if db.err != nil {
		return "", db.err
	}
	opts := multiOptions(options)
	if r, ok := db.driverDB.(driver.RevGetter); ok {
		endQuery, err := db.startQuery()
		if err != nil {
			return "", err
		}
		defer endQuery()
		return r.GetRev(ctx, docID, opts)
	}
	row := db.Get(ctx, docID, opts)
	var doc struct {
		Rev string `json:"_rev"`
	}
	// These last two lines cannot be combined for GopherJS due to a bug.
	// See https://github.com/gopherjs/gopherjs/issues/608
	err = row.ScanDoc(&doc)
	return doc.Rev, err
}

// CreateDoc creates a new doc with an auto-generated unique ID. The generated
// docID and new rev are returned.
func (db *DB) CreateDoc(ctx context.Context, doc interface{}, options ...Option) (docID, rev string, err error) {
	if db.err != nil {
		return "", "", db.err
	}
	if docCreator, ok := db.driverDB.(driver.DocCreator); ok {
		endQuery, err := db.startQuery()
		if err != nil {
			return "", "", err
		}
		defer endQuery()
		return docCreator.CreateDoc(ctx, doc, multiOptions(options))
	}
	docID, ok := extractDocID(doc)
	if !ok {
		// TODO: Consider making uuid algorithm configurable
		docID = uuid.NewString()
	}
	rev, err = db.Put(ctx, docID, doc, options...)
	return docID, rev, err
}

// normalizeFromJSON unmarshals a []byte, json.RawMessage or io.Reader to a
// map[string]interface{}, or passed through any other types.
func normalizeFromJSON(i interface{}) (interface{}, error) {
	switch t := i.(type) {
	case json.Marshaler:
		return t, nil
	case io.Reader:
		body, err := io.ReadAll(t)
		if err != nil {
			return nil, &internal.Error{Status: http.StatusBadRequest, Err: err}
		}
		return json.RawMessage(body), nil
	default:
		return i, nil
	}
}

func extractDocID(i interface{}) (string, bool) {
	if i == nil {
		return "", false
	}
	var id string
	var ok bool
	switch t := i.(type) {
	case map[string]interface{}:
		id, ok = t["_id"].(string)
	case map[string]string:
		id, ok = t["_id"]
	default:
		data, err := json.Marshal(i)
		if err != nil {
			return "", false
		}
		var result struct {
			ID string `json:"_id"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return "", false
		}
		id = result.ID
		ok = result.ID != ""
	}
	if !ok {
		return "", false
	}
	return id, true
}

// Put creates a new doc or updates an existing one, with the specified docID.
// If the document already exists, the current revision must be included in doc,
// with JSON key '_rev', otherwise a conflict will occur. The new rev is
// returned.
//
// doc may be one of:
//
//   - A value to be marshaled to JSON. The resulting JSON structure must
//     conform to CouchDB standards.
//   - An [encoding/json.RawMessage] value containing a valid JSON document
//   - An [io.Reader], from which a valid JSON document may be read.
func (db *DB) Put(ctx context.Context, docID string, doc interface{}, options ...Option) (rev string, err error) {
	if db.err != nil {
		return "", db.err
	}
	if docID == "" {
		return "", missingArg("docID")
	}
	endQuery, err := db.startQuery()
	if err != nil {
		return "", err
	}
	defer endQuery()
	i, err := normalizeFromJSON(doc)
	if err != nil {
		return "", err
	}
	return db.driverDB.Put(ctx, docID, i, multiOptions(options))
}

// Update executes an [update function] for the specified document. If the docID
// is the empty string, the update function is executed for a null document. See
// the CouchDB documentation for [standard updates] and [update of null documents].
//
// Not all backends support update functions. If the backend does not support
// update functions, this method will return an error.
//
// [update function]: https://docs.couchdb.org/en/stable/ddocs/ddocs.html#updatefun
// [standard updates]: https://docs.couchdb.org/en/stable/api/ddoc/render.html#put--db-_design-ddoc-_update-func-docid
// [update of null documents]: https://docs.couchdb.org/en/stable/api/ddoc/render.html#post--db-_design-ddoc-_update-func
func (db *DB) Update(ctx context.Context, ddoc, funcName, docID string, doc interface{}, options ...Option) (newRev string, err error) {
	if db.err != nil {
		return "", db.err
	}
	updateDB, ok := db.driverDB.(driver.Updater)
	if !ok {
		return "", errUpdateNotImplemented
	}

	if ddoc == "" {
		return "", missingArg("ddoc")
	}
	if funcName == "" {
		return "", missingArg("funcName")
	}
	if docID == "" {
		return "", missingArg("docID")
	}
	endQuery, err := db.startQuery()
	if err != nil {
		return "", err
	}
	defer endQuery()
	i, err := normalizeFromJSON(doc)
	if err != nil {
		return "", err
	}
	return updateDB.Update(ctx, ddoc, funcName, docID, i, multiOptions(options))
}

// Delete marks the specified document as deleted. The revision may be provided
// via options, which takes priority over the rev argument.
func (db *DB) Delete(ctx context.Context, docID, rev string, options ...Option) (newRev string, err error) {
	if db.err != nil {
		return "", db.err
	}
	endQuery, err := db.startQuery()
	if err != nil {
		return "", err
	}
	defer endQuery()
	if docID == "" {
		return "", missingArg("docID")
	}
	opts := append(multiOptions{Rev(rev)}, options...)
	return db.driverDB.Delete(ctx, docID, opts)
}

// Flush requests a flush of disk cache to disk or other permanent storage.
//
// See the [CouchDB documentation].
//
// [CouchDB documentation]: http://docs.couchdb.org/en/2.0.0/api/database/compact.html#db-ensure-full-commit
func (db *DB) Flush(ctx context.Context) error {
	if db.err != nil {
		return db.err
	}
	endQuery, err := db.startQuery()
	if err != nil {
		return err
	}
	defer endQuery()
	if flusher, ok := db.driverDB.(driver.Flusher); ok {
		return flusher.Flush(ctx)
	}
	return &internal.Error{Status: http.StatusNotImplemented, Err: errors.New("kivik: flush not supported by driver")}
}

// DBStats contains database statistics..
type DBStats struct {
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
	// If this number is lower than [DBStats.DiskSize], then compaction would
	// free disk space.
	ActiveSize int64 `json:"data_size"`
	// ExternalSize is the size of the documents in the database, as represented
	// as JSON, before compression.
	ExternalSize int64 `json:"-"`
	// Cluster reports the cluster replication configuration variables.
	Cluster *ClusterConfig `json:"cluster,omitempty"`
	// RawResponse is the raw response body returned by the server, useful if
	// you need additional backend-specific information.
	//
	// For the format of this document, see the [CouchDB documentation].
	//
	// [CouchDB documentation]: http://docs.couchdb.org/en/2.1.1/api/database/common.html#get--db
	RawResponse json.RawMessage `json:"-"`
}

// ClusterConfig contains the cluster configuration for the database.
type ClusterConfig struct {
	Replicas    int `json:"n"`
	Shards      int `json:"q"`
	ReadQuorum  int `json:"r"`
	WriteQuorum int `json:"w"`
}

// Stats returns database statistics. See the [CouchDB documentation].
//
// [CouchDB documentation]: https://docs.couchdb.org/en/stable/api/database/common.html#get--db
func (db *DB) Stats(ctx context.Context) (*DBStats, error) {
	if db.err != nil {
		return nil, db.err
	}
	endQuery, err := db.startQuery()
	if err != nil {
		return nil, err
	}
	defer endQuery()
	i, err := db.driverDB.Stats(ctx)
	if err != nil {
		return nil, err
	}
	return driverStats2kivikStats(i), nil
}

func driverStats2kivikStats(i *driver.DBStats) *DBStats {
	var cluster *ClusterConfig
	if i.Cluster != nil {
		c := ClusterConfig(*i.Cluster)
		cluster = &c
	}
	return &DBStats{
		Name:           i.Name,
		CompactRunning: i.CompactRunning,
		DocCount:       i.DocCount,
		DeletedCount:   i.DeletedCount,
		UpdateSeq:      i.UpdateSeq,
		DiskSize:       i.DiskSize,
		ActiveSize:     i.ActiveSize,
		ExternalSize:   i.ExternalSize,
		Cluster:        cluster,
		RawResponse:    i.RawResponse,
	}
}

// Compact begins compaction of the database. Check the CompactRunning field
// returned by [DB.Stats] to see if the compaction has completed.
//
// See the [CouchDB documentation].
//
// This method may return immediately, or may wait for the compaction to
// complete before returning, depending on the backend implementation. In
// particular, CouchDB triggers the compaction and returns immediately, whereas
// PouchDB waits until compaction has completed, before returning.
//
// [CouchDB documentation]: http://docs.couchdb.org/en/2.0.0/api/database/compact.html#db-compact
func (db *DB) Compact(ctx context.Context) error {
	if db.err != nil {
		return db.err
	}
	endQuery, err := db.startQuery()
	if err != nil {
		return err
	}
	defer endQuery()
	return db.driverDB.Compact(ctx)
}

// CompactView compats the view indexes associated with the specified design
// document.
//
// See the [CouchDB documentation].
//
// This method may return immediately, or may wait for the compaction to
// complete before returning, depending on the backend implementation. In
// particular, CouchDB triggers the compaction and returns immediately, whereas
// PouchDB waits until compaction has completed, before returning.
//
// [CouchDB documentation]: http://docs.couchdb.org/en/2.0.0/api/database/compact.html#db-compact-design-doc
func (db *DB) CompactView(ctx context.Context, ddocID string) error {
	if db.err != nil {
		return db.err
	}
	endQuery, err := db.startQuery()
	if err != nil {
		return err
	}
	defer endQuery()
	return db.driverDB.CompactView(ctx, ddocID)
}

// ViewCleanup removes view index files that are no longer required as a result
// of changed views within design documents.
//
// See the [CouchDB documentation].
//
// [CouchDB documentation]: http://docs.couchdb.org/en/2.0.0/api/database/compact.html#db-view-cleanup
func (db *DB) ViewCleanup(ctx context.Context) error {
	if db.err != nil {
		return db.err
	}
	endQuery, err := db.startQuery()
	if err != nil {
		return err
	}
	defer endQuery()
	return db.driverDB.ViewCleanup(ctx)
}

// Security returns the database's security document.
//
// See the [CouchDB documentation].
//
// [CouchDB documentation]: http://couchdb.readthedocs.io/en/latest/api/database/security.html#get--db-_security
func (db *DB) Security(ctx context.Context) (*Security, error) {
	if db.err != nil {
		return nil, db.err
	}
	secDB, ok := db.driverDB.(driver.SecurityDB)
	if !ok {
		return nil, errSecurityNotImplemented
	}
	endQuery, err := db.startQuery()
	if err != nil {
		return nil, err
	}
	defer endQuery()
	s, err := secDB.Security(ctx)
	if err != nil {
		return nil, err
	}
	return &Security{
		Admins:          Members(s.Admins),
		Members:         Members(s.Members),
		Cloudant:        s.Cloudant,
		CouchdbAuthOnly: s.CouchdbAuthOnly,
	}, err
}

// SetSecurity sets the database's security document.
//
// See the [CouchDB documentation].
//
// [CouchDB documentation]: http://couchdb.readthedocs.io/en/latest/api/database/security.html#put--db-_security
func (db *DB) SetSecurity(ctx context.Context, security *Security) error {
	if db.err != nil {
		return db.err
	}
	secDB, ok := db.driverDB.(driver.SecurityDB)
	if !ok {
		return errSecurityNotImplemented
	}
	if security == nil {
		return missingArg("security")
	}
	endQuery, err := db.startQuery()
	if err != nil {
		return err
	}
	defer endQuery()
	sec := &driver.Security{
		Admins:          driver.Members(security.Admins),
		Members:         driver.Members(security.Members),
		Cloudant:        security.Cloudant,
		CouchdbAuthOnly: security.CouchdbAuthOnly,
	}
	return secDB.SetSecurity(ctx, sec)
}

// Copy copies the source document to a new document with an ID of targetID. If
// the database backend does not support COPY directly, the operation will be
// emulated with a Get followed by Put. The target will be an exact copy of the
// source, with only the ID and revision changed.
//
// See the [CouchDB documentation]:
//
// [CouchDB documentation]: http://docs.couchdb.org/en/2.0.0/api/document/common.html#copy--db-docid
func (db *DB) Copy(ctx context.Context, targetID, sourceID string, options ...Option) (targetRev string, err error) {
	if db.err != nil {
		return "", db.err
	}
	if targetID == "" {
		return "", missingArg("targetID")
	}
	if sourceID == "" {
		return "", missingArg("sourceID")
	}
	opts := multiOptions(options)
	if copier, ok := db.driverDB.(driver.Copier); ok {
		endQuery, err := db.startQuery()
		if err != nil {
			return "", err
		}
		defer endQuery()
		return copier.Copy(ctx, targetID, sourceID, opts)
	}
	var doc map[string]interface{}
	if err = db.Get(ctx, sourceID, options...).ScanDoc(&doc); err != nil {
		return "", err
	}
	delete(doc, "_rev")
	doc["_id"] = targetID
	opts2 := map[string]interface{}{}
	opts.Apply(opts2)
	delete(opts2, "rev") // rev has a completely different meaning for Copy and Put
	return db.Put(ctx, targetID, doc, Params(opts2))
}

// PutAttachment uploads the supplied content as an attachment to the specified
// document.
func (db *DB) PutAttachment(ctx context.Context, docID string, att *Attachment, options ...Option) (newRev string, err error) {
	if db.err != nil {
		return "", db.err
	}
	if docID == "" {
		return "", missingArg("docID")
	}
	if e := att.validate(); e != nil {
		return "", e
	}
	endQuery, err := db.startQuery()
	if err != nil {
		return "", err
	}
	defer endQuery()
	a := driver.Attachment(*att)
	return db.driverDB.PutAttachment(ctx, docID, &a, multiOptions(options))
}

// GetAttachment returns a file attachment associated with the document.
func (db *DB) GetAttachment(ctx context.Context, docID, filename string, options ...Option) (*Attachment, error) {
	if db.err != nil {
		return nil, db.err
	}
	endQuery, err := db.startQuery()
	if err != nil {
		return nil, err
	}
	defer endQuery()
	if docID == "" {
		return nil, missingArg("docID")
	}
	if filename == "" {
		return nil, missingArg("filename")
	}
	att, err := db.driverDB.GetAttachment(ctx, docID, filename, multiOptions(options))
	if err != nil {
		return nil, err
	}
	a := Attachment(*att)
	return &a, nil
}

type nilContentReader struct{}

var _ io.ReadCloser = &nilContentReader{}

func (c nilContentReader) Read(_ []byte) (int, error) { return 0, io.EOF }
func (c nilContentReader) Close() error               { return nil }

var nilContent = nilContentReader{}

// GetAttachmentMeta returns meta data about an attachment. The attachment
// content returned will be empty.
func (db *DB) GetAttachmentMeta(ctx context.Context, docID, filename string, options ...Option) (*Attachment, error) {
	if db.err != nil {
		return nil, db.err
	}
	if docID == "" {
		return nil, missingArg("docID")
	}
	if filename == "" {
		return nil, missingArg("filename")
	}
	var att *Attachment
	if metaer, ok := db.driverDB.(driver.AttachmentMetaGetter); ok {
		endQuery, err := db.startQuery()
		if err != nil {
			return nil, err
		}
		defer endQuery()
		a, err := metaer.GetAttachmentMeta(ctx, docID, filename, multiOptions(options))
		if err != nil {
			return nil, err
		}
		att = new(Attachment)
		*att = Attachment(*a)
	} else {
		var err error
		att, err = db.GetAttachment(ctx, docID, filename, options...)
		if err != nil {
			return nil, err
		}
	}
	if att.Content != nil {
		_ = att.Content.Close() // Ensure this is closed
	}
	att.Content = nilContent
	return att, nil
}

// DeleteAttachment deletes an attachment from a document, returning the
// document's new revision. The revision may be provided via options, which
// takes priority over the rev argument.
func (db *DB) DeleteAttachment(ctx context.Context, docID, rev, filename string, options ...Option) (newRev string, err error) {
	if db.err != nil {
		return "", db.err
	}
	endQuery, err := db.startQuery()
	if err != nil {
		return "", err
	}
	defer endQuery()
	if docID == "" {
		return "", missingArg("docID")
	}
	if filename == "" {
		return "", missingArg("filename")
	}
	opts := append(multiOptions{Rev(rev)}, options...)
	return db.driverDB.DeleteAttachment(ctx, docID, filename, opts)
}

// PurgeResult is the result of a purge request.
type PurgeResult struct {
	// Seq is the purge sequence number.
	Seq int64 `json:"purge_seq"`
	// Purged is a map of document ids to revisions, indicated the
	// document/revision pairs that were successfully purged.
	Purged map[string][]string `json:"purged"`
}

// Purge permanently removes the reference to deleted documents from the
// database. Normal deletion only marks the document with the key/value pair
// `_deleted=true`, to ensure proper replication of deleted documents. By
// using Purge, the document can be completely removed. But note that this
// operation is not replication safe, so great care must be taken when using
// Purge, and this should only be used as a last resort.
//
// Purge expects as input a map with document ID as key, and slice of
// revisions as value.
func (db *DB) Purge(ctx context.Context, docRevMap map[string][]string) (*PurgeResult, error) {
	if db.err != nil {
		return nil, db.err
	}
	endQuery, err := db.startQuery()
	if err != nil {
		return nil, err
	}
	defer endQuery()
	if purger, ok := db.driverDB.(driver.Purger); ok {
		res, err := purger.Purge(ctx, docRevMap)
		if err != nil {
			return nil, err
		}
		r := PurgeResult(*res)
		return &r, nil
	}
	return nil, &internal.Error{Status: http.StatusNotImplemented, Message: "kivik: purge not supported by driver"}
}

// BulkGetReference is a reference to a document given to pass to [DB.BulkGet].
type BulkGetReference struct {
	ID        string `json:"id"`
	Rev       string `json:"rev,omitempty"`
	AttsSince string `json:"atts_since,omitempty"`
}

// BulkGet can be called to query several documents in bulk. It is well suited
// for fetching a specific revision of documents, as replicators do for example,
// or for getting revision history.
//
// See the [CouchDB documentation].
//
// [CouchDB documentation]: http://docs.couchdb.org/en/stable/api/database/bulk-api.html#db-bulk-get
func (db *DB) BulkGet(ctx context.Context, docs []BulkGetReference, options ...Option) *ResultSet {
	if db.err != nil {
		return &ResultSet{iter: errIterator(db.err)}
	}
	bulkGetter, ok := db.driverDB.(driver.BulkGetter)
	if !ok {
		return &ResultSet{iter: errIterator(&internal.Error{Status: http.StatusNotImplemented, Message: "kivik: bulk get not supported by driver"})}
	}

	endQuery, err := db.startQuery()
	if err != nil {
		return &ResultSet{iter: errIterator(err)}
	}
	refs := make([]driver.BulkGetReference, len(docs))
	for i, ref := range docs {
		refs[i] = driver.BulkGetReference(ref)
	}
	rowsi, err := bulkGetter.BulkGet(ctx, refs, multiOptions(options))
	if err != nil {
		endQuery()
		return &ResultSet{iter: errIterator(err)}
	}
	return newResultSet(ctx, endQuery, rowsi)
}

// Close cleans up any resources used by the DB. Close is safe to call
// concurrently with other DB operations and will block until all other DB
// operations finish. After calling Close, any other DB operations will return
// [ErrDatabaseClosed].
func (db *DB) Close() error {
	if db.err != nil {
		return db.err
	}
	db.mu.Lock()
	db.closed = true
	db.mu.Unlock()
	db.wg.Wait()
	return db.driverDB.Close()
}

// RevDiff represents a rev diff for a single document, as returned by the
// [DB.RevsDiff] method.
type RevDiff struct {
	Missing           []string `json:"missing,omitempty"`
	PossibleAncestors []string `json:"possible_ancestors,omitempty"`
}

// Diffs is a collection of [RevDiff] values as returned by [DB.RevsDiff]. The
// map key is the document ID.
type Diffs map[string]RevDiff

// RevsDiff returns the subset of document/revision IDs that do not correspond
// to revisions stored in the database. This is used by the replication
// protocol, and is normally never needed otherwise.  revMap must marshal to the
// [expected format].
//
// Use [ResultSet.ID] to return the current document ID, and
// [ResultSet.ScanValue] to access the full JSON value. The [RevsDiff] type
// matches this format and is provided as a convenience for unmarshaling.
//
//	{
//	    "missing": ["rev1",...],
//	    "possible_ancestors": ["revA",...]
//	}
//
// [expected format]: http://docs.couchdb.org/en/stable/api/database/misc.html#db-revs-diff
func (db *DB) RevsDiff(ctx context.Context, revMap interface{}) *ResultSet {
	if db.err != nil {
		return &ResultSet{iter: errIterator(db.err)}
	}
	if rd, ok := db.driverDB.(driver.RevsDiffer); ok {
		endQuery, err := db.startQuery()
		if err != nil {
			return &ResultSet{iter: errIterator(err)}
		}
		rowsi, err := rd.RevsDiff(ctx, revMap)
		if err != nil {
			endQuery()
			return &ResultSet{iter: errIterator(err)}
		}
		return newResultSet(ctx, endQuery, rowsi)
	}
	return &ResultSet{iter: errIterator(&internal.Error{Status: http.StatusNotImplemented, Message: "kivik: _revs_diff not supported by driver"})}
}

// PartitionStats contains partition statistics.
type PartitionStats struct {
	DBName          string
	DocCount        int64
	DeletedDocCount int64
	Partition       string
	ActiveSize      int64
	ExternalSize    int64
	RawResponse     json.RawMessage
}

// PartitionStats returns statistics about the named partition.
//
// See the [CouchDB documentation].
//
// [CouchDB documentation]: https://docs.couchdb.org/en/stable/api/partitioned-dbs.html#db-partition-partition
func (db *DB) PartitionStats(ctx context.Context, name string) (*PartitionStats, error) {
	if db.err != nil {
		return nil, db.err
	}
	endQuery, err := db.startQuery()
	if err != nil {
		return nil, err
	}
	defer endQuery()
	if pdb, ok := db.driverDB.(driver.PartitionedDB); ok {
		stats, err := pdb.PartitionStats(ctx, name)
		if err != nil {
			return nil, err
		}
		s := PartitionStats(*stats)
		return &s, nil
	}
	return nil, &internal.Error{Status: http.StatusNotImplemented, Message: "kivik: partitions not supported by driver"}
}
