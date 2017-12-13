package kivik

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/flimzy/kivik/driver"
)

type mockDriver struct {
	NewClientFunc func(context.Context, string) (driver.Client, error)
}

var _ driver.Driver = &mockDriver{}

func (d *mockDriver) NewClient(ctx context.Context, dsn string) (driver.Client, error) {
	return d.NewClientFunc(ctx, dsn)
}

type mockDB struct {
	id                   string
	AllDocsFunc          func(context.Context, map[string]interface{}) (driver.Rows, error)
	GetFunc              func(context.Context, string, map[string]interface{}) (json.RawMessage, error)
	CreateDocFunc        func(context.Context, interface{}) (string, string, error)
	PutFunc              func(context.Context, string, interface{}) (string, error)
	DeleteFunc           func(context.Context, string, string) (string, error)
	StatsFunc            func(context.Context) (*driver.DBStats, error)
	CompactFunc          func(context.Context) error
	CompactViewFunc      func(context.Context, string) error
	ViewCleanupFunc      func(context.Context) error
	SecurityFunc         func(context.Context) (*driver.Security, error)
	SetSecurityFunc      func(context.Context, *driver.Security) error
	ChangesFunc          func(context.Context, map[string]interface{}) (driver.Changes, error)
	PutAttachmentFunc    func(context.Context, string, string, string, string, io.Reader) (string, error)
	GetAttachmentFunc    func(context.Context, string, string, string) (string, driver.MD5sum, io.ReadCloser, error)
	DeleteAttachmentFunc func(context.Context, string, string, string) (string, error)
	QueryFunc            func(context.Context, string, string, map[string]interface{}) (driver.Rows, error)
}

var _ driver.DB = &mockDB{}

func (db *mockDB) AllDocs(ctx context.Context, opts map[string]interface{}) (driver.Rows, error) {
	return db.AllDocsFunc(ctx, opts)
}

func (db *mockDB) Get(ctx context.Context, docID string, opts map[string]interface{}) (json.RawMessage, error) {
	return db.GetFunc(ctx, docID, opts)
}

func (db *mockDB) CreateDoc(ctx context.Context, doc interface{}) (string, string, error) {
	return db.CreateDocFunc(ctx, doc)
}

func (db *mockDB) Put(ctx context.Context, docID string, doc interface{}) (string, error) {
	return db.PutFunc(ctx, docID, doc)
}

func (db *mockDB) Delete(ctx context.Context, docID, rev string) (string, error) {
	return db.DeleteFunc(ctx, docID, rev)
}

func (db *mockDB) Stats(ctx context.Context) (*driver.DBStats, error) {
	return db.StatsFunc(ctx)
}

func (db *mockDB) Compact(ctx context.Context) error {
	return db.CompactFunc(ctx)
}

func (db *mockDB) CompactView(ctx context.Context, docID string) error {
	return db.CompactViewFunc(ctx, docID)
}

func (db *mockDB) ViewCleanup(ctx context.Context) error {
	return db.ViewCleanupFunc(ctx)
}

func (db *mockDB) Security(ctx context.Context) (*driver.Security, error) {
	return db.SecurityFunc(ctx)
}

func (db *mockDB) SetSecurity(ctx context.Context, security *driver.Security) error {
	return db.SetSecurityFunc(ctx, security)
}

func (db *mockDB) Changes(ctx context.Context, opts map[string]interface{}) (driver.Changes, error) {
	return db.ChangesFunc(ctx, opts)
}

func (db *mockDB) PutAttachment(ctx context.Context, docID, rev, filename, cType string, body io.Reader) (string, error) {
	return db.PutAttachmentFunc(ctx, docID, rev, filename, cType, body)
}

func (db *mockDB) GetAttachment(ctx context.Context, docID, rev, filename string) (string, driver.MD5sum, io.ReadCloser, error) {
	return db.GetAttachmentFunc(ctx, docID, rev, filename)
}

