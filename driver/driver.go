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

// Driver is the interface that must be implemented by a database driver.
type Driver interface {
	// NewClient returns a connection handle to the database. The name is in a
	// driver-specific format.
	NewClient(name string, options map[interface{}]interface{}) (Client, error)
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
	AllDBs(ctx context.Context, options map[interface{}]interface{}) ([]string, error)
	// DBExists returns true if the database exists.
	DBExists(ctx context.Context, dbName string, options map[interface{}]interface{}) (bool, error)
	// CreateDB creates the requested DB. The dbName is validated as a valid
	// CouchDB database name prior to calling this function, so the driver can
	// assume a valid name.
	CreateDB(ctx context.Context, dbName string, options map[interface{}]interface{}) error
	// DestroyDB deletes the requested DB.
	DestroyDB(ctx context.Context, dbName string, options map[interface{}]interface{}) error
	// DB returns a handle to the requested database
	DB(dbName string, options map[interface{}]interface{}) (DB, error)
}

// DBsStatser is an optional interface, added to support CouchDB 2.2.0's
// /_dbs_info endpoint. If this is not supported, or if this method returns
// status 404, Kivik will fall back to calling the method of issuing a
// GET /{db} for each database requested.
type DBsStatser interface {
	// DBsStats returns database statistical information for each database
	// named in dbNames. The returned values should be in the same order as
	// the requested database names, and any missing databases should return
	// a nil *DBStats value.
	DBsStats(ctx context.Context, dbNames []string) ([]*DBStats, error)
}

