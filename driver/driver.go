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

package driver

import (
	"context"
	"encoding/json"
	"io"
	"time"
)

// Options represents a collection of arbitrary client or query options.
//
// An implementation should also implement the [fmt.Stringer] interface for the
// sake of display when used by [github.com/go-kivik/kivik/v4/mockdb].
type Options interface {
	// Apply applies the option to target, if target is of the expected type.
	// Unexpected/recognized target types should be ignored.
	Apply(target interface{})
}

// Driver is the interface that must be implemented by a database driver.
type Driver interface {
	// NewClient returns a connection handle to the database. The name is in a
	// driver-specific format.
	NewClient(name string, options Options) (Client, error)
}

// Version represents a server version response.
type Version struct {
	// Version is the version number reported by the server or backend.
	Version string
	// Vendor is the vendor string reported by the server or backend.
	Vendor string
	// Features is a list of enabled, optional features.  This was added in
	// CouchDB 2.1.0, and can be expected to be empty for older versions.
	Features []string
	// RawResponse is the raw response body as returned by the server.
	RawResponse json.RawMessage
}

// Client is a connection to a database server.
type Client interface {
	// Version returns the server implementation's details.
	Version(ctx context.Context) (*Version, error)
	// AllDBs returns a list of all existing database names.
	AllDBs(ctx context.Context, options Options) ([]string, error)
	// DBExists returns true if the database exists.
	DBExists(ctx context.Context, dbName string, options Options) (bool, error)
	// CreateDB creates the requested database.
	CreateDB(ctx context.Context, dbName string, options Options) error
	// DestroyDB deletes the requested database.
	DestroyDB(ctx context.Context, dbName string, options Options) error
	// DB returns a handle to the requested database.
	DB(dbName string, options Options) (DB, error)
}

// DBsStatser is an optional interface that a [DB] may implement, added to
// support CouchDB 2.2.0's /_dbs_info endpoint. If this is not supported, or
// if this method returns status 404, Kivik will fall back to calling the method
// of issuing a GET /{db} for each database requested.
type DBsStatser interface {
	// DBsStats returns database statistical information for each database
	// named in dbNames. The returned values should be in the same order as
	// the requested database names, and any missing databases should return
	// a nil *DBStats value.
	DBsStats(ctx context.Context, dbNames []string) ([]*DBStats, error)
}

// AllDBsStatser is an optional interface that a [DB] may implement, added to
// support CouchDB 3.2's GET /_dbs_info endpoint. If this is not supported, or
// if this method returns status 404 or 405, Kivik will fall back to using
// _all_dbs + the [DBStatser] interface (or its respective emulation).
type AllDBsStatser interface {
	// AllDBsStats returns database statistical information for each database
	// in the CouchDB instance. See the [CouchDB documentation] for supported
	// options.
	//
	// [CouchDB documentation]: https://docs.couchdb.org/en/stable/api/server/common.html#get--_dbs_info
	AllDBsStats(ctx context.Context, options Options) ([]*DBStats, error)
}

// Replication represents a _replicator document.
type Replication interface {
	// The following methods are called just once, when the Replication is first
	// returned from [ClientReplicator.Replicate] or
	// [ClientReplicator.GetReplications].
	ReplicationID() string
	Source() string
	Target() string

	// The following methods return values may be updated by calls to [Update].
	StartTime() time.Time
	EndTime() time.Time
	State() string
	Err() error

	// These methods may be triggered by user actions.

	// Delete deletes a replication, which cancels it if it is running.
	Delete(context.Context) error
	// Update fetches the latest replication state from the server.
	Update(context.Context, *ReplicationInfo) error
}

// ReplicationInfo represents a snapshot state of a replication, as provided
// by the _active_tasks endpoint.
type ReplicationInfo struct {
	DocWriteFailures int64
	DocsRead         int64
	DocsWritten      int64
	Progress         float64
}

// ClientReplicator is an optional interface that may be implemented by a [Client]
// that supports replication between two database.
type ClientReplicator interface {
	// Replicate initiates a replication.
	Replicate(ctx context.Context, targetDSN, sourceDSN string, options Options) (Replication, error)
	// GetReplications returns a list of replications (i.e. all docs in the
	// _replicator database)
	GetReplications(ctx context.Context, options Options) ([]Replication, error)
}