func (db *mockDB) DeleteAttachment(ctx context.Context, docID, rev, filename string) (string, error) {
	return db.DeleteAttachmentFunc(ctx, docID, rev, filename)
}

func (db *mockDB) Query(ctx context.Context, ddoc, view string, opts map[string]interface{}) (driver.Rows, error) {
	return db.QueryFunc(ctx, ddoc, view, opts)
}

type mockDBOpts struct {
	*mockDB
	CreateDocOptsFunc        func(context.Context, interface{}, map[string]interface{}) (string, string, error)
	PutOptsFunc              func(context.Context, string, interface{}, map[string]interface{}) (string, error)
	DeleteOptsFunc           func(context.Context, string, string, map[string]interface{}) (string, error)
	PutAttachmentOptsFunc    func(context.Context, string, string, string, string, io.Reader, map[string]interface{}) (string, error)
	GetAttachmentOptsFunc    func(context.Context, string, string, string, map[string]interface{}) (string, driver.MD5sum, io.ReadCloser, error)
	DeleteAttachmentOptsFunc func(context.Context, string, string, string, map[string]interface{}) (string, error)
}

var _ driver.DBOpts = &mockDBOpts{}

func (db *mockDBOpts) CreateDocOpts(ctx context.Context, doc interface{}, opts map[string]interface{}) (string, string, error) {
	return db.CreateDocOptsFunc(ctx, doc, opts)
}

func (db *mockDBOpts) PutOpts(ctx context.Context, docID string, doc interface{}, options map[string]interface{}) (string, error) {
	return db.PutOptsFunc(ctx, docID, doc, options)
}

func (db *mockDBOpts) DeleteAttachmentOpts(ctx context.Context, docID, rev, filename string, options map[string]interface{}) (string, error) {
	return db.DeleteAttachmentOptsFunc(ctx, docID, rev, filename, options)
}

func (db *mockDBOpts) DeleteOpts(ctx context.Context, docID, rev string, options map[string]interface{}) (string, error) {
	return db.DeleteOptsFunc(ctx, docID, rev, options)
}

func (db *mockDBOpts) PutAttachmentOpts(ctx context.Context, docID, rev, filename, cType string, body io.Reader, options map[string]interface{}) (string, error) {
	return db.PutAttachmentOptsFunc(ctx, docID, rev, filename, cType, body, options)
}

func (db *mockDBOpts) GetAttachmentOpts(ctx context.Context, docID, rev, filename string, options map[string]interface{}) (string, driver.MD5sum, io.ReadCloser, error) {
	return db.GetAttachmentOptsFunc(ctx, docID, rev, filename, options)
}

type mockFinder struct {
	*mockDB
	CreateIndexFunc func(context.Context, string, string, interface{}) error
	DeleteIndexFunc func(context.Context, string, string) error
	FindFunc        func(context.Context, interface{}) (driver.Rows, error)
	GetIndexesFunc  func(context.Context) ([]driver.Index, error)
}

var _ driver.Finder = &mockFinder{}

func (db *mockFinder) CreateIndex(ctx context.Context, ddoc, name string, index interface{}) error {
	return db.CreateIndexFunc(ctx, ddoc, name, index)
}

func (db *mockFinder) DeleteIndex(ctx context.Context, ddoc, name string) error {
	return db.DeleteIndexFunc(ctx, ddoc, name)
}

func (db *mockFinder) Find(ctx context.Context, query interface{}) (driver.Rows, error) {
	return db.FindFunc(ctx, query)
}

func (db *mockFinder) GetIndexes(ctx context.Context) ([]driver.Index, error) {
	return db.GetIndexesFunc(ctx)
}

type mockExplainer struct {
	*mockDB
	ExplainFunc func(context.Context, interface{}) (*driver.QueryPlan, error)
}

var _ driver.Explainer = &mockExplainer{}

func (db *mockExplainer) Explain(ctx context.Context, query interface{}) (*driver.QueryPlan, error) {
	return db.ExplainFunc(ctx, query)
}

type mockDBFlusher struct {
	*mockDB
	FlushFunc func(context.Context) error
}