// Replication represents a _replicator document.
type Replication interface {
	// The following methods are called just once, when the Replication is first
	// returned from [Replicate] or [GetReplications].
	ReplicationID() string
	Source() string
	Target() string
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

// ReplicationInfo represents a snap-shot state of a replication, as provided
// by the _active_tasks endpoint.
type ReplicationInfo struct {
	DocWriteFailures int64
	DocsRead         int64
	DocsWritten      int64
	Progress         float64
}

// ClientReplicator is an optional interface that may be implemented by a Client
// that supports replication between two database.
type ClientReplicator interface {
	// Replicate initiates a replication.
	Replicate(ctx context.Context, targetDSN, sourceDSN string, options map[interface{}]interface{}) (Replication, error)
	// GetReplications returns a list of replicatoins (i.e. all docs in the
	// _replicator database)
	GetReplications(ctx context.Context, options map[interface{}]interface{}) ([]Replication, error)
}

// Authenticator is an optional interface that may be implemented by a [Client]
// that supports authenitcated connections.
type Authenticator interface {
	// Authenticate attempts to authenticate the client using an authenticator.
	// If the authenticator is not known to the client, an error should be
	// returned.
	Authenticate(ctx context.Context, authenticator interface{}) error
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
	Admins  Members `json:"admins"`
	Members Members `json:"members"`
}

// MarshalJSON satisfies the json.Marshaler interface.
func (s Security) MarshalJSON() ([]byte, error) {
	var v struct {
		Admins  *Members `json:"admins,omitempty"`
		Members *Members `json:"members,omitempty"`
	}
	if len(s.Admins.Names) > 0 || len(s.Admins.Roles) > 0 {
		v.Admins = &s.Admins
	}
	if len(s.Members.Names) > 0 || len(s.Members.Roles) > 0 {
		v.Members = &s.Members
	}
	return json.Marshal(v)
}

// RowsGetter is an optional interface that a DB may implement, to return a
// multi-row resultset from Get(). DB must implement either OldGetter or RowsGetter.
type RowsGetter interface {
	// Get fetches the requested document from the database, and returns the
	// content length (or -1 if unknown), and an io.ReadCloser to access the
	// raw JSON content.
	Get(ctx context.Context, docID string, options map[interface{}]interface{}) (Rows, error)
}

// OldGetter is an optional interface that a DB may implement, to return a single
// document from the Get() method. DB must implement either OldGetter or RowsGetter.
type OldGetter interface {
	// Get fetches the requested document from the database, and returns the
	// content length (or -1 if unknown), and an io.ReadCloser to access the
	// raw JSON content.
	Get(ctx context.Context, docID string, options map[interface{}]interface{}) (*Document, error)
}

// DB is a database handle.
type DB interface {
	// AllDocs returns all of the documents in the database, subject to the
	// options provided.
	AllDocs(ctx context.Context, options map[interface{}]interface{}) (Rows, error)
	// CreateDoc creates a new doc, with a server-generated ID.
	CreateDoc(ctx context.Context, doc interface{}, options map[interface{}]interface{}) (docID, rev string, err error)
	// Put writes the document in the database.
	Put(ctx context.Context, docID string, doc interface{}, options map[interface{}]interface{}) (rev string, err error)
	// Delete marks the specified document as deleted.
	Delete(ctx context.Context, docID string, options map[interface{}]interface{}) (newRev string, err error)
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
	Changes(ctx context.Context, options map[interface{}]interface{}) (Changes, error)
	// PutAttachment uploads an attachment to the specified document, returning
	// the new revision.
	PutAttachment(ctx context.Context, docID string, att *Attachment, options map[interface{}]interface{}) (newRev string, err error)
	// GetAttachment fetches an attachment for the associated document ID.
	GetAttachment(ctx context.Context, docID, filename string, options map[interface{}]interface{}) (*Attachment, error)
	// DeleteAttachment deletes an attachment from a document, returning the
	// document's new revision.
	DeleteAttachment(ctx context.Context, docID, filename string, options map[interface{}]interface{}) (newRev string, err error)
	// Query performs a query against a view, subject to the options provided.
	// ddoc will be the design doc name without the '_design/' previx.
	// view will be the view name without the '_view/' prefix.
	Query(ctx context.Context, ddoc, view string, options map[interface{}]interface{}) (Rows, error)
}

// SecurityDB is an optional interface that extends a DB, for backends which
// support security documents.
type SecurityDB interface {
	// Security returns the database's security document.
	Security(ctx context.Context) (*Security, error)
	// SetSecurity sets the database's security document.
	SetSecurity(ctx context.Context, security *Security) error
}

// Document represents a single document returned by Get
type Document struct {
	// Rev is the revision number returned
	Rev string

	// Body contains the response body, either in raw JSON or multipart/related
	// format.
	Body io.ReadCloser

	// Attachments will be nil except when attachments=true.
	Attachments Attachments
}

// Attachments is an iterator over the attachments included in a document when
// Get is called with `include_docs=true`.
type Attachments interface {
	// Next is called to pupulate att with the next attachment in the result
	// set.
	//
	// Next should return [io.EOF] when there are no more attachments.
	Next(att *Attachment) error

	// Close closes the Attachments iterator.
	Close() error
}

// Purger is an optional interface which may be implemented by a DB to support
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

// BulkDocer is an optional interface which may be implemented by a DB to
// support bulk insert/update operations. For any driver that does not support
// the BulkDocer interface, the [DB.Put] or [DB.CreateDoc] methods will be
// called for each document to emulate the same functionality, with options
// passed through unaltered.
type BulkDocer interface {
	// BulkDocs alls bulk create, update and/or delete operations. It returns an
	// iterator over the results.
	BulkDocs(ctx context.Context, docs []interface{}, options map[interface{}]interface{}) ([]BulkResult, error)
}

// Finder is an optional interface which may be implemented by a DB. It provides
// access to the new (in CouchDB 2.0) MongoDB-style query interface.
type Finder interface {
	// Find executes a query using the new /_find interface. If query is a
	// string, []byte, or [encoding/json.RawMessage], it should be treated as a
	// raw JSON payload. Any other type should be marshaled to JSON.
	Find(ctx context.Context, query interface{}, options map[interface{}]interface{}) (Rows, error)
	// CreateIndex creates an index if it doesn't already exist. If the index
	// already exists, it should do nothing. ddoc and name may be empty, in
	// which case they should be provided by the backend. If index is a string,
	// []byte, or json.RawMessage, it should be treated as a raw JSON payload.
	// Any other type should be marshaled to JSON.
	CreateIndex(ctx context.Context, ddoc, name string, index interface{}, options map[interface{}]interface{}) error
	// GetIndexes returns a list of all indexes in the database.
	GetIndexes(ctx context.Context, options map[interface{}]interface{}) ([]Index, error)
	// Delete deletes the requested index.
	DeleteIndex(ctx context.Context, ddoc, name string, options map[interface{}]interface{}) error
	// Explain returns the query plan for a given query. Explain takes the same
	// arguments as [Finder.Find].
	Explain(ctx context.Context, query interface{}, options map[interface{}]interface{}) (*QueryPlan, error)
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

// AttachmentMetaGetter is an optional interface which may be satisfied by a
// DB. If satisfied, it may be used to fetch meta data about an attachment. If
// not satisfied, GetAttachment will be used instead.
type AttachmentMetaGetter interface {
	// GetAttachmentMeta returns meta information about an attachment.
	GetAttachmentMeta(ctx context.Context, docID, filename string, options map[interface{}]interface{}) (*Attachment, error)
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
	GetRev(ctx context.Context, docID string, options map[interface{}]interface{}) (rev string, err error)
}

// Flusher is an optional interface that may be implemented by a [DB] that can
// force a flush of the database backend file(s) to disk or other permanent
// storage.
type Flusher interface {
	// Flush requests a flush of disk cache to disk or other permanent storage.
	//
	// See http://docs.couchdb.org/en/2.0.0/api/database/compact.html#db-ensure-full-commit
	Flush(ctx context.Context) error
}

// Copier is an optional interface that may be implemented by a [DB].
//
// If a DB does not implement Copier, the functionality will be emulated by
// calling [DB.Get] followed by [DB.Put], with options passed through unaltered,
// except that the 'rev' option will be removed for the [DB.Put] call.
type Copier interface {
	Copy(ctx context.Context, targetID, sourceID string, options map[interface{}]interface{}) (targetRev string, err error)
}

// DesignDocer is an optional interface that may be implemented by a [DB].
type DesignDocer interface {
	// DesignDocs returns all of the design documents in the database, subject
	// to the options provided.
	DesignDocs(ctx context.Context, options map[interface{}]interface{}) (Rows, error)
}

// LocalDocer is an optional interface that may be implemented by a [DB].
type LocalDocer interface {
	// LocalDocs returns all of the local documents in the database, subject to
	// the options provided.
	LocalDocs(ctx context.Context, options map[interface{}]interface{}) (Rows, error)
}

// Pinger is an optional interface that may be implemented by a [Client]. When
// not implemented, Kivik will call [Client.Version] instead, to determine if
// the database is usable.
type Pinger interface {
	// Ping returns true if the database is online and available for requests.
	Ping(ctx context.Context) (bool, error)
}

// ClusterMembership contains the list of known nodes, and cluster nodes, as
// returned by the /_membership endpoint.
// See https://docs.couchdb.org/en/latest/api/server/common.html#get--_membership
type ClusterMembership struct {
	AllNodes     []string `json:"all_nodes"`
	ClusterNodes []string `json:"cluster_nodes"`
}

// Cluster is an optional interface that may be implemented by a [Client] to
// support CouchDB cluster configuration operations.
type Cluster interface {
	// ClusterStatus returns the current cluster status.
	ClusterStatus(ctx context.Context, options map[interface{}]interface{}) (string, error)
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

// DBCloser is an optional interface that may be implemented by a [DB] to clean
// up resources when a database is no longer needed.
type DBCloser interface {
	Close() error
}

// RevDiff represents a rev diff for a single document, as returned by
// [RevsDiffer.RevsDiff].
type RevDiff struct {
	Missing           []string `json:"missing,omitempty"`
	PossibleAncestors []string `json:"possible_ancestors,omitempty"`
}

// RevsDiffer is an optional interface that may be implemented by a DB.
type RevsDiffer interface {
	// RevsDiff returns a Rows iterator, which should populate the ID and Value
	// fields, and nothing else.
	RevsDiff(ctx context.Context, revMap interface{}) (Rows, error)
}
