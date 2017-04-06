package memory

import (
	"context"
	"io"

	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/errors"
)

// database is an in-memory database representation.
type db struct {
	*client
	dbName string
}

type indexDoc struct {
	ID    string        `json:"id"`
	Key   string        `json:"key"`
	Value indexDocValue `json:"value"`
}

type indexDocValue struct {
	Rev string `json:"rev"`
}

func (d *db) SetOption(_ string, _ interface{}) error {
	return errors.New("no options supported")
}

func (d *db) AllDocsContext(ctx context.Context, opts map[string]interface{}) (driver.Rows, error) {
	// FIXME: Unimplemented
	return nil, nil
}

func (d *db) GetContext(_ context.Context, docID string, doc interface{}, opts map[string]interface{}) error {
	// FIXME: Unimplemented
	return nil
}

func (d *db) CreateDocContext(_ context.Context, doc interface{}) (docID, rev string, err error) {
	// FIXME: Unimplemented
	return "", "", nil
}

func (d *db) PutContext(_ context.Context, docID string, doc interface{}) (rev string, err error) {
	// FIXME: Unimplemented
	return "", nil
}

func (d *db) DeleteContext(_ context.Context, docID, rev string) (newRev string, err error) {
	// FIXME: Unimplemented
	return "", nil
}

func (d *db) InfoContext(_ context.Context) (*driver.DBInfo, error) {
	// FIXME: Unimplemented
	return nil, nil
}

func (c *client) CompactContext(_ context.Context) error {
	// FIXME: Unimplemented
	return nil
}

func (d *db) CompactViewContext(_ context.Context, _ string) error {
	// FIXME: Unimplemented
	return nil
}

func (d *db) ViewCleanupContext(_ context.Context) error {
	// FIXME: Unimplemented
	return nil
}

func (d *db) SecurityContext(_ context.Context) (*driver.Security, error) {
	// FIXME: Unimplemented
	return nil, nil
}

func (d *db) SetSecurityContext(_ context.Context, _ *driver.Security) error {
	// FIXME: Unimplemented
	return nil
}

func (d *db) RevsLimitContext(_ context.Context) (limit int, err error) {
	// FIXME: Unimplemented
	return 0, nil
}

func (d *db) SetRevsLimitContext(_ context.Context, limit int) error {
	// FIXME: Unimplemented
	return nil
}

func (d *db) ChangesContext(ctx context.Context, opts map[string]interface{}) (driver.Rows, error) {
	// FIXME: Unimplemented
	return nil, nil
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
	return "", driver.Checksum{}, nil, nil
}