var _ driver.DBFlusher = &mockDBFlusher{}

func (db *mockDBFlusher) Flush(ctx context.Context) error {
	return db.FlushFunc(ctx)
}

type mockRever struct {
	*mockDB
	RevFunc func(context.Context, string) (string, error)
}

var _ driver.Rever = &mockRever{}

func (db *mockRever) Rev(ctx context.Context, docID string) (string, error) {
	return db.RevFunc(ctx, docID)
}

type mockCopier struct {
	*mockDB
	CopyFunc func(context.Context, string, string, map[string]interface{}) (string, error)
}

var _ driver.Copier = &mockCopier{}

func (db *mockCopier) Copy(ctx context.Context, target, source string, options map[string]interface{}) (string, error) {
	return db.CopyFunc(ctx, target, source, options)
}

type errReader string

var _ io.ReadCloser = errReader("")

func (r errReader) Close() error {
	return nil
}

func (r errReader) Read(_ []byte) (int, error) {
	return 0, errors.New(string(r))
}

type mockBulkResults struct {
	result *driver.BulkResult
	err    error
}

var _ driver.BulkResults = &mockBulkResults{}

func (r *mockBulkResults) Next(i *driver.BulkResult) error {
	if r.result != nil {
		*i = *r.result
	}
	return r.err
}

func (r *mockBulkResults) Close() error { return nil }

func body(s string) io.ReadCloser {
	return ioutil.NopCloser(strings.NewReader(s))
}

type mockIterator struct {
	NextFunc  func(interface{}) error
	CloseFunc func() error
}

var _ iterator = &mockIterator{}

func (i *mockIterator) Next(ifce interface{}) error {
	return i.NextFunc(ifce)
}

func (i *mockIterator) Close() error {
	return i.CloseFunc()
}

type mockChanges struct {
	NextFunc  func(*driver.Change) error
	CloseFunc func() error
}

var _ driver.Changes = &mockChanges{}

func (c *mockChanges) Next(ch *driver.Change) error {
	return c.NextFunc(ch)
}

func (c *mockChanges) Close() error {
	return c.CloseFunc()
}

type mockRows struct {
	id            string
	CloseFunc     func() error
	NextFunc      func(*driver.Row) error
	OffsetFunc    func() int64
	TotalRowsFunc func() int64
	UpdateSeqFunc func() string
}

var _ driver.Rows = &mockRows{}

type mockRowsWarner struct {
	*mockRows
	WarningFunc func() string
}

var _ driver.RowsWarner = &mockRowsWarner{}

type mockBookmarker struct {
	*mockRows
	BookmarkFunc func() string
}

var _ driver.Bookmarker = &mockBookmarker{}

func (r *mockRows) Close() error {
	return r.CloseFunc()
}

func (r *mockRows) Next(row *driver.Row) error {
	return r.NextFunc(row)
}

func (r *mockRows) Offset() int64 {
	return r.OffsetFunc()
}

func (r *mockRows) TotalRows() int64 {
	return r.TotalRowsFunc()
}

func (r *mockRows) UpdateSeq() string {
	return r.UpdateSeqFunc()
}

func (r *mockRowsWarner) Warning() string {
	return r.WarningFunc()
}

func (r *mockBookmarker) Bookmark() string {
	return r.BookmarkFunc()
}

type mockReplication struct {
	DeleteFunc        func(context.Context) error
	StartTimeFunc     func() time.Time
	EndTimeFunc       func() time.Time
	ErrFunc           func() error
	ReplicationIDFunc func() string
	SourceFunc        func() string
	TargetFunc        func() string
	StateFunc         func() string
	UpdateFunc        func(context.Context, *driver.ReplicationInfo) error
}

var _ driver.Replication = &mockReplication{}

func (r *mockReplication) Delete(ctx context.Context) error {
	return r.DeleteFunc(ctx)
}

func (r *mockReplication) StartTime() time.Time {
	return r.StartTimeFunc()
}