// DBStats contains database statistics.
type DBStats struct {
	Name           string          `json:"db_name"`
	CompactRunning bool            `json:"compact_running"`
	DocCount       int64           `json:"doc_count"`
	DeletedCount   int64           `json:"doc_del_count"`
	UpdateSeq      string          `json:"update_seq"`
	DiskSize       int64           `json:"disk_size"`
	ActiveSize     int64           `json:"data_size"`
	ExternalSize   int64           `json:"-"`
	Cluster        *ClusterStats   `json:"cluster,omitempty"`
	RawResponse    json.RawMessage `json:"-"`
}

// ClusterStats contains the cluster configuration for the database.
type ClusterStats struct {
	Replicas    int `json:"n"`
	Shards      int `json:"q"`
	ReadQuorum  int `json:"r"`
	WriteQuorum int `json:"w"`
}

// Members represents the members of a database security document.
type Members struct {
	Names []string `json:"names,omitempty"`
	Roles []string `json:"roles,omitempty"`
}

// Security represents a database security document.
type Security struct {
	Admins          Members             `json:"admins,omitempty"`
	Members         Members             `json:"members,omitempty"`
	Cloudant        map[string][]string `json:"cloudant,omitempty"`
	CouchdbAuthOnly *bool               `json:"couchdb_auth_only,omitempty"`
}

// MarshalJSON satisfies the json.Marshaler interface.
func (s Security) MarshalJSON() ([]byte, error) {
	var v struct {
		Admins          *Members            `json:"admins,omitempty"`
		Members         *Members            `json:"members,omitempty"`
		Cloudant        map[string][]string `json:"cloudant,omitempty"`
		CouchdbAuthOnly *bool               `json:"couchdb_auth_only,omitempty"`
	}
	if len(s.Admins.Names) > 0 || len(s.Admins.Roles) > 0 {
		v.Admins = &s.Admins
	}
	if len(s.Members.Names) > 0 || len(s.Members.Roles) > 0 {
		v.Members = &s.Members
	}
	if len(s.Cloudant) > 0 {
		v.Cloudant = s.Cloudant
	}
	if s.CouchdbAuthOnly != nil {
		v.CouchdbAuthOnly = s.CouchdbAuthOnly
	}
	return json.Marshal(v)
}

// DB is a database handle.
type DB interface {
	// AllDocs returns all of the documents in the database, subject to the
	// options provided.
	AllDocs(ctx context.Context, options Options) (Rows, error)
	// Put writes the document in the database.
	Put(ctx context.Context, docID string, doc interface{}, options Options) (rev string, err error)
	// Get fetches the requested document from the database.
	Get(ctx context.Context, docID string, options Options) (*Document, error)
	// Delete marks the specified document as deleted.
	Delete(ctx context.Context, docID string, options Options) (newRev string, err error)
	// Stats returns database statistics.
	Stats(ctx context.Context) (*DBStats, error)
	// Compact initiates compaction of the database.
	Compact(ctx context.Context) error
	// CompactView initiates compaction of the view.
	CompactView(ctx context.Context, ddocID string) error
	// ViewCleanup cleans up stale view files.
	ViewCleanup(ctx context.Context) error
	// Changes returns an iterator for the changes feed. In continuous mode,
	// the iterator will continue indefinitely, until [Changes.Close] is called.
	Changes(ctx context.Context, options Options) (Changes, error)
	// PutAttachment uploads an attachment to the specified document, returning
	// the new revision.
	PutAttachment(ctx context.Context, docID string, att *Attachment, options Options) (newRev string, err error)
	// GetAttachment fetches an attachment for the associated document ID.
	GetAttachment(ctx context.Context, docID, filename string, options Options) (*Attachment, error)
	// DeleteAttachment deletes an attachment from a document, returning the
	// document's new revision.
	DeleteAttachment(ctx context.Context, docID, filename string, options Options) (newRev string, err error)
	// Query performs a query against a view, subject to the options provided.
	// ddoc will be the design doc name without the '_design/' previx.
	// view will be the view name without the '_view/' prefix.
	Query(ctx context.Context, ddoc, view string, options Options) (Rows, error)
	// Close is called to clean up any resources used by the database.
	Close() error
}

// DocCreator is an optional interface that extends a [DB] to support the
// creation of new documents. If not implemented, [DB.Put] will be used to
// emulate the functionality, with missing document IDs generated as V4 UUIDs.
type DocCreator interface {
	// CreateDoc creates a new doc, with a server-generated ID.
	CreateDoc(ctx context.Context, doc interface{}, options Options) (docID, rev string, err error)
}

// OpenRever is an optional interface that extends a [DB] to support the open_revs
// option of the CouchDB get document endpoint. It is used by the replicator.
// Drivers that don't support this endpoint may not be able to replicate as
// efficiently, or at all.
type OpenRever interface {
	// OpenRevs fetches the requested document revisions from the database.
	// revs may be a list of revisions, or a single item with value "all" to
	// request all open revs.
	OpenRevs(ctx context.Context, docID string, revs []string, options Options) (Rows, error)
}

