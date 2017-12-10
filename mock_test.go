package kivik

import (
	"context"
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
	id string
	driver.DB
	ChangesFunc func(context.Context, map[string]interface{}) (driver.Changes, error)
}

var _ driver.DB = &mockDB{}

func (db *mockDB) Changes(ctx context.Context, opts map[string]interface{}) (driver.Changes, error) {
	return db.ChangesFunc(ctx, opts)
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