func (r *mockReplication) EndTime() time.Time {
	return r.EndTimeFunc()
}

func (r *mockReplication) Err() error {
	return r.ErrFunc()
}

func (r *mockReplication) ReplicationID() string {
	return r.ReplicationIDFunc()
}

func (r *mockReplication) Source() string {
	return r.SourceFunc()
}

func (r *mockReplication) Target() string {
	return r.TargetFunc()
}

func (r *mockReplication) State() string {
	return r.StateFunc()
}

func (r *mockReplication) Update(ctx context.Context, rep *driver.ReplicationInfo) error {
	return r.UpdateFunc(ctx, rep)
}

type mockClient struct {
	id            string
	AllDBsFunc    func(context.Context, map[string]interface{}) ([]string, error)
	CreateDBFunc  func(context.Context, string, map[string]interface{}) error
	DBFunc        func(context.Context, string, map[string]interface{}) (driver.DB, error)
	DBExistsFunc  func(context.Context, string, map[string]interface{}) (bool, error)
	DestroyDBFunc func(context.Context, string, map[string]interface{}) error
	VersionFunc   func(context.Context) (*driver.Version, error)
}

var _ driver.Client = &mockClient{}

func (c *mockClient) AllDBs(ctx context.Context, opts map[string]interface{}) ([]string, error) {
	return c.AllDBsFunc(ctx, opts)
}

func (c *mockClient) CreateDB(ctx context.Context, dbname string, opts map[string]interface{}) error {
	return c.CreateDBFunc(ctx, dbname, opts)
}

func (c *mockClient) DB(ctx context.Context, dbname string, opts map[string]interface{}) (driver.DB, error) {
	return c.DBFunc(ctx, dbname, opts)
}

func (c *mockClient) DBExists(ctx context.Context, dbname string, opts map[string]interface{}) (bool, error) {
	return c.DBExistsFunc(ctx, dbname, opts)
}

func (c *mockClient) DestroyDB(ctx context.Context, dbname string, opts map[string]interface{}) error {
	return c.DestroyDBFunc(ctx, dbname, opts)
}

func (c *mockClient) Version(ctx context.Context) (*driver.Version, error) {
	return c.VersionFunc(ctx)
}

type mockClientReplicator struct {
	*mockClient
	GetReplicationsFunc func(context.Context, map[string]interface{}) ([]driver.Replication, error)
	ReplicateFunc       func(context.Context, string, string, map[string]interface{}) (driver.Replication, error)
}

var _ driver.ClientReplicator = &mockClientReplicator{}

func (c *mockClientReplicator) GetReplications(ctx context.Context, opts map[string]interface{}) ([]driver.Replication, error) {
	return c.GetReplicationsFunc(ctx, opts)
}

func (c *mockClientReplicator) Replicate(ctx context.Context, target, source string, opts map[string]interface{}) (driver.Replication, error) {
	return c.ReplicateFunc(ctx, target, source, opts)
}

type mockAuthenticator struct {
	*mockClient
	AuthenticateFunc func(context.Context, interface{}) error
}

var _ driver.Authenticator = &mockAuthenticator{}

func (c *mockAuthenticator) Authenticate(ctx context.Context, a interface{}) error {
	return c.AuthenticateFunc(ctx, a)
}

type mockDBUpdater struct {
	*mockClient
	DBUpdatesFunc func() (driver.DBUpdates, error)
}

var _ driver.DBUpdater = &mockDBUpdater{}

func (c *mockDBUpdater) DBUpdates() (driver.DBUpdates, error) {
	return c.DBUpdatesFunc()
}

type mockDBUpdates struct {
	id        string
	NextFunc  func(*driver.DBUpdate) error
	CloseFunc func() error
}

var _ driver.DBUpdates = &mockDBUpdates{}

func (u *mockDBUpdates) Close() error {
	return u.CloseFunc()
}

func (u *mockDBUpdates) Next(dbupdate *driver.DBUpdate) error {
	return u.NextFunc(dbupdate)
}