// SecurityDB is an optional interface that extends a [DB], for backends which
// support security documents.
type SecurityDB interface {
	// Security returns the database's security document.
	Security(ctx context.Context) (*Security, error)
	// SetSecurity sets the database's security document.
	SetSecurity(ctx context.Context, security *Security) error
}

// Updater is an optional interface that extends a [DB] to support invoking
// update functions.
type Updater interface {
	// Update calls the named update function with the provided document.
	Update(ctx context.Context, ddoc, funcName, docID string, doc interface{}, options Options) (rev string, err error)
}

// Document represents a single document returned by [DB.Get].
type Document struct {
	// Rev is the revision number returned
	Rev string

	// Body returns the document JSON.
	Body io.ReadCloser

	// Attachments will be nil except when attachments=true.
	Attachments Attachments
}

// Attachments is an iterator over the attachments included in a document when
// [DB.Get] is called with `attachments=true`.
type Attachments interface {
	// Next is called to populate att with the next attachment in the result
	// set.
	//
	// Next should return [io.EOF] when there are no more attachments.
	Next(att *Attachment) error

	// Close closes the Attachments iterator.
	Close() error
}

// Purger is an optional interface which may be implemented by a [DB] to support
// document purging.
type Purger interface {
	// Purge permanently removes the references to deleted documents from the
	// database.
	Purge(ctx context.Context, docRevMap map[string][]string) (*PurgeResult, error)
}

// PurgeResult is the result of a purge request.
type PurgeResult struct {
	Seq    int64               `json:"purge_seq"`
	Purged map[string][]string `json:"purged"`
}

// BulkDocer is an optional interface which may be implemented by a [DB] to
// support bulk insert/update operations. For any driver that does not support
// the BulkDocer interface, the [DB.Put] or [DB.CreateDoc] methods will be
// called for each document to emulate the same functionality, with options
// passed through unaltered.
type BulkDocer interface {
	// BulkDocs alls bulk create, update and/or delete operations. It returns an
	// iterator over the results.
	BulkDocs(ctx context.Context, docs []interface{}, options Options) ([]BulkResult, error)
}

// Finder is an optional interface which may be implemented by a [DB]. It
// provides access to the MongoDB-style query interface added in CouchDB 2.
type Finder interface {
	// Find executes a query using the new /_find interface. query is always
	// converted to a [encoding/json.RawMessage] value before passing it to the
	// driver. The type remains `interface{}` for backward compatibility, but
	// will change with Kivik 5.x. See [issue #1015] for details.
	//
	// [issue #1014]: https://github.com/go-kivik/kivik/issues/1015
	Find(ctx context.Context, query interface{}, options Options) (Rows, error)
	// CreateIndex creates an index if it doesn't already exist. If the index
	// already exists, it should do nothing. ddoc and name may be empty, in
	// which case they should be provided by the backend. If index is a string,
	// []byte, or [encoding/json.RawMessage], it should be treated as a raw
	// JSON payload. Any other type should be marshaled to JSON.
	CreateIndex(ctx context.Context, ddoc, name string, index interface{}, options Options) error
	// GetIndexes returns a list of all indexes in the database.
	GetIndexes(ctx context.Context, options Options) ([]Index, error)
	// Delete deletes the requested index.
	DeleteIndex(ctx context.Context, ddoc, name string, options Options) error
	// Explain returns the query plan for a given query. Explain takes the same
	// arguments as [Finder.Find].
	Explain(ctx context.Context, query interface{}, options Options) (*QueryPlan, error)
}

// QueryPlan is the response of an Explain query.
type QueryPlan struct {
	DBName   string                 `json:"dbname"`
	Index    map[string]interface{} `json:"index"`
	Selector map[string]interface{} `json:"selector"`
	Options  map[string]interface{} `json:"opts"`
	Limit    int64                  `json:"limit"`
	Skip     int64                  `json:"skip"`

	// Fields is the list of fields to be returned in the result set, or
	// an empty list if all fields are to be returned.
	Fields []interface{}          `json:"fields"`
	Range  map[string]interface{} `json:"range"`
}

// Index is a MonboDB-style index definition.
type Index struct {
	DesignDoc  string      `json:"ddoc,omitempty"`
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	Definition interface{} `json:"def"`
}

