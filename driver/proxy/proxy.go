package proxy

import (
	"context"
	"io"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
)

// CompleteClient is a composite of all compulsory and optional driver.* client
// interfaces.
type CompleteClient interface {
	driver.Client
	driver.Authenticator
	driver.UUIDer
	driver.Cluster
	driver.Configer
}

// NewClient wraps an existing *kivik.Client connection, allowing it to be used
// as a driver.Client
func NewClient(c *kivik.Client) driver.Client {
	return &client{c}
}

type client struct {
	*kivik.Client
}

var _ CompleteClient = &client{}

func (c *client) AllDBsContext(ctx context.Context, options map[string]interface{}) ([]string, error) {
	return c.Client.AllDBsContext(ctx, options)
}

func (c *client) CreateDBContext(ctx context.Context, dbname string, options map[string]interface{}) error {
	return c.Client.CreateDBContext(ctx, dbname, options)
}

func (c *client) DBExistsContext(ctx context.Context, dbname string, options map[string]interface{}) (bool, error) {
	return c.Client.DBExistsContext(ctx, dbname, options)
}

func (c *client) DestroyDBContext(ctx context.Context, dbname string, options map[string]interface{}) error {
	return c.Client.DestroyDBContext(ctx, dbname, options)
}

func (c *client) ServerInfoContext(ctx context.Context, options map[string]interface{}) (driver.ServerInfo, error) {
	return c.Client.ServerInfoContext(ctx, options)
}

func (c *client) DBContext(ctx context.Context, name string, options map[string]interface{}) (driver.DB, error) {
	d, err := c.Client.DBContext(ctx, name, options)
	return &db{d}, err
}

func (c *client) ConfigContext(ctx context.Context) (driver.Config, error) {
	return c.ConfigContext(ctx)
}

type db struct {
	*kivik.DB
}

var _ driver.DB = &db{}

func (d *db) AllDocsContext(ctx context.Context, opts map[string]interface{}) (driver.Rows, error) {
	kivikRows, err := d.DB.AllDocsContext(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &rows{kivikRows}, nil
}

func (d *db) QueryContext(ctx context.Context, ddoc, view string, opts map[string]interface{}) (driver.Rows, error) {
	kivikRows, err := d.DB.QueryContext(ctx, ddoc, view, opts)
	if err != nil {
		return nil, err
	}
	return &rows{kivikRows}, nil
}

func (d *db) GetContext(ctx context.Context, id string, i interface{}, opts map[string]interface{}) error {
	return d.DB.GetContext(ctx, id, i, opts)
}

func (d *db) InfoContext(ctx context.Context) (*driver.DBInfo, error) {
	i, err := d.DB.InfoContext(ctx)
	dbinfo := driver.DBInfo(*i)
	return &dbinfo, err
}

func (d *db) SecurityContext(ctx context.Context) (*driver.Security, error) {
	s, err := d.DB.SecurityContext(ctx)
	if err != nil {
		return nil, err
	}
	sec := driver.Security{
		Admins:  driver.Members(s.Admins),
		Members: driver.Members(s.Members),
	}
	return &sec, err
}

func (d *db) SetSecurityContext(ctx context.Context, security *driver.Security) error {
	sec := &kivik.Security{
		Admins:  kivik.Members(security.Admins),
		Members: kivik.Members(security.Members),
	}
	return d.DB.SetSecurityContext(ctx, sec)
}

func (d *db) ChangesContext(ctx context.Context, opts map[string]interface{}) (driver.Rows, error) {
	kivikRows, err := d.DB.ChangesContext(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &rows{kivikRows}, nil
}

func (d *db) BulkDocsContext(_ context.Context, _ ...interface{}) (driver.BulkResults, error) {
	// FIXME: Unimplemented
	return nil, nil
}

func (d *db) PutAttachmentContext(_ context.Context, _, _, _, _ string, _ io.Reader) (string, error) {
	// FIXME: Unimplemented
	return "", nil
}

func (d *db) GetAttachmentContext(ctx context.Context, docID, rev, filename string) (contentType string, md5sum driver.Checksum, body io.ReadCloser, err error) {
	// FIXME: Unimplemented
	return "", [16]byte{}, nil, nil
}
