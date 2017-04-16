package fs

import (
	"context"
	"io"

	"github.com/flimzy/kivik/driver"
)

type db struct {
	*client
	dbName string
}

func (d *db) AllDocs(_ context.Context, _ map[string]interface{}) (driver.Rows, error) {
	// FIXME: Unimplemented
	return nil, nil
}

func (d *db) Query(ctx context.Context, ddoc, view string, opts map[string]interface{}) (driver.Rows, error) {
	// FIXME: Unimplemented
	return nil, nil
}

func (d *db) Get(_ context.Context, docID string, doc interface{}, opts map[string]interface{}) error {
	// FIXME: Unimplemented
	return nil
}

func (d *db) CreateDoc(_ context.Context, doc interface{}) (docID, rev string, err error) {
	// FIXME: Unimplemented
	return "", "", nil
}

func (d *db) Put(_ context.Context, docID string, doc interface{}) (rev string, err error) {
	// FIXME: Unimplemented
	return "", nil
}

func (d *db) Delete(_ context.Context, docID, rev string) (newRev string, err error) {
	// FIXME: Unimplemented
	return "", nil
}

func (d *db) Info(_ context.Context) (*driver.DBInfo, error) {
	// FIXME: Unimplemented
	return nil, nil
}

func (d *db) Compact(_ context.Context) error {
	// FIXME: Unimplemented
	return nil
}

func (d *db) CompactView(_ context.Context, _ string) error {
	// FIXME: Unimplemented
	return nil
}

func (d *db) ViewCleanup(_ context.Context) error {
	// FIXME: Unimplemented
	return nil
}

func (d *db) Security(_ context.Context) (*driver.Security, error) {
	// FIXME: Unimplemented
	return nil, nil
}

func (d *db) SetSecurity(_ context.Context, _ *driver.Security) error {
	// FIXME: Unimplemented
	return nil
}

func (d *db) Changes(_ context.Context, _ map[string]interface{}) (driver.Rows, error) {
	// FIXME: Unimplemented
	return nil, nil
}

func (d *db) BulkDocs(_ context.Context, _ ...interface{}) (driver.BulkResults, error) {
	// FIXME: Unimplemented
	return nil, nil
}

func (d *db) PutAttachment(_ context.Context, _, _, _, _ string, _ io.Reader) (string, error) {
	// FIXME: Unimplemented
	return "", nil
}

func (d *db) GetAttachment(ctx context.Context, docID, rev, filename string) (contentType string, md5sum driver.Checksum, body io.ReadCloser, err error) {
	// FIXME: Unimplemented
	return "", driver.Checksum{}, nil, nil
}

func (d *db) DeleteAttachment(ctx context.Context, docID, rev, filename string) (newRev string, err error) {
	// FIXME: Unimplemented
	return "", nil
}