// Attachment represents a file attachment to a document.
type Attachment struct {
	Filename        string        `json:"-"`
	ContentType     string        `json:"content_type"`
	Stub            bool          `json:"stub"`
	Follows         bool          `json:"follows"`
	Content         io.ReadCloser `json:"-"`
	Size            int64         `json:"length"`
	ContentEncoding string        `json:"encoding"`
	EncodedLength   int64         `json:"encoded_length"`
	RevPos          int64         `json:"revpos"`
	Digest          string        `json:"digest"`
}

// AttachmentMetaGetter is an optional interface which may be implemented by a
// [DB]. When not implemented, [DB.GetAttachment] will be used to emulate the
// functionality.
type AttachmentMetaGetter interface {
	// GetAttachmentMeta returns meta information about an attachment.
	GetAttachmentMeta(ctx context.Context, docID, filename string, options Options) (*Attachment, error)
}

// BulkResult is the result of a single doc update in a BulkDocs request.
type BulkResult struct {
	ID    string `json:"id"`
	Rev   string `json:"rev"`
	Error error
}

// RevGetter is an optional interface that may be implemented by a [DB]. If not
// implemented, [DB.Get] will be used to emulate the functionality, with options
// passed through unaltered.
type RevGetter interface {
	// GetRev returns the document revision of the requested document. GetRev
	// should accept the same options as [DB.Get].
	GetRev(ctx context.Context, docID string, options Options) (rev string, err error)
}

// Flusher is an optional interface that may be implemented by a [DB] that can
// force a flush of the database backend file(s) to disk or other permanent
// storage.
type Flusher interface {
	// Flush requests a flush of disk cache to disk or other permanent storage.
	//
	// See the [CouchDB documentation].
	//
	// [CouchDB documentation]: http://docs.couchdb.org/en/2.0.0/api/database/compact.html#db-ensure-full-commit
	Flush(ctx context.Context) error
}

// Copier is an optional interface that may be implemented by a [DB].
//
// If a [DB] does not implement Copier, the functionality will be emulated by
// calling [DB.Get] followed by [DB.Put], with options passed through unaltered,
// except that the 'rev' option will be removed for the [DB.Put] call.
type Copier interface {
	Copy(ctx context.Context, targetID, sourceID string, options Options) (targetRev string, err error)
}

// DesignDocer is an optional interface that may be implemented by a [DB].
type DesignDocer interface {
	// DesignDocs returns all of the design documents in the database, subject
	// to the options provided.
	DesignDocs(ctx context.Context, options Options) (Rows, error)
}

// LocalDocer is an optional interface that may be implemented by a [DB].
type LocalDocer interface {
	// LocalDocs returns all of the local documents in the database, subject to
	// the options provided.
	LocalDocs(ctx context.Context, options Options) (Rows, error)
}

// Pinger is an optional interface that may be implemented by a [Client]. When
// not implemented, Kivik will call [Client.Version] instead to emulate the
// functionality.
type Pinger interface {
	// Ping returns true if the database is online and available for requests.
	Ping(ctx context.Context) (bool, error)
}

// ClusterMembership contains the list of known nodes, and cluster nodes, as
// returned by the [_membership endpoint].
//
// [_membership endpoint]: https://docs.couchdb.org/en/latest/api/server/common.html#get--_membership
type ClusterMembership struct {
	AllNodes     []string `json:"all_nodes"`
	ClusterNodes []string `json:"cluster_nodes"`
}

// Cluster is an optional interface that may be implemented by a [Client] to
// support CouchDB cluster configuration operations.
type Cluster interface {
	// ClusterStatus returns the current cluster status.
	ClusterStatus(ctx context.Context, options Options) (string, error)
	// ClusterSetup performs the action specified by action.
	ClusterSetup(ctx context.Context, action interface{}) error
	// Membership returns a list of all known nodes, and all nodes configured as
	// part of the cluster.
	Membership(ctx context.Context) (*ClusterMembership, error)
}

// ClientCloser is an optional interface that may be implemented by a [Client]
// to clean up resources when a client is no longer needed.
type ClientCloser interface {
	Close() error
}

// RevDiff represents a rev diff for a single document, as returned by
// [RevsDiffer.RevsDiff].
type RevDiff struct {
	Missing           []string `json:"missing,omitempty"`
	PossibleAncestors []string `json:"possible_ancestors,omitempty"`
}

// RevsDiffer is an optional interface that may be implemented by a [DB].
type RevsDiffer interface {
	// RevsDiff returns a Rows iterator, which should populate the ID and Value
	// fields, and nothing else.
	RevsDiff(ctx context.Context, revMap interface{}) (Rows, error)
}
