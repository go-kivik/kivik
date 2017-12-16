package kivik

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"strings"

	"github.com/go-kivik/kivik/driver"
)

type errReader string

var _ io.ReadCloser = errReader("")

func (r errReader) Close() error {
	return nil
}

func (r errReader) Read(_ []byte) (int, error) {
	return 0, errors.New(string(r))
}

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
