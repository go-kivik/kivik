package proxydb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

type statusError struct {
	error
	status int
}

func (e statusError) Unwrap() error   { return e.error }
func (e statusError) HTTPStatus() int { return e.status }

var notYetImplemented = statusError{status: http.StatusNotImplemented, error: errors.New("kivik: not yet implemented in proxy driver")}

// CompleteClient is a composite of all compulsory and optional driver.* client
// interfaces.
type CompleteClient interface {
	driver.Client
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

func (c *client) AllDBs(ctx context.Context, options driver.Options) ([]string, error) {
	return c.Client.AllDBs(ctx, options)
}

func (c *client) CreateDB(ctx context.Context, dbname string, options driver.Options) error {
	return c.Client.CreateDB(ctx, dbname, options)
}

func (c *client) DBExists(ctx context.Context, dbname string, options driver.Options) (bool, error) {
	return c.Client.DBExists(ctx, dbname, options)
}

func (c *client) DestroyDB(ctx context.Context, dbname string, options driver.Options) error {
	return c.Client.DestroyDB(ctx, dbname, options)
}

func (c *client) Version(ctx context.Context) (*driver.Version, error) {
	ver, err := c.Client.Version(ctx)
	if err != nil {
		return nil, err
	}
	return &driver.Version{
		Version:     ver.Version,
		Vendor:      ver.Vendor,
		RawResponse: ver.RawResponse,
	}, nil
}

func (c *client) DB(name string, options driver.Options) (driver.DB, error) {
	d := c.Client.DB(name, options)
	return &db{d}, nil
}

type db struct {
	*kivik.DB
}

var _ driver.DB = &db{}

func (d *db) AllDocs(ctx context.Context, opts driver.Options) (driver.Rows, error) {
	kivikRows := d.DB.AllDocs(ctx, opts)
	return &rows{kivikRows}, kivikRows.Err()
}

func (d *db) Query(ctx context.Context, ddoc, view string, opts driver.Options) (driver.Rows, error) {
	kivikRows := d.DB.Query(ctx, ddoc, view, opts)
	return &rows{kivikRows}, kivikRows.Err()
}

type atts struct {
	*kivik.AttachmentsIterator
}

var _ driver.Attachments = &atts{}

func (a *atts) Close() error { return nil }
func (a *atts) Next(att *driver.Attachment) error {
	next, err := a.AttachmentsIterator.Next()
	if err != nil {
		return err
	}
	*att = driver.Attachment(*next)
	return nil
}

func (d *db) Get(ctx context.Context, id string, opts driver.Options) (*driver.Document, error) {
	row := d.DB.Get(ctx, id, opts)
	rev, err := row.Rev()
	if err != nil {
		return nil, err
	}
	var doc json.RawMessage
	if err := row.ScanDoc(&doc); err != nil {
		return nil, err
	}
	attIter, err := row.Attachments()
	if err != nil && kivik.HTTPStatus(err) != http.StatusNotFound {
		return nil, err
	}

	var attachments *atts
	if attIter != nil {
		attachments = &atts{attIter}
	}

	return &driver.Document{
		Rev:         rev,
		Body:        io.NopCloser(bytes.NewReader(doc)),
		Attachments: attachments,
	}, nil
}

func (d *db) Stats(ctx context.Context) (*driver.DBStats, error) {
	i, err := d.DB.Stats(ctx)
	if err != nil {
		return nil, err
	}
	var cluster *driver.ClusterStats
	if i.Cluster != nil {
		c := driver.ClusterStats(*i.Cluster)
		cluster = &c
	}
	return &driver.DBStats{
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
	}, nil
}

func (d *db) Security(ctx context.Context) (*driver.Security, error) {
	s, err := d.DB.Security(ctx)
	if err != nil {
		return nil, err
	}
	sec := driver.Security{
		Admins:  driver.Members(s.Admins),
		Members: driver.Members(s.Members),
	}
	return &sec, err
}

func (d *db) SetSecurity(ctx context.Context, security *driver.Security) error {
	sec := &kivik.Security{
		Admins:  kivik.Members(security.Admins),
		Members: kivik.Members(security.Members),
	}
	return d.DB.SetSecurity(ctx, sec)
}

func (d *db) Changes(context.Context, driver.Options) (driver.Changes, error) {
	return nil, notYetImplemented
}

func (d *db) BulkDocs(_ context.Context, _ []interface{}) ([]driver.BulkResult, error) {
	// FIXME: Unimplemented
	return nil, notYetImplemented
}

func (d *db) PutAttachment(_ context.Context, _ string, _ *driver.Attachment, _ driver.Options) (string, error) {
	panic("PutAttachment should never be called")
}

func (d *db) GetAttachment(_ context.Context, _, _ string, _ driver.Options) (*driver.Attachment, error) {
	panic("GetAttachment should never be called")
}

func (d *db) GetAttachmentMeta(_ context.Context, _, _, _ string, _ driver.Options) (*driver.Attachment, error) {
	// FIXME: Unimplemented
	return nil, notYetImplemented
}

func (d *db) CreateDoc(_ context.Context, _ interface{}, _ driver.Options) (string, string, error) {
	panic("CreateDoc should never be called")
}

func (d *db) Delete(_ context.Context, _ string, _ driver.Options) (string, error) {
	panic("Delete should never be called")
}

func (d *db) DeleteAttachment(_ context.Context, _, _ string, _ driver.Options) (string, error) {
	panic("DeleteAttachment should never be called")
}

func (d *db) Put(_ context.Context, _ string, _ interface{}, _ driver.Options) (string, error) {
	panic("Put should never be called")
}
